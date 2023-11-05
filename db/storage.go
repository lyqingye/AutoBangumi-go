package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"autobangumi-go/bangumi"
	"autobangumi-go/config"
	"autobangumi-go/downloader/pikpak"
	"autobangumi-go/utils"
	"golang.org/x/exp/maps"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Backend struct {
	db  *gorm.DB
	mtx sync.Mutex
}

const ContextTxKey = "tx"

func NewBackend(cfg config.DBConfig) (*Backend, error) {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", 0), logger.Config{LogLevel: logger.Info}),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxConns)

	backend := Backend{
		db:  db,
		mtx: sync.Mutex{},
	}
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&MBangumi{}, &MSeason{}, &MEpisode{}, &MEpisodeTorrent{}, &MEpisodeDownloadHistory{}, &MAccount{})

	return &backend, err
}

func (b *Backend) Begin() (context.Context, error) {
	tx := b.db.Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	return b.wrapCtx(tx), nil
}

func (b *Backend) Rollback(ctx context.Context) error {
	tx := ctx.Value(ContextTxKey).(*gorm.DB)
	return tx.Rollback().Error
}

func (b *Backend) Commit(ctx context.Context) error {
	tx := ctx.Value(ContextTxKey).(*gorm.DB)
	return tx.Commit().Error
}

func (b *Backend) unwrapCtx(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return b.db
	}
	v := ctx.Value(ContextTxKey)
	if v == nil {
		return b.db
	}
	return v.(*gorm.DB)
}

func (b *Backend) wrapCtx(tx *gorm.DB) context.Context {
	return context.WithValue(context.Background(), ContextTxKey, tx)
}

func (b *Backend) AddBangumi(ctx context.Context, newOrUpdate bangumi.Bangumi) error {
	ptx := b.unwrapCtx(ctx)
	tx := ptx.Begin()
	if err := tx.Error; err != nil {
		return err
	}
	if _, err := b.addBgm(tx, newOrUpdate); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (b *Backend) ImportDownloadBangumi(ctx context.Context, bgm bangumi.Bangumi) error {
	ptx := b.unwrapCtx(ctx)
	tx := ptx.Begin()
	if err := tx.Error; err != nil {
		return err
	}
	if err := b.importDownloadBangumi(tx, bgm); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (b *Backend) importDownloadBangumi(tx *gorm.DB, bgm bangumi.Bangumi) error {
	addedBgm, err := b.addBgm(tx, bgm)
	if err != nil {
		return err
	}

	addSeasons, err := addedBgm.GetSeasons()
	if err != nil {
		return err
	}
	seasonIdMap := make(map[uint]uint)
	for _, season := range addSeasons {
		v := season.(*ProxyMSeason)
		seasonIdMap[v.Number] = v.ID
	}

	if err := tx.Model(&MBangumi{}).
		Where("title = ? or tmdb_id = ?", bgm.GetTitle(), bgm.GetTmDBId()).
		UpdateColumn("downloaded", bgm.IsDownloaded()).
		Error; err != nil {
		return err
	}
	ctx := b.wrapCtx(tx)
	if err = b.MarkBangumiDownloaded(ctx, addedBgm, bgm.IsDownloaded()); err != nil {
		return err
	}

	seasons, err := bgm.GetSeasons()
	if err != nil {
		return err
	}
	for _, season := range seasons {
		seasonId, found := seasonIdMap[season.GetNumber()]
		if !found {
			return errors.New("panic season id not found")
		}
		if season.IsDownloaded() {
			if err = tx.Model(&MSeason{}).Where("id = ?", seasonId).
				UpdateColumn("downloaded", true).
				Error; err != nil {
				return err
			}
		}

		episodes, err := season.GetEpisodes()
		if err != nil {
			return err
		}

		for _, ep := range episodes {
			if !ep.IsDownloaded() {
				continue
			}
			if err = tx.Model(&MEpisode{}).Where("season_id = ? AND number = ?", seasonId, ep.GetNumber()).
				UpdateColumn("downloaded", true).
				Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Backend) addBgm(tx *gorm.DB, bgm bangumi.Bangumi) (bangumi.Bangumi, error) {
	var err error

	// 插入bangumi
	newBgm := MBangumi{
		Title:  bgm.GetTitle(),
		TMDBId: bgm.GetTmDBId(),
	}
	if err := tx.Where("title = ? or tmdb_id = ?", bgm.GetTitle(), bgm.GetTmDBId()).FirstOrCreate(&newBgm).Error; err != nil {
		return nil, err
	}

	// 插入Seasons
	var seasons []bangumi.Season
	seasons, err = bgm.GetSeasons()
	if err != nil {
		goto Exit
	}
	for _, season := range seasons {
		if err = b.addSeason(tx, newBgm.ID, season); err != nil {
			break
		}
	}
Exit:
	return Proxy(newBgm, tx), err
}

func (b *Backend) addSeason(tx *gorm.DB, bangumiId uint, season bangumi.Season) error {
	newSeason := MSeason{
		BangumiId: bangumiId,
		Number:    season.GetNumber(),
		EpCount:   season.GetEpCount(),
	}
	if err := tx.Where("bangumi_id = ? and number = ?", bangumiId, season.GetNumber()).
		FirstOrCreate(&newSeason).Error; err != nil {
		return err
	}

	// 插入每一集
	var episodes []bangumi.Episode
	episodes, err := season.GetEpisodes()
	if err != nil {
		return err
	}
	if len(episodes) == 0 {
		return nil
	}
	for _, ep := range episodes {
		if err = b.addEpisode(tx, newSeason.ID, ep); err != nil {
			return err
		}
	}
	return nil
}

func (b *Backend) addEpisode(tx *gorm.DB, seasonId uint, episode bangumi.Episode) error {
	newEp := MEpisode{
		SeasonId: seasonId,
		Number:   episode.GetNumber(),
	}
	if err := tx.Where("season_id = ? and number = ?", seasonId, episode.GetNumber()).
		FirstOrCreate(&newEp).Error; err != nil {
		return err
	}
	if newEp.Downloaded {
		return nil
	}
	if resourcesFromEp, err := episode.GetResources(); err != nil {
		return err
	} else {
		// 插入种子
		if err = b.addResources(tx, newEp.ID, resourcesFromEp); err != nil {
			return err
		}
	}
	return nil
}

func (b *Backend) addResources(tx *gorm.DB, episodeId uint, resources []bangumi.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	var hashesToInsert = make(map[string]bangumi.Resource)
	for _, res := range resources {
		hashesToInsert[res.GetTorrentHash()] = res
	}

	var torrents []MEpisodeTorrent
	if err := tx.Select([]string{"torrent_hash"}).Where("torrent_hash IN ?", maps.Keys(hashesToInsert)).Find(&torrents).Error; err != nil {
		return err
	}

	var torrentToInsert []MEpisodeTorrent
	var existsTorrents []string
	for _, existsRes := range torrents {
		existsTorrents = append(existsTorrents, existsRes.GetTorrentHash())
	}

	for _, hash := range utils.Difference(existsTorrents, maps.Keys(hashesToInsert)) {
		if newResource, found := hashesToInsert[hash]; found {
			newTorrent := MEpisodeTorrent{
				EpisodeId:    episodeId,
				TorrentHash:  newResource.GetTorrentHash(),
				Bz:           newResource.GetTorrent(),
				Resolution:   newResource.GetResolution(),
				ResourceType: newResource.GetResourceType(),
				Valid:        true,
			}
			newTorrent.SetSubtitleLang(newResource.GetSubtitleLang())
			torrentToInsert = append(torrentToInsert, newTorrent)
		}
	}
	return tx.Model(&MEpisodeTorrent{}).CreateInBatches(&torrentToInsert, 100).Error
}

func (b *Backend) GetBgmByTitle(ctx context.Context, title string) (bangumi.Bangumi, error) {
	return b.getBgmByTitle(b.unwrapCtx(ctx), title)
}

func (b *Backend) GetBgmByTmDBId(ctx context.Context, tmdbId int64) (bangumi.Bangumi, error) {
	return b.getBgmByTmDBId(b.unwrapCtx(ctx), tmdbId)
}

func (b *Backend) ListBangumis(ctx context.Context, fn func(bgm bangumi.Bangumi) bool) error {
	return b.listBangumis(b.unwrapCtx(ctx), fn)
}

func (b *Backend) ListUnDownloadedBangumis(ctx context.Context) ([]bangumi.Bangumi, error) {
	db := b.unwrapCtx(ctx)
	var bangumis []MBangumi
	if err := db.Where("downloaded", false).Find(&bangumis).Error; err != nil {
		return nil, err
	}
	var ret []bangumi.Bangumi
	for _, bgm := range bangumis {
		proxy := Proxy(bgm, b.db)
		ret = append(ret, proxy)
	}
	return ret, nil
}

func (b *Backend) ListDownloadedBangumis(ctx context.Context) ([]bangumi.Bangumi, error) {
	db := b.unwrapCtx(ctx)
	var bangumis []MBangumi
	if err := db.Where("downloaded", true).Find(&bangumis).Error; err != nil {
		return nil, err
	}
	var ret []bangumi.Bangumi
	for _, bgm := range bangumis {
		proxy := Proxy(bgm, b.db)
		ret = append(ret, proxy)
	}
	return ret, nil
}

func (b *Backend) AddEpisodeDownloadHistory(ctx context.Context, episode bangumi.Episode, resourcesId string) (bangumi.EpisodeDownLoadHistory, error) {
	actual := episode.(*ProxyEpisode)
	history := MEpisodeDownloadHistory{
		ResourcesIds: resourcesId,
		EpisodeId:    actual.ID,
		State:        string(bangumi.TryDownload),
	}
	if err := b.unwrapCtx(ctx).Where("episode_id = ?", actual.ID).
		FirstOrCreate(&history).Error; err != nil {
		return nil, err
	}
	return &ProxyEpisodeDownloadHistory{
		MEpisodeDownloadHistory: history,
		db:                      b.db,
	}, nil
}

func (b *Backend) MarkResourceIsInvalid(ctx context.Context, resource bangumi.Resource) error {
	return b.unwrapCtx(ctx).Where("torrent_hash = ?", resource.GetTorrentHash()).
		UpdateColumn("valid", false).Error
}

func (b *Backend) GetEpisodeDownloadHistory(ctx context.Context, episode bangumi.Episode) (bangumi.EpisodeDownLoadHistory, error) {
	actual := episode.(*ProxyEpisode)
	var ret MEpisodeDownloadHistory
	tx := b.unwrapCtx(ctx)
	err := tx.Where("episode_id = ?", actual.ID).First(&ret).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ProxyEpisodeDownloadHistory{
		MEpisodeDownloadHistory: ret,
		db:                      tx,
	}, nil
}

func (b *Backend) RemoveEpisodeDownloadHistory(ctx context.Context, episode bangumi.Episode) error {
	actual := episode.(*ProxyEpisode)
	return b.unwrapCtx(ctx).Delete(MEpisodeDownloadHistory{}, "episode_id = ?", actual.ID).Error
}

func (b *Backend) UpdateDownloadHistory(ctx context.Context, history bangumi.EpisodeDownLoadHistory) error {
	actual := history.(*ProxyEpisodeDownloadHistory)
	return b.unwrapCtx(ctx).Where("episode_id = ?", actual.EpisodeId).UpdateColumns(&actual.MEpisodeDownloadHistory).Error
}

func (b *Backend) MarkEpisodeDownloaded(ctx context.Context, episode bangumi.Episode) error {
	actual := episode.(*ProxyEpisode)
	return b.unwrapCtx(ctx).Model(&MEpisode{}).Where("id = ?", actual.ID).UpdateColumn("downloaded", true).Error
}

func (b *Backend) MarkSeasonDownloaded(ctx context.Context, season bangumi.Season, download bool) error {
	actual := season.(*ProxyMSeason)
	return b.unwrapCtx(ctx).Model(&MSeason{}).Where("id = ?", actual.ID).UpdateColumn("downloaded", download).Error
}

func (b *Backend) MarkBangumiDownloaded(ctx context.Context, bangumi bangumi.Bangumi, download bool) error {
	actual := bangumi.(*ProxyMBangumi)
	return b.unwrapCtx(ctx).Model(&MBangumi{}).Where("id = ?", actual.ID).UpdateColumn("downloaded", download).Error
}

func (b *Backend) GetValidEpisodeResources(ctx context.Context, episode bangumi.Episode) ([]bangumi.Resource, error) {
	resources, err := episode.GetResources()
	if err != nil {
		return nil, err
	}
	if len(resources) == 0 {
		return nil, nil
	}
	var ret []bangumi.Resource
	for _, resource := range resources {
		actual := resource.(*ProxyResource)
		if actual.Valid {
			ret = append(ret, resource)
		}
	}
	return ret, nil
}

func (b *Backend) getBgmByTitle(tx *gorm.DB, title string) (bangumi.Bangumi, error) {
	var bgm MBangumi
	retErr := tx.Where("title", title).First(&bgm).Error
	if retErr != nil {
		if errors.Is(retErr, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, retErr
	}
	return Proxy(bgm, b.db), nil
}

func (b *Backend) getBgmByTmDBId(tx *gorm.DB, tmdbId int64) (bangumi.Bangumi, error) {
	var bgm MBangumi
	retErr := tx.Where("tmdb_id", tmdbId).First(&bgm).Error
	if retErr != nil {
		if errors.Is(retErr, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, retErr
	}
	return Proxy(bgm, b.db), nil
}

func (b *Backend) listBangumis(tx *gorm.DB, fn func(bgm bangumi.Bangumi) bool) error {
	var bangumis []MBangumi
	if err := tx.Find(&bangumis).Error; err != nil {
		return err
	}
	for _, bgm := range bangumis {
		proxy := Proxy(bgm, b.db)
		if fn(proxy) {
			break
		}
	}
	return nil
}

func (b *Backend) ListAccounts() ([]pikpak.Account, error) {
	var accounts []MAccount
	err := b.db.Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	return mAccsToAccs(accounts), nil
}

func (b *Backend) ListAccountsByState(state string) ([]pikpak.Account, error) {
	var accounts []MAccount
	err := b.db.Where("state = ?", state).Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	return mAccsToAccs(accounts), nil
}

func (b *Backend) UpdateAccount(acc pikpak.Account) error {
	value := accToMAcc(acc)
	return b.db.Where("username = ?", acc.Username).UpdateColumns(&value).Error
}

func (b *Backend) AddAccount(acc pikpak.Account) error {
	value := accToMAcc(acc)
	return b.db.Where("username = ?", acc.Username).FirstOrCreate(&value).Error
}

func (b *Backend) GetAccount(username string) (pikpak.Account, error) {
	var ret MAccount
	err := b.db.Where("username = ?", username).Take(&ret).Error
	if err != nil {
		return pikpak.Account{}, err
	}
	return mAccToAcc(ret), nil
}

func mAccsToAccs(fromAccs []MAccount) []pikpak.Account {
	var ret []pikpak.Account
	for _, acc := range fromAccs {
		ret = append(ret, mAccToAcc(acc))
	}
	return ret
}

func mAccToAcc(fromAcc MAccount) pikpak.Account {
	return pikpak.Account{
		Username:       fromAcc.Username,
		Password:       fromAcc.Password,
		State:          fromAcc.State,
		RestrictedTime: fromAcc.RestrictedTime,
	}
}

func accToMAcc(acc pikpak.Account) MAccount {
	return MAccount{
		Username:       acc.Username,
		Password:       acc.Password,
		State:          acc.State,
		RestrictedTime: acc.RestrictedTime,
	}
}
