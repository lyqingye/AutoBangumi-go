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
	bgmMan      *bangumitypes.BangumiManager
}

func NewRSSManager(bgmMan *bangumitypes.BangumiManager,eb *bus.EventBus, db *db.DB, period time.Duration, tmdbClient *mdb.TMDBClient, bangumiTVClient *mdb.BangumiTVClient) (*RSSManager, error) {
	man := RSSManager{
		eb:              eb,
		db:              db,
		logger:          utils.GetLogger("RSSManager"),
		refreshLock:     sync.Mutex{},
		tmdb:            tmdbClient,
		bangumiTvClient: bangumiTVClient,
		bgmMan: bgmMan,
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
	parser, err := mikan.NewMikanRSSParser("https://mikanani.me", man.eb, man.db, man.tmdb, man.bangumiTvClient)
	if err != nil {
		return err
	}
	eb := man.eb
	man.bgmMan.IterInCompleteBangumi(func(man *bangumitypes.BangumiManager, bangumi *bangumitypes.Bangumi) bool {
		err = parser.CompleteBangumi(bangumi)
		if err == nil {
			eb.Publish(bus.RSSTopic, bus.Event{
				EventType: bus.RSSUpdateEventType,
				Inner:     bangumi,
			})
		} else {
			// TODO: logger
		}
		return false
	})
	return nil
}
