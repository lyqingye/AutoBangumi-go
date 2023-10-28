package mikan

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"autobangumi-go/mdb"
	"autobangumi-go/utils"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/pkg/errors"

	"github.com/PuerkitoBio/goquery"
	"github.com/antlabs/strsim"
)

func (parser *MikanRSSParser) Search(title string, tmdbID int64) (*Bangumi, error) {
	var tv *tmdb.TVDetails
	var err error

	if tmdbID == 0 {
		tv, err = parser.searchTMDB(title)
		if err != nil {
			return nil, err
		}
	} else {
		tv, err = parser.getTMDB(tmdbID)
		if err != nil {
			return nil, err
		}
	}

	searchResult, err := parser.Search2(title)
	if err != nil {
		subject, err := parser.searchBangumiTV(title)
		if err != nil {
			return nil, err
		}
		searchResult, err = parser.Search2(subject.NameCn)
		if err != nil {
			return nil, err
		}
	}
	if searchResult.Info.TmDBId == tv.ID {
		return searchResult, nil
	}
	return nil, errors.New("mikan complete error, bangumi not found")
}

func (parser *MikanRSSParser) Search2(keyword string) (*Bangumi, error) {
	resp, err := parser.http.R().SetQueryParam("searchstr", keyword).Get(parser.mikanEndpoint.JoinPath("HOME/Search").String())
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body()))
	if err != nil {
		return nil, err
	}
	rssContent, err := parser.parserRSSFromWebPage(doc)
	if err != nil {
		return nil, err
	}
	result, err := parser.parseMikanRSS(rssContent)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("search bangumi empty: %s", keyword)
	}
	var names []string
	for _, bgm := range result {
		names = append(names, bgm.Info.Title)
	}
	matchResult := strsim.FindBestMatch(keyword, names)
	rs := result[matchResult.BestIndex]
	if rs.GetMikanID() != "" {
		return parser.Search3(rs.GetMikanID())
	}
	return rs, nil
}

func (parser *MikanRSSParser) Search3(bangumiId string) (*Bangumi, error) {
	resp, err := parser.http.R().Get(parser.mikanEndpoint.JoinPath(fmt.Sprintf("HOME/Bangumi/%s", bangumiId)).String())
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body()))
	if err != nil {
		return nil, err
	}
	rssContent, err := parser.parserRSSFromWebPage(doc)
	if err != nil {
		return nil, err
	}
	result, err := parser.parseMikanRSS(rssContent)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("search bangumi empty: %s", bangumiId)
	}
	for _, bgm := range result {
		for _, season := range bgm.Seasons {
			if season.MikanBangumiId == bangumiId {
				return bgm, nil
			}
		}
	}
	return nil, fmt.Errorf("search bangumi empty: %s", bangumiId)
}

func (parser *MikanRSSParser) parserRSSFromWebPage(doc *goquery.Document) (*MikanRss, error) {
	rssContent := MikanRss{}
	doc.Find("table > tbody > tr").
		Each(func(_ int, tr *goquery.Selection) {
			item := MikanRssItem{}
			tr.Find("td > a").Each(func(_ int, a *goquery.Selection) {
				if val, found := a.Attr("class"); found && val == "magnet-link-wrap" {
					item.Title = a.Text()
				}
				if val, found := a.Attr("href"); found {
					if strings.HasPrefix(val, "/Home/Episode") {
						item.Link = parser.mikanEndpoint.JoinPath(val).String()
						item.Torrent.Link = item.Link
					}
				}
			})
			tr.Find("td").Each(func(i int, td *goquery.Selection) {
				_, err := utils.SmartParseDate(td.Text())
				if err == nil {
					item.Torrent.PubDate = td.Text()
				}
			})
			if item.Torrent.PubDate != "" {
				rssContent.Channel.Item = append(rssContent.Channel.Item, item)
			}
		})
	return &rssContent, nil
}

func (parser *MikanRSSParser) searchTMDB(keyword string) (*tmdb.TVDetails, error) {
	keyword = normalizationSearchTitle(keyword)

	cache, err := parser.cm.GetTMDBCache(keyword)
	if err == nil {
		return &cache, nil
	}
	value, err := parser.tmdb.SearchTVShowByKeyword(keyword)
	if err != nil {
		return nil, err
	}
	_ = parser.cm.StoreTMDBCache(keyword, *value)
	return value, nil
}

func (parser *MikanRSSParser) getTMDB(tmdbID int64) (*tmdb.TVDetails, error) {
	cache, err := parser.cm.GetTMDBCacheByID(tmdbID)
	if err == nil {
		return &cache, nil
	}

	value, err := parser.tmdb.GetTVDetailById(tmdbID)
	if err != nil {
		return nil, err
	}

	_ = parser.cm.StoreTMDBCacheById(tmdbID, *value)
	return value, nil
}

func (parser *MikanRSSParser) searchBangumiTV(title string) (*mdb.Subjects, error) {
	cache, err := parser.cm.GetBangumiTVCache(title)
	if err == nil {
		return &cache, nil
	}

	value, err := parser.bangumiTV.SearchAnime2(title)
	if err != nil {
		return nil, err
	}
	_ = parser.cm.StoreBangumiTVCache(title, *value)

	return value, nil
}

func (parser *MikanRSSParser) getBangumiTVSubjects(id int64) (*mdb.Subjects, error) {
	cache, err := parser.cm.GetBangumiTVSubjectsCache(id)
	if err == nil {
		return &cache, nil
	}
	value, err := parser.bangumiTV.GetSubjects(id)
	if err != nil {
		return nil, err
	}

	_ = parser.cm.StoreBangumiTVSubjectsCache(id, *value)

	return value, nil
}

func normalizationSearchTitle(keyword string) string {
	patterns := []string{
		"第([[:digit:]]+|\\p{Han}+)季",
		"第([[:digit:]]+|\\p{Han}+)期",
		"SeasonNum\\s*\\d+",
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		keyword = strings.ReplaceAll(keyword, re.FindString(keyword), "")
	}
	return strings.Split(keyword, " ")[0]
}
