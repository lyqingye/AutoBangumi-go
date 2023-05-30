package downloader

import (
	"autobangumi-go/bangumi"
	"autobangumi-go/downloader/aria2"
	"autobangumi-go/downloader/pikpak"
	"autobangumi-go/downloader/qbittorrent"
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
	qb        *qbittorrent.QbittorrentClient
	callbacks []Callback
}

func NewSmartDownloader(aria2 *aria2.Client, pikpak *pikpak.Pool, qb *qbittorrent.QbittorrentClient) (*SmartDownloader, error) {
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
	l := dl.logger.With().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Logger()
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

	l.Info().Msg("start download episode")
	state, err := dl.downloadUsingPikpakAndAria2(info, seasonNum, episode)
	if err != nil {
		l.Warn().Err(err).Msg("using pikpak download error, fallback to qibittorrent")
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

func (dl *SmartDownloader) findQBTask(episode bangumi.Episode) (*qbittorrent.Torrent, error) {
	// check torrent downloading
	torrentTask, err := dl.qb.GetTorrent(episode.TorrentHash)
	if err != qbittorrent.ErrTorrentNotFound && err != nil {
		return nil, nil
	}
	return torrentTask, nil
}

// DownloadMagnetAndWait download and wait download complete
func (dl *SmartDownloader) downloadUsingPikpakAndAria2(info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) (*bangumi.DownloadState, error) {
	l := dl.logger.With().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Logger()

	l.Info().Msg("using pikpak + aria2 download episode")
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
	files, err := dl.pikpak.OfflineDownAndWait(baseName, magnet, time.Minute*3)
	if err != nil {
		l.Error().Err(err).Msg("pikpak download error")
		return nil, err
	}
	l.Debug().Msg("pikpak download complete")

	var gids []string

	clear := func() {
		for _, gid := range gids {
			_ = dl.aria2.RemoveDownloadResult(gid)
		}
	}

	dir := bangumi.DirNaming(info, seasonNum)
	for _, fi := range files {
		newFilename := bangumi.RenamingEpisodeFileName(info, seasonNum, &episode, fi.Name)
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

	go dl.waitAria2DownloadComplete(gids, info, seasonNum, episode)

	return &bangumi.DownloadState{
		Downloader: "aria2",
		TaskId:     strings.Join(gids, ","),
	}, nil
}

func (dl *SmartDownloader) waitAria2DownloadComplete(gids []string, info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) {
	l := dl.logger.With().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Logger()
	if len(gids) == 0 {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(len(gids))

	for i, gid := range gids {
		l.Info().Str("gid", gid).Msg(fmt.Sprintf("waiting aria2 task complete %d/%d", i, len(gids)))
		dl.aria2.WaitDownloadComplete(gid, func(status arigo.Status) {
			if status.Status == arigo.StatusError {
				err := errors.New(status.ErrorMessage)
				// callback
				l.Error().Err(err).Str("gid", gid).Msg(fmt.Sprintf("waiting aria2 task complete %d/%d", i, len(gids)))
				dl.onErr(err, info, seasonNum, episode)
			} else {
				l.Info().Str("gid", gid).Msg(fmt.Sprintf("aria2 task complete %d/%d", i, len(gids)))
			}
		})
		wg.Done()
	}

	go func() {
		// wait all files download complete
		l.Info().Msg("wait all files download complete")

		wg.Wait()

		l.Info().Msg("all files download complete")

		// callback
		dl.onComplete(info, seasonNum, episode)
	}()
}

func (dl *SmartDownloader) downloadUsingQibittorrent(info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) (*bangumi.DownloadState, error) {
	l := dl.logger.With().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Logger()

	if episode.TorrentHash != "" {
		l.Info().Msg("using qibittorrent download episode")

		// try download
		opts := qbittorrent.AddTorrentOptions{
			Paused: true,
			Rename: bangumi.RenamingEpisodeFileName(info, seasonNum, &episode, info.Title),
		}
		l.Info().Msg("start download episode")
		hash, err := dl.qb.AddTorrentEx(&opts, episode.Torrent, bangumi.DirNaming(info, seasonNum))
		if err != nil {
			return nil, err
		}
		go func() {
			// Wait for torrent parsing complete
			l.Info().Msg("wait for torrent parsing complete")
			for {
				torrentTask, err := dl.qb.GetTorrent(episode.TorrentHash)
				if err == nil {
					if torrentTask != nil && torrentTask.State == qbittorrent.StatePausedDL {
						break
					}
				}
				time.Sleep(time.Second)
			}

			// renaming torrent files
			err = dl.renameTorrent(hash, info, seasonNum, &episode)
			if err != nil {
				l.Error().Err(err).Msg("rename torrent files error")
				return
			}

			// resume
			err = dl.qb.ResumeTorrents([]string{hash})
			if err != nil {
				l.Error().Err(err).Msg("resume torrent error")
				return
			}

			// wait for download complete
			go dl.waitQbDownloadComplete(hash, info, seasonNum, episode)
		}()

	} else {
		l.Warn().Msg("skip episode, torrent hash is empty")
	}
	return &bangumi.DownloadState{
		Downloader: "qb",
		TaskId:     episode.TorrentHash,
	}, nil
}

func (dl *SmartDownloader) waitQbDownloadComplete(hash string, info *bangumi.BangumiInfo, seasonNum uint, episode bangumi.Episode) {
	l := dl.logger.With().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Logger()
	err := dl.qb.WaitForDownloadComplete(hash, time.Second*5, func() bool {
		l.Info().Msg("download complete")
		dl.onComplete(info, seasonNum, episode)
		return true
	})
	if err != nil {
		l.Warn().Err(err).Msg("wait for download complete error, torrent maybe remove")
		dl.onErr(err, info, seasonNum, episode)
	}
}

func (dl *SmartDownloader) renameTorrent(hash string, info *bangumi.BangumiInfo, seasonNum uint, episode *bangumi.Episode) error {
	l := dl.logger.With().Str("title", info.Title).Uint("season", seasonNum).Uint("episode", episode.Number).Logger()
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
			l.Info().Str("filename", fi.Name).Str("new filename", newName).Msg("rename episode")
		} else {
			l.Warn().Str("filename", fi.Name).Msg("unable to rename file, skip to download this file")
			_ = dl.qb.SetFilePriority(hash, []int{fi.Index}, 0)
		}
	}
	return nil
}
