package mikan

import (
	"encoding/xml"
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
	mikan, err := parser.getRss(parser.rssLink)
	if err != nil {
		return nil, err
	}
	return parser.parseMikanRSS(mikan)
}

func (parser *MikanRSSParser) getRss(link string) (*MikanRss, error) {
	resp, err := parser.http.R().EnableTrace().Get(link)
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
	for i, bgm := range bangumis {
		for seasonNumber, season := range bgm.Seasons {
			epGroupByNumber := make(map[uint][]bangumitypes.Episode)
			for _, ep := range season.Episodes {
				if ep.Number > season.EpCount {
					continue
				}
				if ep.Type != bangumitypes.EpisodeTypeNone {
					continue
				}
				dupEps := epGroupByNumber[ep.Number]
				dupEps = append(dupEps, ep)
				epGroupByNumber[ep.Number] = dupEps
			}

			var filteredEps []bangumitypes.Episode

			for _, eps := range epGroupByNumber {
				filteredEps = append(filteredEps, selectEpisode(eps))
			}

			sort.Slice(filteredEps, func(i, j int) bool {
				return filteredEps[i].Number < filteredEps[j].Number
			})

			season.Episodes = filteredEps
			bgm.Seasons[seasonNumber] = season
		}
		bangumis[i] = bgm
	}
}

func selectEpisode(episodes []bangumitypes.Episode) bangumitypes.Episode {
	if len(episodes) == 1 {
		return episodes[0]
	} else {
		// priority: resolution
		sort.Slice(episodes, func(i, j int) bool {
			return episodes[i].Compare(&episodes[j])
		})
		return episodes[0]
	}
}
