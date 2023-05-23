package downloader

import (
	"autobangumi-go/bangumi"
	"autobangumi-go/downloader/aria2"
	"autobangumi-go/downloader/pikpak"
	"autobangumi-go/downloader/qibittorrent"
	"autobangumi-go/utils"
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	torrent "github.com/anacrolix/torrent/metainfo"
	"github.com/rs/zerolog"
	"github.com/siku2/arigo"
)

type Callback interface {
	OnComplete(info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode)
	OnErr(err error, info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode)
}

type SmartDownloader struct {
	logger    zerolog.Logger
	aria2     *aria2.Client
	pikpak    *pikpak.Pool
	qb        *qibittorrent.QbittorrentClient
	callbacks []Callback
}

func NewSmartDownloader(aria2 *aria2.Client, pikpak *pikpak.Pool, qb *qibittorrent.QbittorrentClient) (*SmartDownloader, error) {
	downloader := SmartDownloader{
		logger: utils.GetLogger("smart-downloader"),
		aria2:  aria2,
		pikpak: pikpak,
		qb:     qb,
	}
	return &downloader, nil
}

func (dl *SmartDownloader) AddCallback(callback Callback) {
	dl.callbacks = append(dl.callbacks, callback)
}

func (dl *SmartDownloader) onComplete(info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) {
	for _, callback := range dl.callbacks {
		callback.OnComplete(info, seasonNum, episode)
	}
}

func (dl *SmartDownloader) onErr(err error, info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) {
	for _, callback := range dl.callbacks {
		callback.OnErr(err, info, seasonNum, episode)
	}
}

func (dl *SmartDownloader) DownloadEpisode(info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) (*bangumi.DownloadState, error) {
	if !episode.IsNeedToDownload() {
		// maybe downloading
		switch episode.DownloadState.Downloader {
		case "aria2":
			if episode.DownloadState.TaskId != "" {
				gids := strings.Split(episode.DownloadState.TaskId, ",")
				go dl.waitAria2DownloadComplete(gids, info, seasonNum, episode)
				return nil, nil
			}
			status, err := dl.findAria2Task(info, seasonNum, episode)
			if err != nil {
				return nil, err
			}
			if status != nil {
				go dl.waitAria2DownloadComplete([]string{status.GID}, info, seasonNum, episode)
				return nil, nil
			}
		case "qb":
			if episode.DownloadState.TaskId != "" {
				dl.waitQbDownloadComplete(episode.DownloadState.TaskId, info, seasonNum, episode)
				return nil, nil
			}
			torr, err := dl.findQBTask(episode)
			if err != nil {
				return nil, err
			}
			dl.waitQbDownloadComplete(torr.Hash, info, seasonNum, episode)
			return nil, nil
		}
	}

	dl.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("start download episode")
	state, err := dl.downloadUsingPikpakAndAria2(info, seasonNum, episode)
	if err != nil {
		dl.logger.Warn().Err(err).Msg("using pikpak download error, fallback to qibittorrent")
		return dl.downloadUsingQibittorrent(info, seasonNum, episode)
	}
	return state, nil
}

func (dl *SmartDownloader) findAria2Task(info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) (*arigo.Status, error) {
	// check episode download task already exists
	baseName := fmt.Sprintf("[%s] S%01dE%02d", info.Title, seasonNum, episode.Number)

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
		return target, err
	}
	return nil, nil
}

func (dl *SmartDownloader) findQBTask(episode bangumi.Episode) (*qibittorrent.Torrent, error) {
	// check torrent downloading
	torrentTask, err := dl.qb.GetTorrent(episode.TorrentHash)
	if err != qibittorrent.ErrTorrentNotFound && err != nil {
		return nil, nil
	}
	return torrentTask, nil
}

// DownloadMagnetAndWait download and wait download complete
func (dl *SmartDownloader) downloadUsingPikpakAndAria2(info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) (*bangumi.DownloadState, error) {
	dl.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("using pikpak + aria2 download episode")
	torr, err := torrent.Load(bytes.NewBuffer(episode.Torrent))
	if err != nil {
		return nil, err
	}
	torrHashBytes := torr.HashInfoBytes()
	torrInfo, err := torr.UnmarshalInfo()
	if err != nil {
		return nil, err
	}
	magnet := torr.Magnet(&torrHashBytes, &torrInfo).String()

	baseName := fmt.Sprintf("[%s] S%01dE%02d", info.Title, seasonNum, episode.Number)
	files, err := dl.pikpak.OfflineDownAndWait(baseName, magnet, time.Minute*5)
	if err != nil {
		dl.logger.Error().Err(err).Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("pikpak download error")
		return nil, err
	}
	dl.logger.Debug().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("pikpak download complete")

	var gids []string
	var downloadingFiles []*pikpak.File

	clear := func() {
		for _, gid := range gids {
			_ = dl.aria2.RemoveDownloadResult(gid)
		}
	}

	dir := bangumi.DirNaming(info, seasonNum)
	for _, fi := range files {
		newFilename := bangumi.RenamingEpisodeFileName(info, seasonNum, &episode, fi.Name)
		if newFilename != "" {
			dl.logger.Debug().Str("file", newFilename).Msg("try add aria2 task")
			gid, err := dl.aria2.AddTask(fi.DownloadUrl, dir, newFilename)
			if err != nil {
				dl.logger.Error().Err(err).Str("file", newFilename).Msg("add aria2 task error, will remove all tasks")
				clear()
				return nil, err
			}
			dl.logger.Debug().Str("file", newFilename).Msg("add aria2 task success")
			gids = append(gids, gid.GID)
			downloadingFiles = append(downloadingFiles, fi)
		} else {
			dl.logger.Warn().Str("file", fi.Name).Msg("unable to rename file")
		}
	}

	go dl.waitAria2DownloadComplete(gids, info, seasonNum, episode)

	return &bangumi.DownloadState{
		Downloader: "aria2",
		TaskId:     strings.Join(gids, ","),
	}, nil
}

func (dl *SmartDownloader) waitAria2DownloadComplete(gids []string, info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) {
	if len(gids) == 0 {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(len(gids))

	for i, gid := range gids {
		dl.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Str("gid", gid).Msg(fmt.Sprintf("waiting aria2 task complete %d/%d", i, len(gids)))
		dl.aria2.WaitDownloadComplete(gid, func(status arigo.Status) {
			if status.Status == arigo.StatusError {
				err := errors.New(status.ErrorMessage)
				// callback
				dl.logger.Error().Err(err).Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Str("gid", gid).Msg(fmt.Sprintf("waiting aria2 task complete %d/%d", i, len(gids)))
				dl.onErr(err, info, seasonNum, episode)
			} else {
				dl.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Str("gid", gid).Msg(fmt.Sprintf("aria2 task complete %d/%d", i, len(gids)))
				dl.onComplete(info, seasonNum, episode)
			}
		})
		wg.Done()
	}

	go func() {
		// wait all files download complete
		dl.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("wait all files download complete")

		wg.Wait()

		dl.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("all files download complete")

		// callback
		dl.onComplete(info, seasonNum, episode)
	}()
}

func (dl *SmartDownloader) downloadUsingQibittorrent(info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) (*bangumi.DownloadState, error) {
	if episode.TorrentHash != "" {
		dl.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("using qibittorrent download episode")

		// try download
		opts := qibittorrent.AddTorrentOptions{
			Paused: true,
			Rename: bangumi.RenamingEpisodeFileName(info, seasonNum, &episode, info.Title),
		}
		dl.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("start download episode")
		hash, err := dl.qb.AddTorrentEx(&opts, episode.Torrent, bangumi.DirNaming(info, seasonNum))
		if err != nil {
			return nil, err
		}
		go func() {
			// Wait for torrent parsing complete
			dl.logger.Info().Msg("wait for torrent parsing complete")
			for {
				torrentTask, err := dl.qb.GetTorrent(episode.TorrentHash)
				if err == nil {
					if torrentTask != nil && torrentTask.State == qibittorrent.StatePausedDL {
						break
					}
				}
				time.Sleep(time.Second)
			}

			// renaming torrent files
			err = dl.renameTorrent(hash, info, seasonNum, &episode)
			if err != nil {
				dl.logger.Error().Err(err).Msg("rename torrent files error")
				return
			}

			// resume
			err = dl.qb.ResumeTorrents([]string{hash})
			if err != nil {
				dl.logger.Error().Err(err).Msg("resume torrent error")
				return
			}

			// wait for download complete
			go dl.waitQbDownloadComplete(hash, info, seasonNum, episode)
		}()

	} else {
		dl.logger.Warn().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("skip episode, torrent hash is empty")
	}
	return &bangumi.DownloadState{
		Downloader: "qb",
		TaskId:     episode.TorrentHash,
	}, nil
}

func (dl *SmartDownloader) waitQbDownloadComplete(hash string, info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) {
	err := dl.qb.WaitForDownloadComplete(hash, time.Second*5, func() bool {
		dl.logger.Info().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Msg("download complete")
		dl.onComplete(info, seasonNum, episode)
		return true
	})
	if err != nil {
		dl.logger.Warn().Err(err).Msg("wait for download complete error, torrent maybe remove")
		dl.onErr(err, info, seasonNum, episode)
	}
}

func (dl *SmartDownloader) renameTorrent(hash string, info *bangumi.BangumiInfo, seasonNum uint, episode *bangumi.Episode) error {
	content, err := dl.qb.GetTorrentContent(hash, []int64{})
	if err != nil {
		return err
	}
	for _, fi := range content {
		newName := bangumi.RenamingEpisodeFileName(info, seasonNum, episode, fi.Name)
		if newName != "" {
			err = dl.qb.RenameFile(hash, fi.Name, newName)
			if err != nil {
				return err
			}
			dl.logger.Info().Str("hash", hash).Str("filename", fi.Name).Str("new filename", newName).Msg("rename episode")
		} else {
			dl.logger.Warn().Str("hash", hash).Str("filename", fi.Name).Msg("unable to rename file")
		}
	}
	return nil
}
