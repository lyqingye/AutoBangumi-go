package mikan

import (
	bangumitypes "autobangumi-go/bangumi"
	"autobangumi-go/mdb"
	"autobangumi-go/utils"
	"bytes"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antlabs/strsim"
	tmdb "github.com/cyruzin/golang-tmdb"
)

func (parser *MikanRSSParser) Search(keyword string) (*bangumitypes.Bangumi, error) {
	resp, err := parser.http.R().SetQueryParam("searchstr", keyword).Get(parser.mikanEndpoint.JoinPath("RSS/Search").String())
	if err != nil {
		return nil, err
	}
	rssContent := MikanRss{}
	err = xml.Unmarshal(resp.Body(), &rssContent)
	if err != nil {
		return nil, err
	}
	result, err := parser.parseMikanRSS(&rssContent)
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
	return result[matchResult.BestIndex], nil
}

func (parser *MikanRSSParser) Search2(keyword string) (*bangumitypes.Bangumi, error) {
	resp, err := parser.http.R().SetQueryParam("searchstr", keyword).Get(parser.mikanEndpoint.JoinPath("HOME/Search").String())
	if err != nil {
		return nil, err
	}
	rssContent := MikanRss{}
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body()))
	if err != nil {
		return nil, err
	}
	doc.Find("#sk-container > div.central-container > table > tbody > tr.js-search-results-row").
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
	result, err := parser.parseMikanRSS(&rssContent)
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
	return result[matchResult.BestIndex], nil
}

func (parser *MikanRSSParser) searchBangumiTV(keyword string) (*mdb.Subjects, error) {
	cachedSubject := mdb.Subjects{}
	key := getBangumiTVCacheKeyByKeyword(keyword)
	cached, err := parser.db.Get(key, &cachedSubject)
	if err != nil || !cached {
		subject, err := parser.bangumiTvClient.SearchAnime2(keyword)
		if err != nil {
			return nil, err
		}
		return subject, parser.db.Set(key, subject)
	} else {
		return &cachedSubject, nil
	}
}

func (parser *MikanRSSParser) searchTMDB(keyword string) (*tmdb.TVDetails, error) {
	cachedTV := tmdb.TVDetails{}
	key := getTMDBCacheByKeyword(keyword)
	cached, err := parser.db.Get(key, &cachedTV)
	if err != nil || !cached {
		keyword = normalizationSearchTitle(keyword)
		result, err := parser.tmdb.SearchTVShowByKeyword(keyword)
		if err != nil {
			return nil, err
		}
		return result, parser.db.Set(key, result)
	} else {
		return &cachedTV, nil
	}
}

func (parser *MikanRSSParser) getTMDB(tmdbID int64) (*tmdb.TVDetails, error) {
	return parser.tmdb.GetTVDetailById(tmdbID)
}

func normalizationSearchTitle(keyword string) string {
	patterns := []string{
		"第([[:digit:]]+|\\p{Han}+)季",
		"第([[:digit:]]+|\\p{Han}+)期",
		"Season\\s*\\d+",
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		keyword = strings.ReplaceAll(keyword, re.FindString(keyword), "")
	}
	return strings.Split(keyword, " ")[0]
}
