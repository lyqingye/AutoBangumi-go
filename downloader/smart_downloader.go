package downloader

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"autobangumi-go/bangumi"
	"autobangumi-go/downloader/aria2"
	"autobangumi-go/downloader/pikpak"
	"autobangumi-go/downloader/qbittorrent"
	"autobangumi-go/utils"
	torrent "github.com/anacrolix/torrent/metainfo"
	pikpakgo "github.com/lyqingye/pikpak-go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/siku2/arigo"
)

var (
	ErrNoValidResourceToDownload = errors.New("no valid resource to download")
)

type Callback interface {
	OnComplete(bgm bangumi.Bangumi, seasonNum uint, epNum uint)
	OnErr(err error, bgm bangumi.Bangumi, seasonNum uint, epNum uint)
}

type SmartDownloader struct {
	logger     zerolog.Logger
	aria2      *aria2.Client
	pikpak     *pikpak.Pool
	qb         *qbittorrent.QbittorrentClient
	dhs        DownloadService
	callbacks  []Callback
	mtx        sync.Mutex
	episodeMtx map[string]*sync.Mutex
}

type PikpakDownloaderContext struct {
	Aria2GIDS []string
}

type QbDownloaderContext struct {
	TaskId string
}

func NewSmartDownloader(aria2 *aria2.Client, pikpak *pikpak.Pool, qb *qbittorrent.QbittorrentClient, dhs DownloadService) (*SmartDownloader, error) {
	downloader := SmartDownloader{
		logger:     utils.GetLogger("smart-downloader"),
		aria2:      aria2,
		pikpak:     pikpak,
		qb:         qb,
		dhs:        dhs,
		mtx:        sync.Mutex{},
		episodeMtx: map[string]*sync.Mutex{},
	}
	if qb == nil && aria2 == nil {
		return nil, errors.New("qb and aria2 disable! no available downloader")
	}
	return &downloader, nil
}

func (dl *SmartDownloader) AddCallback(callback Callback) {
	dl.callbacks = append(dl.callbacks, callback)
}

func (dl *SmartDownloader) onComplete(bgm bangumi.Bangumi, seasonNum uint, epNum uint) {
	for _, callback := range dl.callbacks {
		callback.OnComplete(bgm, seasonNum, epNum)
	}
}

func (dl *SmartDownloader) onErr(err error, bgm bangumi.Bangumi, seasonNum uint, epNum uint) {
	for _, callback := range dl.callbacks {
		callback.OnErr(err, bgm, seasonNum, epNum)
	}
}

func (dl *SmartDownloader) attachDownloadingEpisode(l zerolog.Logger, bgm bangumi.Bangumi, season bangumi.Season, ep bangumi.Episode, resources bangumi.Resources) (bool, error) {
	history, err := dl.dhs.GetEpisodeDownloadHistory(ep)
	if err != nil {
		return false, err
	}
	if history != nil {
		switch history.GetState() {
		case bangumi.TryDownload:
			if time.Now().Sub(history.LastUpdatedTime()).Minutes() > 5 {
				goto RemoveDownloadHistory
			}
			return false, nil
		case bangumi.Downloaded:
			return true, nil
		case bangumi.Downloading:
			switch history.GetDownloader() {
			case bangumi.PikpakDownloader:
				ctx := utils.MustFromJson[PikpakDownloaderContext](history.GetDownloaderContext())
				if len(ctx.Aria2GIDS) > 0 {
					go dl.waitAria2DownloadComplete(l, ctx.Aria2GIDS, bgm, season.GetNumber(), ep.GetNumber(), history)
					return true, nil
				}
				status, err := dl.findAria2Task(bgm, season.GetNumber(), ep.GetNumber())
				if err != nil {
					return false, err
				}
				if status != nil {
					go dl.waitAria2DownloadComplete(l, []string{status.GID}, bgm, season.GetNumber(), ep.GetNumber(), history)
					return false, nil
				}
			}
		case bangumi.DownloadErr:
			// 不可恢复的错误, 说明资源不可用
			if strings.Contains(history.GetErrMsg(), pikpakgo.ErrWaitForOfflineDownloadTimeout.Error()) {
				if history.GetResourcesIds() == resources.GetId() {
					return true, nil
				}
				return false, nil
			}
		}
	}
RemoveDownloadHistory:
	return false, dl.dhs.RemoveEpisodeDownloadHistory(ep)
}

func (dl *SmartDownloader) getLock(bgm bangumi.Bangumi, season bangumi.Season, ep bangumi.Episode) *sync.Mutex {
	dl.mtx.Lock()
	defer dl.mtx.Unlock()
	key := fmt.Sprintf("%d-%d-%d", bgm.GetTmDBId(), season.GetNumber(), ep.GetNumber())
	lock, found := dl.episodeMtx[key]
	if found {
		return lock
	}
	newLock := new(sync.Mutex)
	dl.episodeMtx[key] = newLock
	return newLock
}

func (dl *SmartDownloader) DownloadEpisode(bgm bangumi.Bangumi, season bangumi.Season, ep bangumi.Episode, resources bangumi.Resources) error {
	lock := dl.getLock(bgm, season, ep)
	if lock.TryLock() {
		defer lock.Unlock()
	} else {
		return nil
	}

	l := dl.logger.With().Str("title", bgm.GetTitle()).Uint("season", season.GetNumber()).Uint("episode", ep.GetNumber()).Logger()
	attached, err := dl.attachDownloadingEpisode(l, bgm, season, ep, resources)
	if err != nil {
		return err
	}
	if attached {
		return nil
	}

	history, err := dl.dhs.AddEpisodeDownloadHistory(ep, resources.GetId())
	if err != nil {
		return err
	}

	l.Info().Msg("start download episode")
	if gids, err := dl.batchDownloadUsingPikpakAndAria2(l, bgm, season.GetNumber(), ep.GetNumber(), resources); err != nil {
		ctx := utils.MustToJson(PikpakDownloaderContext{gids})
		history.SetDownloader(bangumi.PikpakDownloader, ctx, bangumi.DownloadErr, err)
		if err := dl.dhs.UpdateDownloadHistory(history); err != nil {
			return err
		}
	} else {
		if len(gids) == 0 {
			history.SetDownloader(bangumi.PikpakDownloader, "", bangumi.DownloadErr, errors.New("empty files"))
			return dl.dhs.UpdateDownloadHistory(history)
		}
		go dl.waitAria2DownloadComplete(l, gids, bgm, season.GetNumber(), ep.GetNumber(), history)

		ctx := utils.MustToJson(PikpakDownloaderContext{gids})
		history.SetDownloader(bangumi.PikpakDownloader, ctx, bangumi.Downloading, nil)
		return dl.dhs.UpdateDownloadHistory(history)
	}
	return nil
}

func (dl *SmartDownloader) findAria2Task(bgm bangumi.Bangumi, seasonNum uint, epNum uint) (*arigo.Status, error) {
	// check episode download task already exists
	baseName := fmt.Sprintf("[%s] S%01dE%02d", bgm.GetTitle(), seasonNum, epNum)

	var target *arigo.Status
	err := dl.aria2.IterAllTasks(func(status arigo.Status) bool {
		for _, file := range status.Files {
			if strings.Contains(file.Path, baseName) {
				target = &status
				return true
			}
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return target, nil
}

func (dl *SmartDownloader) findQBTask(resource bangumi.Resource) (*qbittorrent.Torrent, error) {
	// check torrent downloading
	torrentTask, err := dl.qb.GetTorrent(resource.GetTorrentHash())
	if !errors.Is(err, qbittorrent.ErrTorrentNotFound) && err != nil {
		return nil, nil
	}
	return torrentTask, nil
}

type offlineDownloadResult struct {
	resource bangumi.Resource
	result   pikpak.Files
}

func (dl *SmartDownloader) batchDownloadUsingPikpakAndAria2(l zerolog.Logger, bgm bangumi.Bangumi, seasonNum uint, epNum uint, resources []bangumi.Resource) ([]string, error) {
	l.Info().Msg("using pikpak + aria2 download episode")
	wg := sync.WaitGroup{}
	wg.Add(len(resources))
	mtx := sync.Mutex{}
	results := make([]offlineDownloadResult, 0, len(resources))

	for _, resource := range resources {
		copyResource := resource
		go func() {
			defer wg.Done()
			torr, err := torrent.Load(bytes.NewBuffer(copyResource.GetTorrent()))
			if err != nil {
				return
			}
			torrHashBytes := torr.HashInfoBytes()
			torrInfo, err := torr.UnmarshalInfo()
			if err != nil {
				return
			}
			magnet := torr.Magnet(&torrHashBytes, &torrInfo).String()
			baseName := fmt.Sprintf("[%s] S%01dE%02d %s", bgm.GetTitle(), seasonNum, epNum, torrHashBytes.String())
			files, err := dl.pikpak.OfflineDownAndWait(baseName, magnet)
			if err != nil {
				l.Error().Err(err).Msg("pikpak download error")
				if errors.Is(err, pikpakgo.ErrWaitForOfflineDownloadTimeout) {
					if dbErr := dl.dhs.MarkResourceIsInvalid(copyResource); dbErr != nil {
						l.Error().Err(dbErr).Msg("mark resource invalid error")
					}
				}
				return
			}
			mtx.Lock()
			defer mtx.Unlock()
			results = append(results, offlineDownloadResult{
				resource: copyResource,
				result:   files,
			})
		}()
	}

	wg.Wait()
	if len(results) == 0 {
		return nil, ErrNoValidResourceToDownload
	}

	sort.SliceStable(results, func(i, j int) bool {
		return bangumi.CompareResource(results[i].resource, results[j].resource)
	})

	for i, result := range results {
		files := result.result
		if i != 0 {
			for _, fi := range files {
				_ = dl.pikpak.RemoveFile(fi.DownloadUrl)
			}
		}
	}

	return dl.downloadPikpakFilesUsingAria2(l, bgm, seasonNum, epNum, results[0].result)
}

func (dl *SmartDownloader) downloadPikpakFilesUsingAria2(l zerolog.Logger, bgm bangumi.Bangumi, seasonNum uint, epNum uint, files pikpak.Files) ([]string, error) {
	var gids []string
	clear := func() {
		for _, gid := range gids {
			_ = dl.aria2.RemoveDownloadResult(gid)
		}
	}
	dir := bangumi.DirNaming(bgm, seasonNum)
	for _, fi := range files {
		newFilename := bangumi.RenamingEpisodeFileName(bgm, seasonNum, epNum, fi.Name)
		if newFilename != "" {
			l.Debug().Str("file", newFilename).Msg("try add aria2 task")
			gid, err := dl.aria2.AddTask(fi.DownloadUrl, dir, newFilename)
			if err != nil {
				l.Error().Err(err).Str("file", newFilename).Msg("add aria2 task error, will remove all tasks")
				clear()
				return nil, err
			}
			l.Debug().Str("file", newFilename).Msg("add aria2 task success")
			gids = append(gids, gid.GID)
		} else {
			l.Warn().Str("file", fi.Name).Msg("unable to rename file")
		}
	}

	return gids, nil
}

func (dl *SmartDownloader) waitAria2DownloadComplete(l zerolog.Logger, gids []string, bgm bangumi.Bangumi, seasonNum uint, epNum uint, history bangumi.EpisodeDownLoadHistory) {
	if len(gids) == 0 {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(len(gids))

	for i, gid := range gids {
		l.Info().Str("gid", gid).Msg(fmt.Sprintf("waiting aria2 task complete %d/%d", i, len(gids)))
		dl.aria2.WaitDownloadComplete(gid, func(status arigo.Status) {
			if status.Status == arigo.StatusError {
				// remove aria2 task
				_ = dl.aria2.ForceRemove(gid)

				err := errors.New(status.ErrorMessage)
				history.SetDownloadState(bangumi.DownloadErr, err)
				if err := dl.dhs.UpdateDownloadHistory(history); err != nil {
					l.Error().Err(err).Msg("update download history")
				}

				// callback
				l.Error().Err(err).Str("gid", gid).Msg(fmt.Sprintf("waiting aria2 task complete %d/%d", i, len(gids)))
				dl.onErr(err, bgm, seasonNum, epNum)
			} else {
				l.Info().Str("gid", gid).Msg(fmt.Sprintf("aria2 task complete %d/%d", i, len(gids)))

				// recycle pikpak storage
				for _, fi := range status.Files {
					for _, uri := range fi.URIs {
						err := dl.pikpak.RemoveFile(uri.URI)
						if err == nil {
							l.Debug().Str("filename", fi.Path).Msg("remove pikpak file")
						}
					}
				}
			}

		})
		wg.Done()
	}

	go func() {
		// wait all files download complete
		l.Info().Msg("wait all files download complete")

		wg.Wait()

		l.Info().Msg("all files download complete")
		history.SetDownloadState(bangumi.Downloaded, nil)

		if err := dl.dhs.UpdateDownloadHistory(history); err != nil {
			l.Error().Err(err).Msg("update download history")
		}

		// callback
		dl.onComplete(bgm, seasonNum, epNum)
	}()
}

func (dl *SmartDownloader) downloadUsingQbittorrent(l zerolog.Logger, bgm bangumi.Bangumi, seasonNum uint, epNum uint, resource bangumi.Resource) (string, error) {
	var hash string
	var err error
	if resource.GetTorrent() != nil {
		l.Info().Msg("using qbittorrent download episode")

		// try download
		opts := qbittorrent.AddTorrentOptions{
			Paused: true,
			Rename: bangumi.RenamingEpisodeFileName(bgm, seasonNum, epNum, bgm.GetTitle()),
		}
		l.Info().Msg("start download episode")
		hash, err = dl.qb.AddTorrentEx(&opts, resource.GetTorrent(), bangumi.DirNaming(bgm, seasonNum))
		if err != nil {
			return hash, errors.Wrap(err, "add torrent")
		}
		// Wait for torrent parsing complete
		l.Info().Msg("wait for torrent parsing complete")
		for {
			torrentTask, err := dl.qb.GetTorrent(resource.GetTorrentHash())
			if err == nil {
				if torrentTask != nil && torrentTask.State == qbittorrent.StatePausedDL {
					break
				}
			}
			time.Sleep(time.Second)
		}

		// renaming torrent files
		err = dl.renameTorrent(l, hash, bgm, seasonNum, epNum)
		if err != nil {
			return hash, errors.Wrap(err, "rename torrent files error")
		}

		content, err := dl.qb.GetTorrentContent(hash, []int64{})
		if err != nil {
			return hash, errors.Wrap(err, "get torrent content")
		}

		if len(content) == 0 {
			l.Warn().Msg("the torrent not available content to download, skip them")
			_ = dl.qb.DeleteTorrents([]string{hash}, true)
		}

		// resume
		err = dl.qb.ResumeTorrents([]string{hash})
		if err != nil {
			return hash, errors.Wrap(err, "resume torrent")
		}
	} else {
		l.Warn().Msg("skip episode, torrent hash is empty")
	}

	return hash, nil
}

func (dl *SmartDownloader) waitQbDownloadComplete(l zerolog.Logger, hash string, info bangumi.Bangumi, seasonNum uint, epNum uint, resource bangumi.Resource, history bangumi.EpisodeDownLoadHistory) {
	isTimeout := false
	err := dl.qb.WatchTorrentProperties(hash, time.Second*5, func(torr *qbittorrent.TorrentProperties) bool {
		if torr.CompletionDate != -1 {
			l.Info().Msg("download complete")
			dl.onComplete(info, seasonNum, epNum)
			return true
		}

		expire := (time.Now().Unix() - int64(torr.AdditionDate)) > int64(time.Hour.Seconds())

		if expire && torr.DlSpeed == 0 && (torr.TotalDownloaded/torr.TotalSize*100) < 50 {
			isTimeout = true
			return true
		}

		return false
	})

	// 下载失败
	if err != nil {
		l.Warn().Err(err).Msg("wait for download complete error, torrent maybe remove")
		history.SetDownloadState(bangumi.DownloadErr, err)
		dl.onErr(err, info, seasonNum, epNum)

		// download download state
		if err := dl.dhs.UpdateDownloadHistory(history); err != nil {
			l.Error().Err(err).Msg("update download history")
		}
	}

	// 下载超时
	if isTimeout {
		l.Warn().Msg("torrent has not been downloaded by qb for more than an hour, try fallback to pikpak + aria2")
		err = dl.qb.DeleteTorrents([]string{hash}, true)
		if err != nil {
			l.Error().Err(err).Msg("delete qb torrent")
		}

		// reset downloader state
		history.IncRetryCount()
		history.SetDownloadState(bangumi.DownloadErr, errors.New("download timeout"))
		if err := dl.dhs.UpdateDownloadHistory(history); err != nil {
			l.Error().Err(err).Msg("update download history")
		}

		//if history.GetRetryCount() < 5 {
		//	go func() {
		//		err = dl.DownloadEpisode(info, seasonNum, epNum, resource)
		//		if err != nil {
		//			l.Error().Err(err).Msg("fallback to pikpak + aria2 error")
		//		}
		//	}()
		//}
	}
}

func (dl *SmartDownloader) renameTorrent(l zerolog.Logger, hash string, bgm bangumi.Bangumi, seasonNum uint, epNum uint) error {
	content, err := dl.qb.GetTorrentContent(hash, []int64{})
	if err != nil {
		return err
	}
	for _, fi := range content {
		newName := bangumi.RenamingEpisodeFileName(bgm, seasonNum, epNum, fi.Name)
		if newName != "" {
			err = dl.qb.RenameFile(hash, fi.Name, newName)
			if err != nil {
				return err
			}
			l.Info().Str("filename", fi.Name).Str("new filename", newName).Msg("rename episode")
		} else {
			err = dl.qb.SetFilePriority(hash, []int{fi.Index}, 0)
			if err != nil {
				l.Error().Err(err).Str("filename", fi.Name).Msg("unable to rename file, skip download file error")
			} else {
				l.Warn().Str("filename", fi.Name).Msg("unable to rename file, skip to download this file")
			}
		}
	}
	return nil
}
