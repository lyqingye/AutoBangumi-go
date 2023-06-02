package rss

import (
	"sync"
	"time"

	bangumitypes "autobangumi-go/bangumi"
	"autobangumi-go/bus"
	"autobangumi-go/db"
	"autobangumi-go/mdb"
	"autobangumi-go/utils"

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
// - (SubjectId,SeasonNum,Episode) -> Episode

type RSSManager struct {
	eb              *bus.EventBus
	db              *db.DB
	ticker          *time.Ticker
	logger          zerolog.Logger
	refreshLock     sync.Mutex
	tmdb            *mdb.TMDBClient
	bangumiTvClient *mdb.BangumiTVClient
	bgmMan          *bangumitypes.Manager
}

func NewRSSManager(bgmMan *bangumitypes.Manager, eb *bus.EventBus, db *db.DB, period time.Duration, tmdbClient *mdb.TMDBClient, bangumiTVClient *mdb.BangumiTVClient) (*RSSManager, error) {
	man := RSSManager{
		eb:              eb,
		db:              db,
		logger:          utils.GetLogger("RSSManager"),
		refreshLock:     sync.Mutex{},
		tmdb:            tmdbClient,
		bangumiTvClient: bangumiTVClient,
		bgmMan:          bgmMan,
	}
	man.ticker = time.NewTicker(period)
	return &man, nil
}
