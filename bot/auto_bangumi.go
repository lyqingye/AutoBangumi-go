package bot

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"autobangumi-go/bangumi"
	"autobangumi-go/config"
	"autobangumi-go/db"
	"autobangumi-go/downloader"
	"autobangumi-go/downloader/aria2"
	"autobangumi-go/downloader/pikpak"
	"autobangumi-go/downloader/qbittorrent"
	"autobangumi-go/jellyfin"
	"autobangumi-go/mdb"
	"autobangumi-go/rss/mikan"
	"autobangumi-go/rss/mikan/cache"
	"autobangumi-go/utils"
	"github.com/nssteinbrenner/anitogo"
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
	backend        *db.Backend
	jellyfin       *jellyfin.Client
	mtx            sync.Mutex
	bgmMtx         map[int64]*sync.Mutex
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
	bot.backend = backend
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
	bot.bgmMtx = make(map[int64]*sync.Mutex)

	if config.Jellyfin.AutoScanLibraryWhenDownloadFinished {
		jellyfinClient, err := jellyfin.NewClient(config.Jellyfin.Endpoint, config.Jellyfin.Username, config.Jellyfin.Password)
		if err != nil {
			return nil, err
		}
		bot.jellyfin = jellyfinClient
	}
	bot.dl.AddCallback(&bot)

	return &bot, nil
}

func (ab *AutoBangumi) AddBangumi(title string, tmdbID int64) (bangumi.Bangumi, error) {
	bgm, err := ab.bgmMan.GetBgmByTmDBId(tmdbID)
	if err != nil {
		return nil, err
	}
	if bgm != nil {
		return nil, errors.New("bangumi already exists")
	}
	parser, err := ab.getParser(title)
	if err != nil {
		return nil, err
	}
	defer parser.Close()
	searchResult, err := parser.Search(title, tmdbID)
	if err != nil {
		return nil, err
	}
	err = ab.bgmMan.AddBangumi(searchResult)
	if err != nil {
		return nil, err
	}
	addedBgm, err := ab.bgmMan.GetBgmByTmDBId(tmdbID)
	if err != nil {
		return nil, err
	}
	if addedBgm == nil {
		return nil, errors.New("panic added bangumi not found in storage")
	}
	go func() {
		if err := ab.completeBangumi(addedBgm); err != nil {
			ab.logger.Error().Err(err).Msg("complete bangumi error")
		}
	}()
	return searchResult, nil
}

func (ab *AutoBangumi) Start() {
	ab.logger.Info().Msg("starting auto bangumi ab")
	ab.scanBangumis()
	go ab.cleanDownloadedBangumiCache()
	ab.runLoop()
}

func (ab *AutoBangumi) scanBangumis() {
	if !ab.cfg.WebDAV.ImportBangumiOnStartup {
		return
	}
	fs, err := NewWebDavFileSystem(ab.cfg.WebDAV.Host, ab.cfg.WebDAV.Username, ab.cfg.WebDAV.Password)
	if err != nil {
		ab.logger.Error().Err(err).Msg("create webdav file system error")
		return
	}
	err = ab.ScanFileSystemBangumis(fs, ab.cfg.WebDAV.Dir)
	if err != nil {
		ab.logger.Error().Err(err).Msg("scan bangumi error")
	}
}

func (ab *AutoBangumi) runLoop() {
	if err := ab.tick(); err != nil {
		ab.logger.Error().Err(err).Msg("tick err")
	}
	for range ab.ticker.C {
		if err := ab.tick(); err != nil {
			ab.logger.Error().Err(err).Msg("tick err")
		}
	}
}

func (ab *AutoBangumi) AddPikpakAccount(username, password string) error {
	return ab.accountStorage.AddAccount(pikpak.Account{
		Username:       username,
		Password:       password,
		State:          pikpak.StateNormal,
		RestrictedTime: 0,
	})
}

func (ab *AutoBangumi) cleanDownloadedBangumiCache() {
	for range time.NewTicker(ab.cfg.Cache.ClearCacheInterval).C {
		bangumis, err := ab.bgmMan.ListDownloadedBangumis(nil)
		if err != nil {
			ab.logger.Error().Err(err).Msg("list downloaded bangumis err")
			continue
		}
		for _, bgm := range bangumis {
			ab.logger.Info().Msgf("cleaning downloaded bangumi cache, bangumi: %s", bgm.GetTitle())
			cacheDBHome := filepath.Join(ab.cfg.Cache.CacheDir, bgm.GetTitle())
			_ = os.RemoveAll(cacheDBHome)
		}
	}
}

func (ab *AutoBangumi) tick() error {
	bangumis, err := ab.bgmMan.ListUnDownloadedBangumis()
	if err != nil || len(bangumis) == 0 {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(len(bangumis))
	for _, bgm := range bangumis {
		logger := ab.logger.With().Str("bgm", bgm.GetTitle()).Logger()
		copyBgm := bgm
		go func() {
			defer wg.Done()

			parser, err := ab.getParser(copyBgm.GetTitle())
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
			err = ab.bgmMan.AddBangumi(updatedBgm)
			if err != nil {
				logger.Error().Err(err).Msg("insert bangumi to storage error")
			} else {
				if err := ab.completeBangumi(copyBgm); err != nil {
					ab.logger.Error().Err(err).Msg("complete bangumi error")
				}
			}
		}()
	}
	wg.Wait()
	return nil
}

func (ab *AutoBangumi) getBgmLock(bgm bangumi.Bangumi) *sync.Mutex {
	ab.mtx.Lock()
	defer ab.mtx.Unlock()
	lock, found := ab.bgmMtx[bgm.GetTmDBId()]
	if found {
		return lock
	}
	newLock := new(sync.Mutex)
	ab.bgmMtx[bgm.GetTmDBId()] = newLock
	return newLock
}

func (ab *AutoBangumi) completeBangumi(bgm bangumi.Bangumi) error {
	lock := ab.getBgmLock(bgm)
	if lock.TryLock() {
		defer lock.Unlock()
	} else {
		return nil
	}
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
			copyEp := ep
			go func() {
				if err := ab.downloadSingleEpisode(ab.logger, bgm, season, copyEp); err != nil {
					ab.logger.Error().Err(err).Msg("download episode error")
				}
			}()
		}
	}
	return nil
}

func (ab *AutoBangumi) downloadSingleEpisode(l zerolog.Logger, bgm bangumi.Bangumi, season bangumi.Season, ep bangumi.Episode) error {
	if ep.IsDownloaded() {
		return nil
	}
	resources, err := ab.bgmMan.GetValidEpisodeResources(ep)
	if err != nil {
		return err
	}
	if len(resources) == 0 {
		l.Warn().Msg("episode no available resources to download")
		return nil
	}
	return ab.dl.DownloadEpisode(bgm, season, ep, resources)
}

func (ab *AutoBangumi) getParser(cacheKey string) (*mikan.MikanRSSParser, error) {
	cacheDBHome := filepath.Join(ab.cfg.Cache.CacheDir, cacheKey)
	cacheDB, err := db.NewDB(cacheDBHome)
	if err != nil {
		return nil, err
	}
	cm := cache.NewKVCacheManager(cacheDB)
	parser, err := mikan.NewMikanRSSParser(ab.cfg.Mikan.Endpoint, ab.tmdb, ab.bgmTV, cm)
	if err != nil {
		_ = cm.Close()
		return nil, err
	}
	return parser, err
}

func (ab *AutoBangumi) ScanFileSystemBangumis(fs FileSystem, bangumisDir string) error {
	files, err := fs.ReadDir(bangumisDir)
	if err != nil {
		return err
	}
	for _, fi := range files {
		if !fi.IsDir() {
			continue
		}
		err = ab.ScanFileSystemBangumi(fs, fi.Name(), filepath.Join(bangumisDir, fi.Name()))
		if err != nil {
			continue
		}
	}
	return nil
}

func (ab *AutoBangumi) ScanFileSystemBangumi(fs FileSystem, bangumiName string, path string) error {
	re := regexp.MustCompile(`\(\d{4}\)`)
	bangumiName = re.ReplaceAllString(bangumiName, "")

	// 搜索tmdb
	detail, err := ab.tmdb.SearchTVShowByKeyword(bangumiName)
	if err != nil {
		return err
	}
	// 确保是动漫
	isAnime := false
	for _, genres := range detail.Genres {
		if genres.ID == 16 {
			isAnime = true
			break
		}
	}
	if !isAnime {
		return nil
	}

	// 缓存Season元数据
	seasonEpCount := make(map[uint]uint)
	for _, season := range detail.Seasons {
		if season.EpisodeCount == 0 {
			continue
		}
		seasonEpCount[uint(season.SeasonNumber)] = uint(season.EpisodeCount)
	}

	bgmFromFs := NewBangumiFromFs(detail.Name, detail.ID)
	err = fs.WalkDir(path, func(seasonFileName string, info os.FileInfo) (bool, error) {
		episodeName := info.Name()
		seasonNum, epNum, err := bangumi.ParseEpisodeFilename(episodeName)
		if err == nil && seasonNum > 0 && epNum > 0 {
			if epCount, found := seasonEpCount[seasonNum]; found {
				bgmFromFs.AddDownloadEpisode(seasonNum, epCount, epNum)
			}
		} else {
			seasonNum, err = bangumi.ParseSeasonFilename(seasonFileName)
			if err == nil {
				parsedElements := anitogo.Parse(episodeName, anitogo.DefaultOptions)
				if len(parsedElements.EpisodeNumber) > 0 {
					parsedEpNum, err := strconv.ParseUint(parsedElements.EpisodeNumber[0], 10, 32)
					if err == nil && seasonNum > 0 && parsedEpNum > 0 {
						if epCount, found := seasonEpCount[seasonNum]; found {
							bgmFromFs.AddDownloadEpisode(seasonNum, epCount, uint(parsedEpNum))
						}
					}
				}
			}
		}
		return false, nil
	})

	if err != nil {
		return err
	}

	if bgmFromFs.IsDownloaded() && len(bgmFromFs.seasons) == len(seasonEpCount) {
		return nil
	}
	return ab.backend.ImportDownloadBangumi(nil, &bgmFromFs)
}

func (ab *AutoBangumi) OnComplete(bgm bangumi.Bangumi, seasonNum uint, epNum uint) {
	if ab.cfg.Jellyfin.AutoScanLibraryWhenDownloadFinished && ab.jellyfin != nil {
		err := ab.jellyfin.StartLibraryScan()
		if err != nil {
			ab.logger.Error().Err(err).Msg("start scan jellyfin library error")
		} else {
			ab.logger.Info().Msg("episode download complete! start scan jellyfin library")
		}
	}
}

func (bot *AutoBangumi) OnErr(err error, bgm bangumi.Bangumi, seasonNum uint, epNum uint) {
}
