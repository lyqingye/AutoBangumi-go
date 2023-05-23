package bot

import (
	"autobangumi-go/bangumi"
	"autobangumi-go/bus"
	"autobangumi-go/db"
	"autobangumi-go/downloader/qibittorrent"
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
}

func (config *AutoBangumiConfig) Validate() error {
	_, err := url.Parse(config.QBEndpoint)
	if err != nil {
		return err
	}
	if config.QBUsername == "" || config.QBPassword == "" {
		return errors.New("empty qb username or password")
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
	qb     *qibittorrent.QbittorrentClient
	db     *db.DB
	eb     *bus.EventBus
	rssMan *rss.RSSManager
	logger zerolog.Logger
	bgmMan *bangumi.BangumiManager
}

func NewAutoBangumi(config *AutoBangumiConfig) (*AutoBangumi, error) {
	bot := AutoBangumi{}
	bot.logger = utils.GetLogger("AutoBangumi")
	database, err := db.NewDB(config.DBDir)
	if err != nil {
		return nil, err
	}
	eb := bus.NewEventBus()
	bot.eb = eb
	bangumiTVClient, err := mdb.NewBangumiTVClient(config.BangumiTVApiEndpoint)
	if err != nil {
		return nil, err
	}
	tmdbClient, err := mdb.NewTMDBClient(config.TMDBToken)
	if err != nil {
		return nil, err
	}
	bgmMan, err := bangumi.NewBangumiManager(config.BangumiHome)
	if err != nil {
		return nil, err
	}
	rssMan, err := rss.NewRSSManager(bgmMan, eb, database, config.RSSUpdatePeriod, tmdbClient, bangumiTVClient)
	if err != nil {
		return nil, err
	}
	bot.db = database
	bot.rssMan = rssMan
	bot.bgmMan = bgmMan
	qb, err := qibittorrent.NewQbittorrentClient(config.QBEndpoint, config.QBUsername, config.QBPassword, config.QBDownloadDir)
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
	return &bot, nil
}

func (bot *AutoBangumi) Start() {
	bot.logger.Info().Msg("starting auto bangumi bot")
	go bot.qbAutoResumePausedTorrents()
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

func (bot *AutoBangumi) handleEpisodeUpdate(info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) error {
	if episode.TorrentHash != "" {

		// check torrent downloading
		torrentTask, err := bot.qb.GetTorrent(episode.TorrentHash)
		if err != qibittorrent.ErrTorrentNotFound && err != nil {
			return nil
		}
		if err == nil && torrentTask != nil {
			// download complete
			if torrentTask.CompletionOn != 0 {
				bot.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("download complete")
				bot.bgmMan.MarkEpisodeComplete(info, seasonNum, episode)
			}

			// downloading
			bot.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("downloading")
			bot.qb.WaitForDownloadComplete(torrentTask.Hash, time.Second*5, func() bool {
				bot.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("download complete")
				bot.bgmMan.MarkEpisodeComplete(info, seasonNum, episode)
				return true
			})
			return nil
		}

		// try download
		opts := qibittorrent.AddTorrentOptions{
			Paused: true,
			Rename: bangumi.RenamingEpisodeFileName(info, seasonNum, &episode, info.Title),
		}
		bot.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("start download episode")
		hash, err := bot.qb.AddTorrentEx(&opts, episode.Torrent, bangumi.DirNaming(info, seasonNum))
		if err != nil {
			return err
		}

		go func() {
			// Wait for torrent parsing complete
			bot.logger.Info().Msg("wait for torrent parsing complete")
			for {
				if torrentTask != nil && torrentTask.State == qibittorrent.StatePausedDL {
					break
				}
				time.Sleep(time.Second)
				torrentTask, _ = bot.qb.GetTorrent(episode.TorrentHash)
			}

			// renaming torrent files
			err = bot.renameTorrent(hash, info, seasonNum, &episode)
			if err != nil {
				bot.logger.Error().Err(err).Msg("rename torrent files error")
				return
			}

			// resume
			err = bot.qb.ResumeTorrents([]string{hash})
			if err != nil {
				bot.logger.Error().Err(err).Msg("resume torrent error")
				return
			}

			// wait for download complete
			go func() {
				bot.qb.WaitForDownloadComplete(hash, time.Second*5, func() bool {
					bot.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("download complete")
					bot.bgmMan.MarkEpisodeComplete(info, seasonNum, episode)
					return true
				})
			}()
		}()

	} else {
		bot.logger.Warn().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("skip episode, torrent hash is empty")
	}
	return nil
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

func (bot *AutoBangumi) qbAutoResumePausedTorrents() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		torrents, err := bot.qb.ListAllTorrent(qibittorrent.FilterPausedTorrentList)
		if err == nil {
			var hashes []string
			for _, torrent := range torrents {
				if torrent.State == qibittorrent.StateError {
					continue
				}
				if torrent.CompletionOn == 0 {
					hashes = append(hashes, torrent.Hash)
				}
			}
			if len(hashes) != 0 {
				err = bot.qb.ResumeTorrents(hashes)
				if err != nil {
					bot.logger.Err(err).Msg("resume paused torrent error")
				}
			}
		}
	}
}

func (bot *AutoBangumi) renameTorrent(hash string, info *bangumi.BangumiInfo, seasonNum uint, episode *bangumi.Episode) error {
	content, err := bot.qb.GetTorrentContent(hash, []int64{})
	if err != nil {
		return err
	}
	for _, fi := range content {
		newName := bangumi.RenamingEpisodeFileName(info, seasonNum, episode, fi.Name)
		if newName != "" {
			err = bot.qb.RenameFile(hash, fi.Name, newName)
			if err != nil {
				return err
			}
			bot.logger.Info().Str("hash", hash).Str("filename", fi.Name).Str("new filename", newName).Msg("rename episode")
		} else {
			bot.logger.Warn().Str("hash", hash).Str("filename", fi.Name).Msg("unable to rename file")
		}
	}
	return nil
}
