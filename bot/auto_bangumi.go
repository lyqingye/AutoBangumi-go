package bot

import (
	"autobangumi-go/bangumi"
	"autobangumi-go/bus"
	"autobangumi-go/db"
	"autobangumi-go/downloader"
	"autobangumi-go/downloader/aria2"
	"autobangumi-go/downloader/pikpak"
	"autobangumi-go/mdb"
	"autobangumi-go/rss"
	"autobangumi-go/utils"
	"errors"
	"net/url"
	"time"

	"github.com/rs/zerolog"
)

type AutoBangumiConfig struct {
	QBEndpoint           string
	QBUsername           string
	QBPassword           string
	QBDownloadDir        string
	DBDir                string
	BangumiTVApiEndpoint string
	TMDBToken            string
	RSSUpdatePeriod      time.Duration
	BangumiHome          string
	Aria2WsUrl           string
	Aria2Secret          string
	Aria2DownloadDir     string
	PikPakConfigPath     string
}

func (config *AutoBangumiConfig) Validate() error {
	_, err := url.Parse(config.QBEndpoint)
	if err != nil {
		return err
	}
	if config.QBUsername == "" || config.QBPassword == "" {
		return errors.New("empty qb username or password")
	}

	if config.Aria2WsUrl == "" {
		return errors.New("empty aria2 ws url")
	}

	if config.BangumiTVApiEndpoint == "" {
		config.BangumiTVApiEndpoint = "https://api.bgm.tv/v0"
	}

	if config.TMDBToken == "" {
		return errors.New("tmdb token is empty")
	}

	if config.BangumiHome == "" {
		return errors.New("empty bangumi home")
	}

	return nil
}

type AutoBangumi struct {
	qb     *qbittorrent.QbittorrentClient
	db     *db.DB
	eb     *bus.EventBus
	rssMan *rss.RSSManager
	logger zerolog.Logger
	bgmMan *bangumi.BangumiManager
	bgmTV  *mdb.BangumiTVClient
	tmdb   *mdb.TMDBClient
	dl     *downloader.SmartDownloader
	aria2  *aria2.Client
}

func NewAutoBangumi(config *AutoBangumiConfig) (*AutoBangumi, error) {
	bot := AutoBangumi{}
	bot.logger = utils.GetLogger("AutoBangumi")

	// database
	database, err := db.NewDB(config.DBDir)
	if err != nil {
		return nil, err
	}
	eb := bus.NewEventBus()
	bot.eb = eb

	// bangumi TV
	bangumiTVClient, err := mdb.NewBangumiTVClient(config.BangumiTVApiEndpoint)
	if err != nil {
		return nil, err
	}
	bot.bgmTV = bangumiTVClient

	// TMDB
	tmdbClient, err := mdb.NewTMDBClient(config.TMDBToken)
	if err != nil {
		return nil, err
	}
	bot.tmdb = tmdbClient

	// bangumi manager
	bgmMan, err := bangumi.NewBangumiManager(config.BangumiHome, eb)
	if err != nil {
		return nil, err
	}

	// rss manager
	rssMan, err := rss.NewRSSManager(bgmMan, eb, database, config.RSSUpdatePeriod, tmdbClient, bangumiTVClient)
	if err != nil {
		return nil, err
	}
	bot.db = database
	bot.rssMan = rssMan
	bot.bgmMan = bgmMan

	// qb
	qb, err := qbittorrent.NewQbittorrentClient(config.QBEndpoint, config.QBUsername, config.QBPassword, config.QBDownloadDir)
	if err != nil {
		return nil, err
	}
	err = qb.Login()
	if err != nil {
		return nil, err
	}
	bot.qb = qb
	eb.Subscribe(bus.RSSTopic, &bot)
	eb.Subscribe(bus.QBTopic, &bot)

	// aria2
	aria2Client, err := aria2.NewClient(config.Aria2WsUrl, config.Aria2Secret, config.Aria2DownloadDir)
	if err != nil {
		return nil, err
	}
	bot.aria2 = aria2Client

	// pikpak pool
	pikpakPool, err := pikpak.NewPool(config.PikPakConfigPath)
	if err != nil {
		return nil, err
	}

	// smart downloader
	dl, err := downloader.NewSmartDownloader(aria2Client, pikpakPool, qb)
	if err != nil {
		return nil, err
	}
	bot.dl = dl
	dl.AddCallback(&bot)

	return &bot, nil
}

func (bot *AutoBangumi) Start() {
	bot.logger.Info().Msg("starting auto bangumi bot")
	bot.eb.Start()
	bot.rssMan.Start()
}

func (bot *AutoBangumi) HandleEvent(event bus.Event) {
	bot.logger.Info().Str("event type", event.EventType).Msg("recv event")
	var err error
	switch event.EventType {
	case bus.RSSUpdateEventType:
		bgm := event.Inner.(bangumi.Bangumi)
		bot.handleBangumiUpdate(&bgm)
	}
	if err != nil {
		bot.logger.Error().Err(err).Str("event type", event.EventType).Msg("handle event error")
	}
}

func (bot *AutoBangumi) handleBangumiUpdate(bangumi *bangumi.Bangumi) {
	bot.logger.Info().Msg("handle bangumi update event")
	for _, season := range bangumi.Seasons {
		for _, episode := range season.ListIncompleteEpisodes() {
			err := bot.handleEpisodeUpdate(&bangumi.Info, season.Number, episode)
			if err != nil {
				bot.logger.Error().Err(err).Msg("handle episode update error")
			} else {
				bot.logger.Info().Str("title", bangumi.Info.Title).Uint("season", season.Number).Uint("episode", episode.Number).Msg("episode update")
			}
		}
	}
}

func (bot *AutoBangumi) handleEpisodeUpdate(info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) error {
	state, err := bot.dl.DownloadEpisode(info, seasonNum, episode)
	if err != nil {
		return err
	}
	if state != nil {
		bot.bgmMan.DownloaderTouchEpisode(info, seasonNum, episode, *state)
	}
	return nil
}

func (bot *AutoBangumi) OnComplete(info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) {
	bot.bgmMan.MarkEpisodeComplete(info, seasonNum, episode)
}

func (bot *AutoBangumi) OnErr(err error, info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) {

}
