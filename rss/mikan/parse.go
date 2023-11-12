package mikan

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	bangumitypes "autobangumi-go/bangumi"
	"autobangumi-go/mdb"
	"autobangumi-go/utils"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"

	"github.com/PuerkitoBio/goquery"
	torrent "github.com/anacrolix/torrent/metainfo"
	"github.com/nssteinbrenner/anitogo"
)

func (parser *MikanRSSParser) parserItemLink(item MikanRssItem) (*ParseItemResult, error) {
	var err error
	ret := ParseItemResult{}
	item.Title = strings.ReplaceAll(item.Title, "【", "[")
	item.Title = strings.ReplaceAll(item.Title, "】", "]")
	// skip collection
	re := regexp.MustCompile(`\d{1,3}-\d{1,3}`)
	if re.MatchString(item.Title) {
		parser.logger.Warn().Str("link", item.Link).Str("title", item.Title).Msg("ignore collection")
		return nil, errors.New("collection not support")
	}

	cache, err := parser.cm.GetParseCache(item.Link)
	if err != nil {
		// Episode information from rss item
		ret.Resource.RawFilename = item.Title
		pubDate, err := utils.SmartParseDate(item.Torrent.PubDate)
		if err != nil {
			return nil, err
		}
		ret.Resource.TorrentPubDate = pubDate

		// Parse Episode from item description webpage, result information is reliable
		// - title (nullable)
		// - Mikan bangumi Id (nullable)
		// - BangumiTV subject Id (nullable)
		// - Torrent
		// - TorrentHash
		// - Magnet
		// - FileSize
		fromWebPage, err := parser.parseEpisodeByItem(item.Link)
		if err != nil {
			parser.logger.Warn().Err(err).Str("link", item.Link).Str("title", item.Title).Msg("parse episode error")
			return nil, errors.Wrap(err, "parse item")
		}

		ret.Resource.Magnet = fromWebPage.Resource.Magnet
		ret.Resource.Torrent = fromWebPage.Resource.Torrent
		ret.Resource.TorrentHash = fromWebPage.Resource.TorrentHash
		ret.Resource.FileSize = fromWebPage.Resource.FileSize
		ret.MikanBangumiId = fromWebPage.MikanBangumiId
		ret.SubjectId = fromWebPage.SubjectId

		// Parse Episode from filename, result information is unreliable
		// - BangumiTitle (nullable)
		// - Episode number (nullable)
		// - SeasonNum number (nullable)
		// - Lang (nullable)
		// - Resolution (nullable)
		// - ResourceType (nullable)
		// - Subgroup (nullable)

		fromFilename, err := parser.parseEpisodeByFilename(item.Title)
		if err != nil {
			return nil, err
		}

		// Skip collections
		if fromFilename.Resource.Type == bangumitypes.ResourceTypeCollection {
			parser.logger.Warn().Err(err).Str("title", item.Title).Msg("skip collection")
			return nil, errors.New("collection not support")
		}

		if fromFilename.EpNum == 0 {
			parser.logger.Warn().Str("filename", item.Title).Msg("parse episode number from filename err")
			return nil, errors.New("parse episode number from filename err")
		}

		ret.EpNum = fromFilename.EpNum
		ret.Resource.SubtitleLang = fromFilename.Resource.SubtitleLang
		ret.Resource.Resolution = fromFilename.Resource.Resolution
		ret.Resource.Subgroup = fromFilename.Resource.Subgroup
		ret.Resource.Type = fromFilename.Resource.Type

		if fromWebPage.Title == "" && fromFilename.Title == "" {
			return nil, errors.New("could not found title from link page and filename")
		}

		if fromWebPage.Title != "" {
			ret.Title = fromWebPage.Title
		} else if fromFilename.Title != "" {
			ret.Title = fromFilename.Title
		}
	} else {
		ret = cache
	}

	// Try to predict Episode season number using bangumi TV and tmdb
	// 1. get air date from bangumi tv
	// 2. get seasons from tmdb
	// 3. using air data to predict season number
	//
	if ret.SubjectId == 0 || ret.SeasonNum == 0 {
		// the subject id comes from parsing item link page
		// if the subgroup does not have a link associated with bangumi-tv when publishing resources
		// then we will try searching based on the title
		var subject *mdb.Subjects
		if ret.SubjectId != 0 {
			subject, err = parser.getBangumiTVSubjects(ret.SubjectId)
			if err != nil {
				return nil, err
			}
		} else {
			subject, err = parser.searchBangumiTV(ret.Title)
			if err != nil {
				return nil, err
			}
			ret.SubjectId = subject.ID

			// cache mikan bangumiId -> BangumiTV Id
			if ret.MikanBangumiId != "" {
				_ = parser.cm.StoreMikanBangumiToBangumiTV(ret.MikanBangumiId, ret.SubjectId)
			}
		}

		// now we get episode air date
		subjectAirDate, err := utils.SmartParseDate(subject.Date)
		if err != nil {
			return nil, errors.Wrap(err, "parse date")
		}

		var searchTitles []string
		searchTitles = append(searchTitles, ret.Title, subject.NameCn, subject.Name)
		searchTitles = append(searchTitles, subject.GetAliasNames()...)

		// try search seasons from tmdb
		for _, searchTitle := range searchTitles {
			if searchTitle == "" {
				continue
			}
			tvDetails, err := parser.searchTMDB(searchTitle)
			if err == nil {
				ret.TmDBId = tvDetails.ID
				// using tmdb title as bangumi title
				ret.Title = tvDetails.Name

				// predict season number using air date
				minDiff := time.Duration(math.MaxInt64)
				closeIndex := -1
				for i, season := range tvDetails.Seasons {
					// NOTE:
					// If season number is 0 , then the season maybe a Special or TV
					if season.SeasonNumber == 0 {
						continue
					}

					// the season is not air
					if season.AirDate == "" {
						continue
					}

					// using episode air date to predict episode season
					seasonAriDate, err := utils.SmartParseDate(season.AirDate)
					if err != nil {
						return nil, err
					}
					diff := subjectAirDate.Sub(seasonAriDate).Abs()
					if diff <= minDiff {
						minDiff = diff
						closeIndex = i
					}
				}

				if closeIndex != -1 {
					ret.SeasonNum = uint(tvDetails.Seasons[closeIndex].SeasonNumber)
					ret.EpCount = uint(tvDetails.Seasons[closeIndex].EpisodeCount)
				}

				if ret.SeasonNum != 0 {
					break
				}

			} else {
				parser.logger.Warn().Err(err).Msg("search tmdb error")
			}
		}
	}

	if ret.SeasonNum == 0 || ret.EpCount == 0 {
		return nil, errors.Errorf("unknown season, mikan link: %s", item.Link)
	}

	return &ret, parser.cm.StoreParseCache(item.Link, ret)
}

type ParseItemResult struct {
	Title          string
	TmDBId         int64
	SubjectId      int64
	MikanBangumiId string

	SeasonNum uint
	EpNum     uint
	EpCount   uint
	Resource  TorrentResource
}

func (parser *MikanRSSParser) parseEpisodeByItem(link string) (*ParseItemResult, error) {
	result := ParseItemResult{}

	resp, err := parser.http.R().Get(link)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body()))
	if err != nil {
		return nil, err
	}
	titleSelector := doc.Find("#sk-container > div.pull-left.leftbar-container > p.bangumi-title > a.w-other-c")
	subGroupSelector := doc.Find("#sk-container > div.pull-left.leftbar-container > p.bangumi-info > a.magnet-link-wrap")
	buttonSelector := doc.Find("#sk-container > div.pull-left.leftbar-container > div.leftbar-nav > a.episode-btn")
	result.Title = titleSelector.Text()

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

	result.Resource.Subgroup = subGroupSelector.Text()
	for _, node := range buttonSelector.Nodes {
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				if strings.HasSuffix(attr.Val, ".torrent") {
					torrentDownloadUrl := parser.mikanEndpoint.JoinPath(attr.Val)
					resp, err := parser.http.R().Get(torrentDownloadUrl.String())
					if err != nil {
						return nil, err
					}
					result.Resource.Torrent = resp.Body()
					torr, err := torrent.Load(bytes.NewBuffer(result.Resource.Torrent))
					if err != nil {
						return nil, err
					}
					info, err := torr.UnmarshalInfo()
					if err != nil {
						return nil, err
					}
					var torrentFilesSize int64
					for _, fi := range info.Files {
						torrentFilesSize += fi.Length
					}
					if torrentFilesSize > info.Length {
						result.Resource.FileSize = uint64(torrentFilesSize)
					} else {
						result.Resource.FileSize = uint64(info.Length)
					}
				}
				if strings.HasPrefix(attr.Val, "magnet:?xt") {
					result.Resource.Magnet = attr.Val
				}
			}
		}
	}

	if mikanBangumiLink != "" {
		if idx := strings.Index(mikanBangumiLink, "#"); idx != -1 {
			mikanBangumiLink = mikanBangumiLink[:idx]
		}
		result.MikanBangumiId = strings.ReplaceAll(mikanBangumiLink, "/Home/Bangumi/", "")
		var cachedSubjectId int64
		if result.MikanBangumiId != "" {
			cachedSubjectId, err = parser.cm.GetMikanBangumiToBangumiTV(result.MikanBangumiId)
		}
		if err == nil {
			result.SubjectId = cachedSubjectId
		} else {
			resp, err = parser.http.R().Get(parser.mikanEndpoint.JoinPath(mikanBangumiLink).String())
			if err != nil {
				return nil, err
			}
			doc, err = goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body()))
			if err != nil {
				return nil, err
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
					return nil, err
				}
				result.SubjectId = subjectId
				_ = parser.cm.StoreMikanBangumiToBangumiTV(result.MikanBangumiId, subjectId)
			}
		}
	}

	torr, err := torrent.Load(bytes.NewBuffer(result.Resource.Torrent))
	if err != nil {
		return nil, err
	}
	result.Resource.TorrentHash = torr.HashInfoBytes().HexString()
	return &result, nil
}

func (parser *MikanRSSParser) parseEpisodeByFilename(filename string) (*ParseItemResult, error) {
	ret := ParseItemResult{}

	parsedElements := anitogo.Parse(filename, anitogo.DefaultOptions)

	if len(parsedElements.EpisodeNumber) == 1 {
		epStr := parsedElements.EpisodeNumber[0]
		epNumber, err := strconv.ParseUint(epStr, 10, 32)
		if err == nil && epNumber > 0 {
			ret.EpNum = uint(epNumber)
		}
	} else if len(parsedElements.EpisodeNumber) == 2 {
		startEp, err1 := strconv.ParseUint(parsedElements.EpisodeNumber[0], 10, 32)
		endEp, err2 := strconv.ParseUint(parsedElements.EpisodeNumber[1], 10, 32)
		if err1 == nil && err2 == nil && startEp > 0 && endEp > startEp {
			ret.Resource.Type = bangumitypes.ResourceTypeCollection
		}
	}

	if len(parsedElements.AnimeSeason) > 0 {
		seasonStr := parsedElements.AnimeSeason[0]
		season, err := strconv.ParseUint(seasonStr, 10, 32)
		if err == nil && season > 0 {
			ret.SeasonNum = uint(season)
		}
	}

	ret.Resource.Subgroup = parsedElements.ReleaseGroup

	if len(parsedElements.Language) > 0 {
		for _, l := range parsedElements.Language {
			ret.Resource.SubtitleLang = append(ret.Resource.SubtitleLang, normalizationLang(l))
		}
	} else {
		for keyword, lang := range bangumitypes.SubTitleLangKeyword {
			if strings.Contains(filename, keyword) {
				ret.Resource.SubtitleLang = append(ret.Resource.SubtitleLang, lang)
			}
		}
	}

	if len(parsedElements.Subtitles) > 0 {
		for _, l := range parsedElements.Subtitles {
			ret.Resource.SubtitleLang = append(ret.Resource.SubtitleLang, normalizationLang(l))
		}
	} else {
		for keyword, lang := range bangumitypes.SubTitleLangKeyword {
			if strings.Contains(filename, keyword) {
				ret.Resource.SubtitleLang = append(ret.Resource.SubtitleLang, lang)
			}
		}
	}
	if len(ret.Resource.SubtitleLang) == 0 {
		ret.Resource.SubtitleLang = append(ret.Resource.SubtitleLang, bangumitypes.SubtitleUnknown)
	}
	ret.Resource.SubtitleLang = removeDuplicateSubtitleLang(ret.Resource.SubtitleLang)

	if len(parsedElements.AnimeType) == 0 {
		ret.Resource.Type = bangumitypes.ResourceTypeNone
	} else {
		ret.Resource.Type = normalizationEpisodeType(parsedElements.AnimeType[0])
	}

	ret.Resource.Resolution = normalizationResolution(parsedElements.VideoResolution)

	if parsedElements.AnimeTitle != "" {
		ret.Title = strings.Split(parsedElements.AnimeTitle, "/")[0]
	}

	return &ret, nil
}

func (parser *MikanRSSParser) parseMikanRSS(mikan *MikanRss) ([]*Bangumi, error) {
	bangumiMap := make(map[int64]*Bangumi)
	var parseResultList []*ParseItemResult
	for i, item := range mikan.Channel.Item {
		if item.Link != "" {
			parser.logger.Debug().Str("title", item.Title).Msg(fmt.Sprintf("parse Episode %d/%d", i+1, len(mikan.Channel.Item)))
			parseResult, err := parser.parserItemLink(item)
			if err != nil {
				parser.logger.Err(err).Msg("parse item err")
			} else {
				parseResultList = append(parseResultList, parseResult)
			}
		}
	}

	for _, pr := range parseResultList {
		var foundBgm bool
		var bgm *Bangumi
		if bgm, foundBgm = bangumiMap[pr.SubjectId]; !foundBgm {
			bgm = &Bangumi{
				Info: BangumiInfo{
					Title:   pr.Title,
					TmDBId:  pr.TmDBId,
					MikanID: pr.MikanBangumiId,
				},
				Seasons: make(map[uint]Season),
			}
		}

		var foundSeason bool
		var season Season
		if season, foundSeason = bgm.Seasons[pr.SeasonNum]; !foundSeason {
			season = Season{
				SubjectId:      pr.SubjectId,
				MikanBangumiId: pr.MikanBangumiId,
				Number:         pr.SeasonNum,
				EpCount:        pr.EpCount,
				Episodes:       make(map[uint]Episode),
			}
		}

		var foundEpisode bool
		var episode Episode

		if episode, foundEpisode = season.Episodes[pr.EpNum]; !foundEpisode {
			episode = Episode{
				Number:    pr.EpNum,
				Resources: make([]TorrentResource, 0),
			}
		}
		episode.Resources = append(episode.Resources, pr.Resource)
		season.Episodes[pr.EpNum] = episode
		season.RemoveInvalidEpisode()
		bgm.Seasons[pr.SeasonNum] = season

		bangumiMap[pr.SubjectId] = bgm
	}

	return maps.Values(bangumiMap), nil
}

func removeDuplicateSubtitleLang(langs []bangumitypes.SubtitleLang) []bangumitypes.SubtitleLang {
	allKeys := make(map[bangumitypes.SubtitleLang]bool)
	var list []bangumitypes.SubtitleLang
	for _, item := range langs {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func normalizationResolution(resolution string) bangumitypes.Resolution {
	switch strings.ToLower(resolution) {
	case "2160p":
		return bangumitypes.Resolution2160p
	case "1080p", "1920x1080":
		return bangumitypes.Resolution1080p
	case "720p", "1024x720", "1280x720":
		return bangumitypes.Resolution720p
	default:
		return bangumitypes.ResolutionUnknown
	}
}

func normalizationLang(lang string) bangumitypes.SubtitleLang {
	for k, v := range bangumitypes.SubTitleLangKeyword {
		if strings.Contains(lang, k) {
			return v
		}
	}
	return bangumitypes.SubtitleUnknown
}

func normalizationEpisodeType(epType string) bangumitypes.ResourceType {
	et := bangumitypes.ResourceType(strings.ToUpper(epType))
	switch et {
	case bangumitypes.ResourceTypeSP, bangumitypes.ResourceTypeOVA, bangumitypes.ResourceTypeSpecial:
		return et
	default:
		return bangumitypes.ResourceTypeUnknown
	}
}
