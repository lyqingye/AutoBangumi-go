package rss

import (
	bangumitypes "pikpak-bot/bangumi"
	"pikpak-bot/bus"
	"pikpak-bot/db"
	"pikpak-bot/mdb"
	"pikpak-bot/rss/mikan"
	"pikpak-bot/utils"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	KeyRSSLink          = []byte("rss-links")
	KeyRSSSubscribeInfo = []byte("rss-subscribe")
)

type SubscribeInfo struct {
	RSSLink  string
	Bangumis []bangumitypes.Bangumi
}

// RSSManager
// Bangumi Storage Structure
// Indexes:
// - RssLink -> BangumiInfo
// - SubjectId -> BangumiInfo
// - (SubjectId,Season,Episode) -> Episode

type RSSManager struct {
	eb              *bus.EventBus
	db              *db.DB
	ticker          *time.Ticker
	logger          zerolog.Logger
	refreshLock     sync.Mutex
	tmdb            *mdb.TMDBClient
	bangumiTvClient *mdb.BangumiTVClient
	stateLock       sync.Mutex
	inComplete      map[string]bangumitypes.Bangumi
	complete        map[string]bangumitypes.Bangumi
}

func NewRSSManager(eb *bus.EventBus, db *db.DB, period time.Duration, tmdbClient *mdb.TMDBClient, bangumiTVClient *mdb.BangumiTVClient) (*RSSManager, error) {
	man := RSSManager{
		eb:              eb,
		db:              db,
		logger:          utils.GetLogger("RSSManager"),
		refreshLock:     sync.Mutex{},
		tmdb:            tmdbClient,
		bangumiTvClient: bangumiTVClient,
		stateLock:       sync.Mutex{},
		complete:        make(map[string]bangumitypes.Bangumi),
		inComplete:      make(map[string]bangumitypes.Bangumi),
	}
	man.ticker = time.NewTicker(period)
	return &man, nil
}

func (man *RSSManager) Start() {
	man.logger.Info().Msg("start rss manager")
	man.Refresh()
	for range man.ticker.C {
		man.Refresh()
	}
}

func (man *RSSManager) Refresh() {
	err := man.refreshInComplete()
	if err != nil {
		man.logger.Error().Err(err).Msg("refresh incomplete err")
	}
}

func (man *RSSManager) refreshInComplete() error {
	man.stateLock.Lock()
	defer man.stateLock.Unlock()
	parser, err := mikan.NewMikanRSSParser("https://mikanani.me", man.eb, man.db, man.tmdb, man.bangumiTvClient)
	if err != nil {
		return err
	}
	for title, bangumi := range man.inComplete {
		err = parser.CompleteBangumi(&bangumi)
		if err != nil {
			continue
		}
		man.inComplete[title] = bangumi
		man.eb.Publish(bus.RSSTopic, bus.Event{
			EventType: bus.RSSUpdateEventType,
			Inner:     bangumi,
		})
	}
	return nil
}

func (man *RSSManager) MarkEpisodeComplete(info *bangumitypes.BangumiInfo, seasonNum uint, episode bangumitypes.Episode) {
	man.stateLock.Lock()
	defer man.stateLock.Unlock()
	if bangumi, found := man.inComplete[info.Title]; found {
		if season, foundSeason := bangumi.Seasons[seasonNum]; foundSeason {
			if !season.IsComplete(episode.Number) {
				season.Complete = append(season.Complete, episode.Number)
				bangumi.Seasons[seasonNum] = season
			}
		}
		man.inComplete[info.Title] = bangumi
	}
}
