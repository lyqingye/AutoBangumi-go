package mikan

import (
	"encoding/xml"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"net/url"
	bangumitypes "pikpak-bot/bangumi"
	"pikpak-bot/bus"
	"pikpak-bot/db"
	"pikpak-bot/mdb"
	"pikpak-bot/utils"
	"sort"
)

type MikanRSSParser struct {
	mikanEndpoint   *url.URL
	rssLink         string
	http            *resty.Client
	eb              *bus.EventBus
	db              *db.DB
	logger          zerolog.Logger
	tmdb            *mdb.TMDBClient
	bangumiTvClient *mdb.BangumiTVClient
}

func NewMikanRSSParser(rss string, eb *bus.EventBus, parseCacheDB *db.DB, tmdbClient *mdb.TMDBClient, bangumiTVClient *mdb.BangumiTVClient) (*MikanRSSParser, error) {
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

func (parser *MikanRSSParser) RssLink() string {
	return parser.rssLink
}

func (parser *MikanRSSParser) Parse() ([]bangumitypes.Bangumi, error) {
	var err error
	mikan, err := parser.getRss()
	if err != nil {
		return nil, err
	}
	return parser.parseMikanRSS(mikan)
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

func filterBangumi(bangumis []bangumitypes.Bangumi) {
	for i, b := range bangumis {
		epMap := make(map[string][]bangumitypes.Episode)
		for _, ep := range b.Episodes {
			if ep.EPNumber > b.EPCount {
				continue
			}
			if ep.EpisodeType != bangumitypes.EpisodeTypeNone {
				continue
			}
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
					result := bangumitypes.ResolutionPriority[epI] > bangumitypes.ResolutionPriority[epJ] || bangumitypes.SubtitlePriority[langI] > bangumitypes.SubtitlePriority[langJ]
					if err1 == nil && err2 == nil {
						result = result || iDate.Unix() > jDate.Unix()
					}
					return result
				})
				processedEps = append(processedEps, eps[0])
			}
		}
		sort.Slice(processedEps, func(i, j int) bool {
			return processedEps[i].EPNumber < processedEps[j].EPNumber
		})
		bangumis[i].Episodes = processedEps
	}
}
