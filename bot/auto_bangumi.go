package bot

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"autobangumi-go/bangumi"
	"autobangumi-go/config"
	"autobangumi-go/db"
	"autobangumi-go/downloader"
	"autobangumi-go/downloader/aria2"
	"autobangumi-go/downloader/pikpak"
	"autobangumi-go/downloader/qbittorrent"
	"autobangumi-go/mdb"
	"autobangumi-go/rss/mikan"
	"autobangumi-go/rss/mikan/cache"
	"autobangumi-go/utils"
	pikpakgo "github.com/lyqingye/pikpak-go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type AutoBangumi struct {
	qb             *qbittorrent.QbittorrentClient
	logger         zerolog.Logger
	bgmMan         *bangumi.Manager
	bgmTV          *mdb.BangumiTVClient
	tmdb           *mdb.TMDBClient
	dl             *downloader.SmartDownloader
	aria2          *aria2.Client
	ticker         *time.Ticker
	cfg            *config.Config
	accountStorage pikpak.AccountStorage
	mtx            sync.Mutex
}

func NewAutoBangumi(config *config.Config) (*AutoBangumi, error) {
	bot := AutoBangumi{}
	bot.logger = utils.GetLogger("AutoBangumi")

	// bangumi TV
	bangumiTVClient, err := mdb.NewBangumiTVClient(config.BangumiTV.Endpoint)
	if err != nil {
		return nil, err
	}
	bot.bgmTV = bangumiTVClient

	// TMDB
	tmdbClient, err := mdb.NewTMDBClient(config.TMDB.Token)
	if err != nil {
		return nil, err
	}
	bot.tmdb = tmdbClient

	backend, err := db.NewBackend(config.DB)
	if err != nil {
		return nil, err
	}
	// bangumi manager
	bot.bgmMan = bangumi.NewManager(backend)

	// qb
	if config.QB.Enable {
		qb, err := qbittorrent.NewQbittorrentClient(config.QB.Endpoint, config.QB.Username, config.QB.Password, config.QB.DownloadDir)
		if err != nil {
			return nil, err
		}
		err = qb.Login()
		if err != nil {
			return nil, err
		}
		bot.qb = qb
	}

	// aria2
	aria2Client, err := aria2.NewClient(config.Aria2.WsUrl, config.Aria2.Secret, config.Aria2.DownloadDir)
	if err != nil {
		return nil, err
	}
	bot.aria2 = aria2Client

	// pikpak pool
	pikpakPool, err := pikpak.NewPool(backend, config.Pikpak)
	if err != nil {
		return nil, err
	}

	// smart downloader
	dl, err := downloader.NewSmartDownloader(aria2Client, pikpakPool, bot.qb, bot.bgmMan)
	if err != nil {
		return nil, err
	}
	bot.dl = dl

	bot.ticker = time.NewTicker(config.AutoBangumi.BangumiCompleteInterval)

	bot.cfg = config
	bot.accountStorage = backend
	bot.mtx = sync.Mutex{}

	return &bot, nil
}

func (bot *AutoBangumi) AddBangumi(title string, tmdbID int64) (bangumi.Bangumi, error) {
	bgm, err := bot.bgmMan.GetBgmByTmDBId(tmdbID)
	if err != nil {
		return nil, err
	}
	if bgm != nil {
		return nil, errors.New("bangumi already exists")
	}
	parser, err := bot.getParser(title)
	if err != nil {
		return nil, err
	}
	defer parser.Close()
	searchResult, err := parser.Search(title, tmdbID)
	if err != nil {
		return nil, err
	}
	err = bot.bgmMan.AddBangumi(searchResult)
	if err != nil {
		return nil, err
	}
	addedBgm, err := bot.bgmMan.GetBgmByTmDBId(tmdbID)
	if err != nil {
		return nil, err
	}
	if addedBgm == nil {
		return nil, errors.New("panic added bangumi not found in storage")
	}
	go func() {
		if err := bot.completeBangumi(addedBgm); err != nil {
			bot.logger.Error().Err(err).Msg("complete bangumi error")
		}
	}()
	return searchResult, nil
}

func (bot *AutoBangumi) Start() {
	bot.logger.Info().Msg("starting auto bangumi bot")
	go bot.cleanDownloadedBangumiCache()
	bot.runLoop()
}

func (bot *AutoBangumi) runLoop() {
	if err := bot.tick(); err != nil {
		bot.logger.Error().Err(err).Msg("tick err")
	}
	for range bot.ticker.C {
		if err := bot.tick(); err != nil {
			bot.logger.Error().Err(err).Msg("tick err")
		}
	}
}

func (bot *AutoBangumi) AddPikpakAccount(username, password string) error {
	return bot.accountStorage.AddAccount(pikpak.Account{
		Username:       username,
		Password:       password,
		State:          pikpak.StateNormal,
		RestrictedTime: 0,
	})
}

func (bot *AutoBangumi) cleanDownloadedBangumiCache() {
	for range time.NewTicker(bot.cfg.Cache.ClearCacheInterval).C {
		bangumis, err := bot.bgmMan.ListDownloadedBangumis(nil)
		if err != nil {
			bot.logger.Error().Err(err).Msg("list downloaded bangumis err")
			continue
		}
		for _, bgm := range bangumis {
			bot.logger.Info().Msgf("cleaning downloaded bangumi cache, bangumi: %s", bgm.GetTitle())
			cacheDBHome := filepath.Join(bot.cfg.Cache.CacheDir, bgm.GetTitle())
			_ = os.RemoveAll(cacheDBHome)
		}
	}
}

func (bot *AutoBangumi) tick() error {
	bangumis, err := bot.bgmMan.ListUnDownloadedBangumis()
	if err != nil || len(bangumis) == 0 {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(len(bangumis))
	for _, bgm := range bangumis {
		logger := bot.logger.With().Str("bgm", bgm.GetTitle()).Logger()
		copyBgm := bgm
		go func() {
			defer wg.Done()

			parser, err := bot.getParser(copyBgm.GetTitle())
			if err != nil {
				logger.Error().Err(err).Msg("new mikan rss parser error")
				return
			}
			defer parser.Close()
			updatedBgm, err := parser.Search(copyBgm.GetTitle(), copyBgm.GetTmDBId())
			if err != nil {
				logger.Error().Err(err).Msg("update bangumi error")
				return
			}
			err = bot.bgmMan.AddBangumi(updatedBgm)
			if err != nil {
				logger.Error().Err(err).Msg("insert bangumi to storage error")
			}
		}()
	}
	wg.Wait()

	for _, bgm := range bangumis {
		if err := bot.completeBangumi(bgm); err != nil {
			bot.logger.Error().Err(err).Msg("complete bangumi error")
		}
	}
	return nil
}

func (bot *AutoBangumi) completeBangumi(bgm bangumi.Bangumi) error {
	bot.mtx.Lock()
	defer bot.mtx.Unlock()
	seasons, err := bgm.GetSeasons()
	if err != nil {
		return err
	}
	for _, season := range seasons {
		if season.IsDownloaded() {
			continue
		}
		episodes, err := season.GetEpisodes()
		if err != nil {
			return err
		}
		for _, ep := range episodes {
			if ep.IsDownloaded() {
				continue
			}
			dlHistories, err := bot.bgmMan.GetEpisodeResourceDownloadHistories(nil, ep)
			if err != nil {
				return err
			}
			needDownload := true
			var downloadingResource bangumi.Resource
			for _, history := range dlHistories {
				switch history.GetState() {
				case bangumi.TryDownload, bangumi.Downloading, bangumi.Downloaded:
					downloadingResource, err = bot.bgmMan.GetResource(nil, history.GetTorrentHash())
					if err != nil {
						return err
					}
					needDownload = false
					break
				case bangumi.DownloadErr:
					needDownload = false
					// 如果是超时，那就没必要重试了。。。
					if !strings.Contains(history.GetErrMsg(), pikpakgo.ErrWaitForOfflineDownloadTimeout.Error()) {
						// 可以选择重试
						downloadingResource, err = bot.bgmMan.GetResource(nil, history.GetTorrentHash())
						if err != nil {
							return err
						}
					}
					break
				}

				if !needDownload {
					break
				}
			}
			if needDownload {
				resources, err := bot.bgmMan.GetUnDownloadedEpisodeResources(ep)
				if err != nil {
					return err
				}

				if len(resources) == 0 {
					continue
				}

				resourceToDownload := bangumi.SelectBestResource(resources)
				if resourceToDownload == nil {
					continue
				}
				err = bot.dl.DownloadEpisode(bgm, season.GetNumber(), ep.GetNumber(), resourceToDownload)
				if err != nil {
					bot.logger.Error().Err(err).Msg("download episode")
				}
			} else if downloadingResource != nil {
				err = bot.dl.DownloadEpisode(bgm, season.GetNumber(), ep.GetNumber(), downloadingResource)
				if err != nil {
					bot.logger.Error().Err(err).Msg("attach downloading episode")
				}
			}
		}
	}
	return nil
}

func (bot *AutoBangumi) getParser(cacheKey string) (*mikan.MikanRSSParser, error) {
	cacheDBHome := filepath.Join(bot.cfg.Cache.CacheDir, cacheKey)
	cacheDB, err := db.NewDB(cacheDBHome)
	if err != nil {
		return nil, err
	}
	cm := cache.NewKVCacheManager(cacheDB)
	parser, err := mikan.NewMikanRSSParser(bot.cfg.Mikan.Endpoint, bot.tmdb, bot.bgmTV, cm)
	if err != nil {
		_ = cm.Close()
		return nil, err
	}
	return parser, err
}
