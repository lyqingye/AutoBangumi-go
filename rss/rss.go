package rss

import (
	bangumitypes "autobangumi-go/bangumi"
	"autobangumi-go/bus"
	"autobangumi-go/db"
	"autobangumi-go/mdb"
	"autobangumi-go/rss/mikan"
	"autobangumi-go/utils"
	"sync"
	"time"

	"github.com/rs/zerolog"
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
	bgmMan          *bangumitypes.BangumiManager
}

func NewRSSManager(bgmMan *bangumitypes.BangumiManager, eb *bus.EventBus, db *db.DB, period time.Duration, tmdbClient *mdb.TMDBClient, bangumiTVClient *mdb.BangumiTVClient) (*RSSManager, error) {
	man := RSSManager{
		eb:              eb,
		db:              db,
		logger:          utils.GetLogger("RSSManager"),
		refreshLock:     sync.Mutex{},
		tmdb:            tmdbClient,
		bangumiTvClient: bangumiTVClient,
		bgmMan:          bgmMan,
	}

	err := man.watchNewBangumi()
	if err != nil {
		return nil, err
	}
	man.ticker = time.NewTicker(period)
	return &man, nil
}

func (man *RSSManager) watchNewBangumi() error {
	parser, err := mikan.NewMikanRSSParser("https://mikanani.me", man.eb, man.db, man.tmdb, man.bangumiTvClient)
	if err != nil {
		return err
	}
	man.eb.SubscribeWithFn(bus.BangumiManTopic, func(event bus.Event) {
		if event.EventType == bus.BangumiManAddNewEvent {
			man.logger.Info().Msg("recv add new bangumi event")
			if bangumi, ok := event.Inner.(bangumitypes.Bangumi); ok {
				man.logger.Info().Str("title", bangumi.Info.Title).Msg("add new bangumi")
				man.bgmMan.GetAndLockInCompleteBangumi(bangumi.Info.Title, func(bgmMan *bangumitypes.BangumiManager, bangumi *bangumitypes.Bangumi) {
					man.refreshBangumi(parser, bgmMan, bangumi)
				})
			}
		}
	})
	return nil
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
	man.bgmMan.IterInCompleteBangumi(func(bgmMan *bangumitypes.BangumiManager, bangumi *bangumitypes.Bangumi) bool {
		man.refreshBangumi(parser, bgmMan, bangumi)
		return false
	})
	return nil
}

func (man *RSSManager) refreshBangumi(parser *mikan.MikanRSSParser, bgmMan *bangumitypes.BangumiManager, bangumi *bangumitypes.Bangumi) {
	man.logger.Info().Str("title", bangumi.Info.Title).Msg("refresh bangumi")
	err := parser.CompleteBangumi(bangumi)
	if err == nil {
		_ = bgmMan.Flush(bangumi)
		man.eb.Publish(bus.RSSTopic, bus.Event{
			EventType: bus.RSSUpdateEventType,
			Inner:     *bangumi,
		})
	} else {
		man.logger.Error().Err(err).Str("title", bangumi.Info.Title).Msg("complete bangumi error")
	}
}
