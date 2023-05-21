package mikan

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	bangumitypes "pikpak-bot/bangumi"
	"pikpak-bot/mdb"
	"pikpak-bot/utils"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	torrent "github.com/anacrolix/torrent/metainfo"
	"github.com/dustin/go-humanize"
	"github.com/nssteinbrenner/anitogo"
)

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
	var bangumiInfo bangumitypes.BangumiInfo
	var seasonNumber uint
	var epCount uint
	var mikanBangumiId string
	var subjectId int64

	cache, found := parser.getParseCache(item.Link)
	if !found {
		// Episode information from rss item
		fileSize, err := humanize.ParseBytes(item.Torrent.ContentLength)
		if err == nil {
			episode.FileSize = fileSize
		}
		episode.RawFilename = item.Title
		pubDate, err := utils.SmartParseDate(item.Torrent.PubDate)
		if err != nil {
			return err
		}
		episode.TorrentPubDate = pubDate

		// Parse Episode from item description webpage, result information is reliable
		// - Title (nullable)
		// - Mikan bangumi Id (nullable)
		// - BangumiTV subject Id (nullable)
		// - Torrent
		// - TorrentHash
		// - Magnet
		fromWebPage, err := parser.parseEpisodeByItem(item.Link)
		if err != nil {
			parser.logger.Warn().Err(err).Str("link", item.Link).Str("title", item.Title).Msg("parse episode error")
			return err
		}

		episode.Magnet = fromWebPage.Episode.Magnet
		episode.Torrent = fromWebPage.Episode.Torrent
		episode.TorrentHash = fromWebPage.Episode.TorrentHash
		mikanBangumiId = fromWebPage.MikanBangumiId
		subjectId = fromWebPage.SubjectId

		// episode.MikanBangumiId = episodeFromItem.MikanBangumiId
		// episode.SubjectId = episodeFromItem.SubjectId

		// Parse Episode from filename, result information is unreliable
		// - BangumiTitle (nullable)
		// - Eposide Number (nullable)
		// - Season Number (nullable)
		// - Lang (nullable)
		// - Resolution (nullable)
		// - EpisodeType (nullable)
		// - Subgroup (nullable)
		fromFilename, err := parser.parseEpisodeByFilename(item.Title)
		if err != nil {
			return err
		}
		if fromFilename.Episode.Number == 0 {
			parser.logger.Warn().Str("filename", item.Title).Msg("parse episode number from filename err")
			return errors.New("parse episode number from filename err")
		}

		episode.Number = fromFilename.Episode.Number
		episode.SubtitleLang = fromFilename.Episode.SubtitleLang
		episode.Resolution = fromFilename.Episode.Resolution
		episode.Subgroup = fromFilename.Episode.Subgroup
		episode.Type = fromFilename.Episode.Type

		if fromWebPage.Bangumi.Title == "" && fromFilename.Bangumi.Title == "" {
			return errors.New("could not found title from link page and filename")
		}

		if fromWebPage.Bangumi.Title != "" {
			bangumiInfo.Title = fromWebPage.Bangumi.Title
		} else if fromFilename.Bangumi.Title != "" {
			bangumiInfo.Title = fromFilename.Bangumi.Title
		}
	} else {
		episode = cache.Episode
		bangumiInfo = cache.Bangumi
		seasonNumber = cache.Season
		epCount = cache.EpCount
		subjectId = cache.SubjectId
		mikanBangumiId = cache.MikanBangumiId
	}

	// Try predict Episode season number using bangumi TV and tmdb
	// 1. get air date from bangumi tv
	// 2. get seasons from tmdb
	// 3. using air data to predict season number
	//
	if subjectId == 0 || seasonNumber == 0 {
		// the subject id comes from parsing item link page
		// if the subgroup does not have a link associated with bangumitv when publishing resources
		// then we will try searching based on the title
		var subject *mdb.Subjects
		if subjectId != 0 {
			subject, err = parser.getBangumiTVSubject(subjectId)
			if err != nil {
				return err
			}
		} else {
			subject, err = parser.searchBangumiTV(bangumiInfo.Title)
			if err != nil {
				return err
			}
			subjectId = subject.ID

			// cache mikan bangumiId -> BangumiTV Id
			if mikanBangumiId != "" {
				_ = parser.db.Set(getMikanBangumiToBangumiTVCache(mikanBangumiId), &subjectId)
			}
		}

		// now we get episode air date
		subjectAirDate, err := utils.ParseDate(subject.Date)

		var searchTitles []string
		searchTitles = append(searchTitles, bangumiInfo.Title, subject.NameCn, subject.Name)
		searchTitles = append(searchTitles, subject.GetAliasNames()...)
		if err != nil {
			return err
		}

		// try search seasons from tmdb
		for _, searchTitle := range searchTitles {
			if searchTitle == "" {
				continue
			}
			tvDetails, err := parser.searchTMDB(searchTitle)
			if err == nil {
				bangumiInfo.TmDBId = tvDetails.ID
				// using tmdb Title as bangumi title
				bangumiInfo.Title = tvDetails.Name

				// predict season number using air date
				minDiff := time.Duration(math.MaxInt64)
				closeIndex := -1
				for i, season := range tvDetails.Seasons {
					// NOTE:
					// If season number is 0 , then the season maybe a Special or TV
					if season.SeasonNumber == 0 {
						continue
					}

					// using episode air date to predict episode season
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
					seasonNumber = uint(tvDetails.Seasons[closeIndex].SeasonNumber)
					epCount = uint(tvDetails.Seasons[closeIndex].EpisodeCount)
				}

				if seasonNumber != 0 {
					break
				}

			} else {
				parser.logger.Warn().Err(err).Msg("search tmdb error")
			}
		}
	}

	if seasonNumber == 0 {
		err = fmt.Errorf("unknown season, mikan link: %s", item.Link)
		parser.logger.Err(err).Msg("parse error")
		return err
	}

	var bangumi *bangumitypes.Bangumi
	if existBangumi, found := cacheBangumi[bangumiInfo.TmDBId]; found {
		bangumi = existBangumi
	} else {
		bangumi = &bangumitypes.Bangumi{
			Info:    bangumiInfo,
			Seasons: make(map[uint]bangumitypes.Season),
		}
	}

	if err := episode.Validate(); err != nil {
		parser.logger.Warn().Err(err).Str("link", item.Link).Str("title", item.Title).Msg("parse episode error")
		return err
	}

	seasonInfo := bangumi.Seasons[seasonNumber]
	seasonInfo.MikanBangumiId = mikanBangumiId
	seasonInfo.SubjectId = subjectId
	seasonInfo.Number = seasonNumber
	seasonInfo.EpCount = epCount
	seasonInfo.Episodes = append(seasonInfo.Episodes, episode)

	bangumi.Info = bangumiInfo
	bangumi.Seasons[seasonNumber] = seasonInfo

	cacheBangumi[bangumiInfo.TmDBId] = bangumi

	// save parse cache
	cache.Season = seasonNumber
	cache.EpCount = epCount
	cache.SubjectId = subjectId
	cache.MikanBangumiId = mikanBangumiId
	cache.Bangumi = bangumiInfo
	cache.Episode = episode

	parser.storeParseCache(item.Link, cache)
	return nil
}

type ParseItemResult struct {
	Bangumi        bangumitypes.BangumiInfo
	Episode        bangumitypes.Episode
	Season         uint
	EpCount        uint
	SubjectId      int64
	MikanBangumiId string
}

func (parser *MikanRSSParser) parseEpisodeByItem(link string) (*ParseItemResult, error) {
	result := ParseItemResult{}
	bangumi := bangumitypes.BangumiInfo{}
	episode := bangumitypes.Episode{}

	resp, err := parser.http.R().EnableTrace().Get(link)
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
	bangumi.Title = titleSelector.Text()

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

	episode.Subgroup = subGroupSelector.Text()
	for _, node := range buttonSelector.Nodes {
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				if strings.HasSuffix(attr.Val, ".torrent") {
					torrentDownloadUrl := parser.mikanEndpoint.JoinPath(attr.Val)
					resp, err := parser.http.R().Get(torrentDownloadUrl.String())
					if err != nil {
						return nil, err
					}
					episode.Torrent = resp.Body()
				}
				if strings.HasPrefix(attr.Val, "magnet:?xt") {
					episode.Magnet = attr.Val
				}
			}
		}
	}

	if mikanBangumiLink != "" {
		if idx := strings.Index(mikanBangumiLink, "#"); idx != -1 {
			mikanBangumiLink = mikanBangumiLink[:idx]
		}
		result.MikanBangumiId = strings.ReplaceAll(mikanBangumiLink, "/Home/Bangumi/", "")
		key := getMikanBangumiToBangumiTVCache(result.MikanBangumiId)
		var cachedSubjectId int64
		var cached bool
		if result.MikanBangumiId != "" {
			cached, err = parser.db.Get(key, &cachedSubjectId)
		}
		if err == nil && cached {
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
				_ = parser.db.Set(key, &subjectId)
			}
		}
	}

	torr, err := torrent.Load(bytes.NewBuffer(episode.Torrent))
	if err != nil {
		return nil, err
	}
	episode.TorrentHash = torr.HashInfoBytes().HexString()
	result.Bangumi = bangumi
	result.Episode = episode
	return &result, nil
}

func (parser *MikanRSSParser) parseEpisodeByFilename(filename string) (*ParseItemResult, error) {
	bangumi := bangumitypes.BangumiInfo{}
	episode := bangumitypes.Episode{}
	var seasonNumber uint

	parsedElements := anitogo.Parse(filename, anitogo.DefaultOptions)

	if len(parsedElements.EpisodeNumber) > 0 {
		epStr := parsedElements.EpisodeNumber[0]
		epNumber, err := strconv.ParseUint(epStr, 10, 32)
		if err == nil && epNumber > 0 {
			episode.Number = uint(epNumber)
		}
	}

	if len(parsedElements.AnimeSeason) > 0 {
		seasonStr := parsedElements.AnimeSeason[0]
		season, err := strconv.ParseUint(seasonStr, 10, 32)
		if err == nil && season > 0 {
			seasonNumber = uint(season)
		}
	}

	if len(parsedElements.Language) > 0 {
		for _, l := range parsedElements.Language {
			episode.SubtitleLang = append(episode.SubtitleLang, normalizationLang(l))
		}
	}

	if len(parsedElements.Subtitles) > 0 {
		for _, l := range parsedElements.Subtitles {
			episode.SubtitleLang = append(episode.SubtitleLang, normalizationLang(l))
		}
	}

	if len(parsedElements.AnimeType) == 0 {
		episode.Type = bangumitypes.EpisodeTypeNone
	} else {
		episode.Type = normalizationEpisodeType(parsedElements.AnimeType[0])
	}

	if len(episode.SubtitleLang) == 0 {
		episode.SubtitleLang = []string{bangumitypes.SubtitleUnknown}
	}
	episode.Resolution = normalizationResolution(parsedElements.VideoResolution)

	if parsedElements.AnimeTitle != "" {
		bangumi.Title = strings.Split(parsedElements.AnimeTitle, "/")[0]
	}

	return &ParseItemResult{
		Bangumi: bangumi,
		Episode: episode,
		Season:  seasonNumber,
	}, nil
}

func (parser *MikanRSSParser) parseMikanRSS(mikan *MikanRss) ([]bangumitypes.Bangumi, error) {
	bangumiMap := make(map[int64]*bangumitypes.Bangumi)
	for i, item := range mikan.Channel.Item {
		if item.Link != "" {
			parser.logger.Debug().Str("title", item.Title).Msg(fmt.Sprintf("parse Episode %d/%d", i+1, len(mikan.Channel.Item)))
			err := parser.parserItemLink(item, bangumiMap)
			if err != nil {
				parser.logger.Err(err).Msg("parse item err")
			}
		}
	}
	var bangumis []bangumitypes.Bangumi
	for _, bangumi := range bangumiMap {
		bangumis = append(bangumis, *bangumi)
	}
	filterBangumi(bangumis)
	return bangumis, nil
}

func normalizationResolution(resolution string) string {
	switch resolution {
	case "2160P", "2160p":
		return bangumitypes.Resolution2160p
	case "1080P", "1080p", "1920x1080", "1920X1080":
		return bangumitypes.Resolution1080p
	case "720P", "720p", "1024x720", "1280x720":
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

func normalizationEpisodeType(epType string) string {
	epType = strings.ToUpper(epType)
	switch epType {
	case bangumitypes.EpisodeTypeSP, bangumitypes.EpisodeTypeOVA, bangumitypes.EpisodeTypeSpecial:
		return epType
	default:
		return bangumitypes.EpisodeTypeUnknown
	}
}
