package rss

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"net/url"
	bangumitypes "pikpak-bot/bangumi"
	"pikpak-bot/bus"
	"pikpak-bot/db"
	"pikpak-bot/mdb"
	"pikpak-bot/utils"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	torrent "github.com/anacrolix/torrent/metainfo"

	tmdb "github.com/cyruzin/golang-tmdb"

	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/go-resty/resty/v2"
	"github.com/nssteinbrenner/anitogo"
	"github.com/rs/zerolog"
)

var (
	TMDBJPLangOptions = map[string]string{
		"language": "ja-JP",
	}
	TMDBZHLangOptions = map[string]string{
		"language": "zh-CN",
	}
	TMDBENLangOptions = map[string]string{
		"language": "en-US",
	}
)

type MikanRssItem struct {
	Guid struct {
		IsPermaLink string `xml:"isPermaLink,attr"`
	} `xml:"guid"`
	Link        string `xml:"link"`
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Torrent     struct {
		Xmlns         string `xml:"xmlns,attr"`
		Link          string `xml:"link"`
		ContentLength string `xml:"contentLength"`
		PubDate       string `xml:"pubDate"`
	} `xml:"torrent"`
	Enclosure struct {
		Type   string `xml:"type,attr"`
		Length string `xml:"length,attr"`
		URL    string `xml:"url,attr"`
	} `xml:"enclosure"`
}

type MikanRss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Title       string         `xml:"title"`
		Link        string         `xml:"link"`
		Description string         `xml:"description"`
		Item        []MikanRssItem `xml:"item"`
	} `xml:"channel"`
}

var (
	KeyParseCacheByLink = []byte("parse-cache-link")
)

// ParseCache
// Indexes:
// - Mikan ItemLink -> Cache
type ParseCache struct {
	TMDBId    int64
	SubjectId int64
	Season    uint
	Episode   bangumitypes.Episode
}

func getParseCacheKeyByLink(link string) []byte {
	return append(KeyParseCacheByLink, []byte(link)...)
}

var (
	KeyBangumiTVCacheBySubjectId    = []byte("bangumiTV-cache-subject-id")
	KeyBangumiTVCacheByKeyword      = []byte("bangumiTV-cache-keyword")
	KeyTMDBCacheByKeyword           = []byte("TMDB-cache-keyword")
	KeyBlackItemLink                = []byte("mikan-black-item")
	KeyMikanBangumiToBangumiTvCache = []byte("mikan-bangumi-to-bangumi-tv-cache")
)

func getBangumiTVCacheKeyBySubjectId(subject int64) []byte {
	return append(KeyBangumiTVCacheBySubjectId, []byte(strconv.FormatInt(subject, 10))...)
}

func getBangumiTVCacheKeyByKeyword(keyword string) []byte {
	return append(KeyBangumiTVCacheByKeyword, []byte(keyword)...)
}
func getTMDBCacheByKeyword(keyword string) []byte {
	return append(KeyTMDBCacheByKeyword, []byte(keyword)...)
}

func getBlackItemLinkKey(itemLink string) []byte {
	return append(KeyBlackItemLink, []byte(itemLink)...)
}

func getMikanBangumiToBangumiTVCache(mikanBangumiId string) []byte {
	return append(KeyMikanBangumiToBangumiTvCache, []byte(mikanBangumiId)...)
}

type MikanRSSParser struct {
	mikanEndpoint   *url.URL
	rssLink         string
	http            *resty.Client
	eb              *bus.EventBus
	db              *db.DB
	logger          zerolog.Logger
	tmdb            *tmdb.Client
	bangumiTvClient *mdb.BangumiTVClient
}

func NewMikanRSSParser(rss string, eb *bus.EventBus, parseCacheDB *db.DB, tmdbClient *tmdb.Client, bangumiTVClient *mdb.BangumiTVClient) (*MikanRSSParser, error) {
	uri, err := url.Parse(rss)
	if err != nil {
		return nil, err
	}
	endpoint, err := url.Parse(uri.Hostname())
	if err != nil {
		return nil, err
	}
	endpoint.Scheme = uri.Scheme
	parser := MikanRSSParser{
		logger:          utils.GetLogger("mikanRSS"),
		mikanEndpoint:   endpoint,
		rssLink:         rss,
		http:            resty.New(),
		eb:              eb,
		db:              parseCacheDB,
		tmdb:            tmdbClient,
		bangumiTvClient: bangumiTVClient,
	}
	return &parser, nil
}

func (parser *MikanRSSParser) Parse() (*RSSInfo, error) {
	var err error
	rssInfo := RSSInfo{}
	mikan, err := parser.getRss()
	if err != nil {
		return nil, err
	}
	bangumiMap := make(map[int64]*bangumitypes.Bangumi)
	for i, item := range mikan.Channel.Item {
		if item.Link != "" {
			if parser.isBlackItemLink(item.Link) {
				continue
			}
			parser.logger.Debug().Str("title", item.Title).Msg(fmt.Sprintf("parse Episode %d/%d", i+1, len(mikan.Channel.Item)))
			err := parser.parserItemLink(item, bangumiMap)
			if err != nil {
				parser.blackItemLink(item.Link)
			}
		}
	}
	for _, bangumi := range bangumiMap {
		rssInfo.Bangumis = append(rssInfo.Bangumis, *bangumi)
	}
	filterBangumi(&rssInfo)
	return &rssInfo, nil
}

func (parser *MikanRSSParser) parserItemLink(item MikanRssItem, cacheBangumi map[int64]*bangumitypes.Bangumi) error {
	var err error
	item.Title = strings.ReplaceAll(item.Title, "【", "[")
	item.Title = strings.ReplaceAll(item.Title, "】", "]")

	// skip collection
	re := regexp.MustCompile(`\d{1,3}-\d{1,3}`)
	if re.MatchString(item.Title) {
		parser.logger.Warn().Str("link", item.Link).Str("title", item.Title).Msg("ignore collection")
		return nil
	}

	var episode bangumitypes.Episode
	cache, found := parser.getParseCache(item.Link)
	if !found {
		episode, err = parser.ParseEpisode(item.Link)
		if err != nil {
			if parser.eb != nil {
				parser.eb.Publish(bus.RSSTopic, bus.Event{
					EventType: bus.RSSParseErrEventType,
					Inner:     fmt.Errorf("faild to parse episode, link: %s, title: %s", item.Link, item.Title),
				})
			}
			parser.logger.Warn().Err(err).Str("link", item.Link).Str("title", item.Title).Msg("parse episode error")
			return err
		}
		episode.EpisodeTitle = item.Title
		episodeTitle := item.Title

		// using anitogo parse filename
		parsedElements := anitogo.Parse(episodeTitle, anitogo.DefaultOptions)
		if len(parsedElements.EpisodeNumber) > 0 {
			epStr := parsedElements.EpisodeNumber[0]
			epNumber, err := strconv.ParseUint(epStr, 10, 32)
			if err == nil && epNumber > 0 {
				episode.EPNumber = uint(epNumber)
			}
		}
		if len(parsedElements.AnimeSeason) > 0 {
			seasonStr := parsedElements.AnimeSeason[0]
			season, err := strconv.ParseUint(seasonStr, 10, 32)
			if err == nil && season > 0 {
				episode.Season = uint(season)
			}
		}
		if len(parsedElements.Language) > 0 {
			for _, l := range parsedElements.Language {
				episode.Lang = append(episode.Lang, normalizationLang(l))
			}
		}
		if len(parsedElements.Subtitles) > 0 {
			for _, l := range parsedElements.Subtitles {
				episode.Lang = append(episode.Lang, normalizationLang(l))
			}
		}
		if len(episode.Lang) == 0 {
			episode.Lang = []string{bangumitypes.SubtitleUnknown}
		}
		episode.Resolution = normalizationResolution(parsedElements.VideoResolution)
		fileSize, err := humanize.ParseBytes(item.Torrent.ContentLength)
		if err == nil {
			episode.FileSize = fileSize
		}
		episode.Date = item.Torrent.PubDate
		cache.Episode = episode
	} else {
		episode = cache.Episode
	}

	var bangumi = &bangumitypes.Bangumi{
		Title:     episode.BangumiTitle,
		Season:    cache.Season,
		TmDBId:    cache.TMDBId,
		SubjectId: cache.SubjectId,
	}

	// parse bangumi information using tmdb
	if episode.SubjectId == 0 || episode.Season == 0 {
		var subject *mdb.Subjects
		if episode.SubjectId != 0 {
			subject, err = parser.getBangumiTVSubject(episode.SubjectId)
			if err != nil {
				return err
			}
		} else {
			subject, err = parser.searchBangumiTV(episode.BangumiTitle)
			if err != nil {
				return err
			}
			episode.SubjectId = subject.ID
			cache.SubjectId = subject.ID

			// cache mikan bangumiId -> BangumiTV Id
			_ = parser.db.Set(getMikanBangumiToBangumiTVCache(episode.MikanBangumiId), &episode.SubjectId)
		}
		bangumi.EPCount = uint(subject.Eps)
		var searchTitles []string
		searchTitles = append(searchTitles, episode.BangumiTitle, subject.NameCn, subject.Name)
		searchTitles = append(searchTitles, subject.GetAliasNames()...)
		subjectAirDate, err := utils.ParseDate(subject.Date)
		if err != nil {
			return err
		}

		// try search tmdb
		for _, searchTitle := range searchTitles {
			if searchTitle == "" {
				continue
			}
			tvDetails, err := parser.searchTMDB(searchTitle)
			if err == nil {
				bangumi.TmDBId = tvDetails.ID
				bangumi.Title = tvDetails.Name
				episode.BangumiTitle = bangumi.Title
				// predict season number using air date
				minDiff := time.Duration(math.MaxInt64)
				closeIndex := -1
				for i, season := range tvDetails.Seasons {
					// skip special season
					// FIXME:
					if season.SeasonNumber == 0 {
						continue
					}
					seasonAriDate, err := utils.ParseDate(season.AirDate)
					if err != nil {
						return err
					}
					diff := subjectAirDate.Sub(seasonAriDate).Abs()
					if diff <= minDiff {
						minDiff = diff
						closeIndex = i
					}
				}
				if closeIndex != -1 {
					bangumi.Season = uint(tvDetails.Seasons[closeIndex].SeasonNumber)
				}
				if bangumi.Season != 0 {
					break
				}
			} else {
				parser.logger.Error().Err(err).Msg("search tmdb error")
			}
		}
	}

	if bangumi.Season == 0 {
		err = fmt.Errorf("unknown season, mikan link: %s", item.Link)
		parser.logger.Err(err).Msg("parse error")
		return err
	}

	if existBangumi, found := cacheBangumi[episode.SubjectId]; found {
		bangumi = existBangumi
	}

	episode.Season = bangumi.Season
	if err := episode.Validate(); err != nil {
		parser.eb.Publish(bus.RSSTopic, bus.Event{
			EventType: bus.RSSParseErrEventType,
			Inner:     fmt.Errorf("faild to parse episode, link: %s, title: %s", item.Link, item.Title),
		})
		parser.logger.Warn().Err(err).Str("link", item.Link).Str("title", item.Title).Msg("parse episode error")
		return err
	}
	bangumi.SubjectId = episode.SubjectId
	bangumi.Episodes = append(bangumi.Episodes, episode)
	cacheBangumi[episode.SubjectId] = bangumi

	// save parse cache
	cache.Episode = episode
	cache.TMDBId = bangumi.TmDBId
	cache.Season = episode.Season
	cache.SubjectId = episode.SubjectId
	parser.storeParseCache(item.Link, cache)
	return nil
}

func (parser *MikanRSSParser) getRss() (*MikanRss, error) {
	resp, err := parser.http.R().EnableTrace().Get(parser.rssLink)
	if err != nil {
		return nil, err
	}
	rssContent := MikanRss{}
	err = xml.Unmarshal(resp.Body(), &rssContent)
	if err != nil {
		return nil, err
	}
	return &rssContent, nil
}

func (parser *MikanRSSParser) ParseEpisode(link string) (bangumitypes.Episode, error) {
	resp, err := parser.http.R().EnableTrace().Get(link)
	ep := bangumitypes.Episode{}
	if err != nil {
		return ep, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body()))
	if err != nil {
		return ep, err
	}
	titleSelector := doc.Find("#sk-container > div.pull-left.leftbar-container > p.bangumi-title > a.w-other-c")
	subGroupSelector := doc.Find("#sk-container > div.pull-left.leftbar-container > p.bangumi-info > a.magnet-link-wrap")
	buttonSelector := doc.Find("#sk-container > div.pull-left.leftbar-container > div.leftbar-nav > a.episode-btn")
	ep.BangumiTitle = titleSelector.Text()

	var mikanBangumiLink string
	for _, node := range titleSelector.Nodes {
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				if strings.HasPrefix(attr.Val, "/Home/Bangumi/") {
					mikanBangumiLink = attr.Val
				}
			}
		}
	}
	ep.Subgroup = subGroupSelector.Text()
	for _, node := range buttonSelector.Nodes {
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				if strings.HasSuffix(attr.Val, ".torrent") {
					torrentDownloadUrl := parser.mikanEndpoint.JoinPath(attr.Val)
					resp, err := parser.http.R().Get(torrentDownloadUrl.String())
					if err != nil {
						return ep, err
					}
					ep.Torrent = resp.Body()
				}
				if strings.HasPrefix(attr.Val, "magnet:?xt") {
					ep.Magnet = attr.Val
				}
			}
		}
	}

	if mikanBangumiLink != "" {
		if idx := strings.Index(mikanBangumiLink, "#"); idx != -1 {
			mikanBangumiLink = mikanBangumiLink[:idx]
		}
		ep.MikanBangumiId = strings.ReplaceAll(mikanBangumiLink, "/Home/Bangumi/", "")
		key := getMikanBangumiToBangumiTVCache(ep.MikanBangumiId)
		var cachedSubjectId int64
		cached, err := parser.db.Get(key, &cachedSubjectId)
		if err == nil && cached {
			ep.SubjectId = cachedSubjectId
		} else {
			resp, err = parser.http.R().EnableTrace().Get(parser.mikanEndpoint.JoinPath(mikanBangumiLink).String())
			if err != nil {
				return ep, err
			}
			doc, err = goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body()))
			if err != nil {
				return ep, err
			}

			var bangumiTVLink = ""
			bangumiTVLinkSelector := doc.Find("#sk-container > div.pull-left.leftbar-container > p.bangumi-info > a.w-other-c")
			for _, node := range bangumiTVLinkSelector.Nodes {
				for _, attr := range node.Attr {
					if attr.Key == "href" {
						if strings.HasPrefix(attr.Val, "https://bgm.tv") {
							bangumiTVLink = attr.Val
						}
					}
				}
			}
			if bangumiTVLink != "" {
				subjectIdStr := strings.ReplaceAll(bangumiTVLink, "https://bgm.tv/subject/", "")
				subjectId, err := strconv.ParseInt(subjectIdStr, 10, 64)
				if err != nil {
					return ep, err
				}
				ep.SubjectId = subjectId
				_ = parser.db.Set(key, &subjectId)
			}
		}
	}

	torr, err := torrent.Load(bytes.NewBuffer(ep.Torrent))
	if err != nil {
		return ep, err
	}
	ep.TorrentHash = torr.HashInfoBytes().HexString()

	return ep, nil
}

func (parser *MikanRSSParser) getParseCache(itemLink string) (*ParseCache, bool) {
	cache := ParseCache{}
	found, err := parser.db.Get(getParseCacheKeyByLink(itemLink), &cache)
	if err != nil {
		return nil, false
	}
	return &cache, found
}

func (parser *MikanRSSParser) storeParseCache(itemLink string, cache *ParseCache) {
	err := parser.db.Set(getParseCacheKeyByLink(itemLink), cache)
	if err != nil {
		parser.logger.Err(err).Msg("store parse cache error")
	}
}

func (parser *MikanRSSParser) blackItemLink(itemLink string) {
	err := parser.db.Set(getBlackItemLinkKey(itemLink), nil)
	if err != nil {
		parser.logger.Err(err).Msg("black item link error")
	}
}

func (parser *MikanRSSParser) isBlackItemLink(itemLink string) bool {
	found, err := parser.db.Has(getBlackItemLinkKey(itemLink))
	return found && err == nil
}

func (parser *MikanRSSParser) getBangumiTVSubject(subjectId int64) (*mdb.Subjects, error) {
	cachedSubject := mdb.Subjects{}
	key := getBangumiTVCacheKeyBySubjectId(subjectId)
	cached, err := parser.db.Get(key, &cachedSubject)
	if err != nil || !cached {
		subject, err := parser.bangumiTvClient.GetSubjects(subjectId)
		if err != nil {
			return nil, err
		}
		return subject, parser.db.Set(key, subject)
	} else {
		return &cachedSubject, nil
	}
}

func (parser *MikanRSSParser) searchBangumiTV(keyword string) (*mdb.Subjects, error) {
	cachedSubject := mdb.Subjects{}
	key := getBangumiTVCacheKeyByKeyword(keyword)
	cached, err := parser.db.Get(key, &cachedSubject)
	if err != nil || !cached {
		subject, err := parser.bangumiTvClient.SearchAnime(keyword)
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
		for _, opts := range []map[string]string{TMDBZHLangOptions, TMDBJPLangOptions, TMDBENLangOptions} {
			searchResult, err := parser.tmdb.GetSearchTVShow(keyword, opts)
			if err != nil {
				return nil, err
			}
			if len(searchResult.Results) > 0 {
				tvDetails, err := parser.tmdb.GetTVDetails(int(searchResult.Results[0].ID), opts)
				if err == nil {
					return tvDetails, parser.db.Set(key, tvDetails)
				}
			}
		}
		return nil, errors.New("tmdb search result empty")
	} else {
		return &cachedTV, nil
	}
}

func filterBangumi(rssInfo *RSSInfo) {
	for i, b := range rssInfo.Bangumis {
		epMap := make(map[string][]bangumitypes.Episode)
		for _, ep := range b.Episodes {
			key := fmt.Sprintf("%d-%d", ep.Season, ep.EPNumber)
			dupEps := epMap[key]
			dupEps = append(dupEps, ep)
			epMap[key] = dupEps
		}
		var processedEps []bangumitypes.Episode
		for _, eps := range epMap {
			if len(eps) == 1 {
				processedEps = append(processedEps, eps[0])
			} else {
				// priority: resolution
				sort.Slice(eps, func(i, j int) bool {
					epI := eps[i].Resolution
					epJ := eps[j].Resolution
					langI := eps[i].Lang[0]
					langJ := eps[j].Lang[0]
					iDate, err1 := utils.SmartParseDate(eps[i].Date)
					jDate, err2 := utils.SmartParseDate(eps[j].Date)
					result := ResolutionPriority[epI] > ResolutionPriority[epJ] || SubtitlePriority[langI] > SubtitlePriority[langJ]
					if err1 == nil && err2 == nil {
						result = result || iDate.Unix() > jDate.Unix()
					} else {
						println("")
					}
					return result
				})
				processedEps = append(processedEps, eps[0])
			}
		}
		sort.Slice(processedEps, func(i, j int) bool {
			return processedEps[i].EPNumber < processedEps[j].EPNumber
		})
		rssInfo.Bangumis[i].Episodes = processedEps
	}
}

func normalizationResolution(resolution string) string {
	switch resolution {
	case "1080P", "1080p", "1920x1080", "1920X1080":
		return bangumitypes.Resolution1080p
	case "720P", "720p", "1024x720", "1024X720":
		return bangumitypes.Resolution720p
	default:
		return bangumitypes.ResolutionUnknown
	}
}

func normalizationLang(lang string) string {
	for k, v := range bangumitypes.SubTitleLangKeyword {
		if strings.Contains(lang, k) {
			return v
		}
	}
	return bangumitypes.SubtitleUnknown
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
	return keyword
}
