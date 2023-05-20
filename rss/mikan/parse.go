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
	cache, found := parser.getParseCache(item.Link)
	if !found {
		// Episode information from rss item
		fileSize, err := humanize.ParseBytes(item.Torrent.ContentLength)
		if err == nil {
			episode.FileSize = fileSize
		}
		episode.EpisodeTitle = item.Title
		episode.Date = item.Torrent.PubDate
		episode.EpisodeTitle = item.Title

		// Parse Episode from item description webpage, result information is reliable
		// - Title (nullable)
		// - Mikan bangumi Id (nullable)
		// - BangumiTV subject Id (nullable)
		// - Torrent
		// - TorrentHash
		// - Magnet
		episodeFromItem, err := parser.parseEpisodeByItem(item.Link)
		if err != nil {
			parser.logger.Warn().Err(err).Str("link", item.Link).Str("title", item.Title).Msg("parse episode error")
			return err
		}
		episode.Magnet = episodeFromItem.Magnet
		episode.Torrent = episodeFromItem.Torrent
		episode.TorrentHash = episodeFromItem.TorrentHash
		episode.MikanBangumiId = episodeFromItem.MikanBangumiId
		episode.SubjectId = episodeFromItem.SubjectId

		// Parse Episode from filename, result information is unreliable
		// - BangumiTitle (nullable)
		// - Eposide Number (nullable)
		// - Season Number (nullable)
		// - Lang (nullable)
		// - Resolution (nullable)
		// - EpisodeType (nullable)
		// - Subgroup (nullable)
		episodeFromFilename, err := parser.parseEpisodeByFilename(item.Title)
		if err != nil {
			return err
		}
		if episodeFromFilename.EPNumber == 0 {
			parser.logger.Warn().Str("filename", item.Title).Msg("parse episode number from filename err")
			return errors.New("parse episode number from filename err")
		}
		episode.Season = episodeFromFilename.Season
		episode.EPNumber = episodeFromFilename.EPNumber
		episode.Lang = episodeFromFilename.Lang
		episode.Resolution = episodeFromFilename.Resolution
		episode.Subgroup = episodeFromFilename.Subgroup
		episode.EpisodeType = episodeFromFilename.EpisodeType

		if episodeFromItem.BangumiTitle == "" && episodeFromFilename.BangumiTitle == "" {
			return errors.New("could not found title from link page and filename")
		}

		if episodeFromItem.BangumiTitle != "" {
			episode.BangumiTitle = episodeFromItem.BangumiTitle
		} else if episodeFromFilename.BangumiTitle != "" {
			episode.BangumiTitle = episodeFromFilename.BangumiTitle
		}

		cache.Episode = episode
	} else {
		episode = cache.Episode
	}

	var bangumi = &bangumitypes.Bangumi{
		Title:     episode.BangumiTitle,
		Season:    cache.Season,
		TmDBId:    cache.TMDBId,
		SubjectId: cache.SubjectId,
		EPCount:   cache.EPCount,
	}

	// Try predict Episode season number using bangumi TV and tmdb
	// 1. get air date from bangumi tv
	// 2. get seasons from tmdb
	// 3. using air data to predict season number
	//
	if episode.SubjectId == 0 || episode.Season == 0 {
		// the subject id comes from parsing item link page
		// if the subgroup does not have a link associated with bangumitv when publishing resources
		// then we will try searching based on the title
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
			if episode.MikanBangumiId != "" {
				_ = parser.db.Set(getMikanBangumiToBangumiTVCache(episode.MikanBangumiId), &episode.SubjectId)
			}
		}

		// now we get episode air date
		subjectAirDate, err := utils.ParseDate(subject.Date)

		var searchTitles []string
		searchTitles = append(searchTitles, episode.BangumiTitle, subject.NameCn, subject.Name)
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
				cache.TMDBId = tvDetails.ID
				// using tmdb Title as bangumi title
				cache.BangumiTitle = tvDetails.Name

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
					cache.Season = uint(tvDetails.Seasons[closeIndex].SeasonNumber)
					cache.EPCount = uint(tvDetails.Seasons[closeIndex].EpisodeCount)
				}

				if cache.Season != 0 {
					break
				}

			} else {
				parser.logger.Warn().Err(err).Msg("search tmdb error")
			}
		}
	}

	if cache.Season == 0 {
		err = fmt.Errorf("unknown season, mikan link: %s", item.Link)
		parser.logger.Err(err).Msg("parse error")
		return err
	}

	if existBangumi, found := cacheBangumi[episode.SubjectId]; found {
		bangumi = existBangumi
	}

	bangumi.Season = cache.Season
	bangumi.EPCount = cache.EPCount
	bangumi.TmDBId = cache.TMDBId
	bangumi.SubjectId = cache.SubjectId
	bangumi.Title = cache.BangumiTitle

	episode.Season = bangumi.Season
	episode.BangumiTitle = bangumi.Title
	episode.SubjectId = bangumi.SubjectId

	if err := episode.Validate(); err != nil {
		parser.logger.Warn().Err(err).Str("link", item.Link).Str("title", item.Title).Msg("parse episode error")
		return err
	}

	bangumi.Episodes = append(bangumi.Episodes, episode)
	cacheBangumi[episode.SubjectId] = bangumi

	// save parse cache
	parser.storeParseCache(item.Link, cache)
	return nil
}

func (parser *MikanRSSParser) parseEpisodeByItem(link string) (bangumitypes.Episode, error) {
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
		var cached bool
		if ep.MikanBangumiId != "" {
			cached, err = parser.db.Get(key, &cachedSubjectId)
		}
		if err == nil && cached {
			ep.SubjectId = cachedSubjectId
		} else {
			resp, err = parser.http.R().Get(parser.mikanEndpoint.JoinPath(mikanBangumiLink).String())
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

func (parser *MikanRSSParser) parseEpisodeByFilename(filename string) (bangumitypes.Episode, error) {
	episode := bangumitypes.Episode{}
	parsedElements := anitogo.Parse(filename, anitogo.DefaultOptions)

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

	if len(parsedElements.AnimeType) == 0 {
		episode.EpisodeType = bangumitypes.EpisodeTypeNone
	} else {
		episode.EpisodeType = normalizationEpisodeType(parsedElements.AnimeType[0])
	}

	if len(episode.Lang) == 0 {
		episode.Lang = []string{bangumitypes.SubtitleUnknown}
	}
	episode.Resolution = normalizationResolution(parsedElements.VideoResolution)

	if parsedElements.AnimeTitle != "" {
		episode.BangumiTitle = strings.Split(parsedElements.AnimeTitle, "/")[0]
	}

	return episode, nil
}

func (parser *MikanRSSParser) parseMikanRSS(mikan *MikanRss) ([]bangumitypes.Bangumi, error) {
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
