package aria2

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"pikpak-bot/downloader"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog/log"
	"github.com/siku2/arigo"
)

var (
	ErrNoConn       = errors.New("did not connect to the server")
	TaskTypeStopped = "stopped"
	TaskTypeActive  = "actived"
	TaskTypeWatting = "watting"
)

var Aria2DownloadOptions arigo.Options = arigo.Options{
	AutoFileRenaming: false,
	MaxFileNotFound:  100,
	MaxTries:         100,
}

type Aria2OnlineDownloader struct {
	jsonRpc string
	secret  string
	client  *arigo.Client
	dir     string
}

func NewAria2OnlineDownloader(dir, jsonRpc string, secret string) (*Aria2OnlineDownloader, error) {
	downloader := Aria2OnlineDownloader{
		jsonRpc: jsonRpc,
		secret:  secret,
		dir:     dir,
	}

	connFn := func() (*arigo.Client, error) {
		return arigo.NewClientFromUrl(context.Background(), jsonRpc, secret)
	}

	go func() {
		for {
			if downloader.client == nil {
				client, err := connFn()
				if err == nil {
					downloader.client = client
					client.Run()
				} else {
					println(err.Error())
				}
			}
			downloader.client = nil
			time.Sleep(time.Second)
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * 30)
		for range ticker.C {
			downloader.resumeStoppedTasks()
		}
	}()

	return &downloader, nil
}

func (a *Aria2OnlineDownloader) AddTask(url string, dir string, filename string) (string, error) {
	if a.client == nil {
		return "", ErrNoConn
	}
	opts := Aria2DownloadOptions
	opts.Out = filename
	opts.Dir = filepath.Join(a.dir, dir)
	status, err := a.client.AddURI([]string{url}, &opts)
	if err != nil {
		return "", err
	}
	return status.GID, nil
}

func (a *Aria2OnlineDownloader) QueryStatus(tid string) (downloader.DownloadStatus, error) {
	if a.client == nil {
		return downloader.DownloadStatus{}, ErrNoConn
	}
	aria2Status, err := a.client.TellStatus(tid)
	if err != nil {
		return downloader.DownloadStatus{}, nil
	}
	return aria2StatusToStatus(aria2Status), nil
}

func (a *Aria2OnlineDownloader) RemoveTask(tid string) error {
	if a.client == nil {
		return ErrNoConn
	}
	return a.client.ForceRemove(tid)
}

func (a *Aria2OnlineDownloader) PauseTask(tid string) error {
	if a.client == nil {
		return ErrNoConn
	}
	return a.client.ForcePause(tid)
}

func (a *Aria2OnlineDownloader) ResumeTask(tid string) error {
	if a.client == nil {
		return ErrNoConn
	}
	return a.client.Unpause(tid)
}

func (a *Aria2OnlineDownloader) ListTasks() ([]downloader.DownloadStatus, error) {
	if a.client == nil {
		return nil, ErrNoConn
	}
	statusList, err := a.client.TellActive([]string{}...)
	if err != nil {
		return nil, err
	}
	var result []downloader.DownloadStatus
	for _, s := range statusList {
		result = append(result, aria2StatusToStatus(s))
	}
	return result, nil
}

func (a *Aria2OnlineDownloader) resumeStoppedTasks() {
	a.walkTasks(TaskTypeStopped, func(cli *arigo.Client, as arigo.Status) bool {
		var err error
		switch as.Status {
		case arigo.StatusPaused:
			err = cli.Unpause(as.GID)
			if err != nil {
				log.Error().Err(err).Str("GID", as.GID).Msg("failed to Unpause task")
			}
			log.Info().Str("GID", as.GID).Msg("resume paused task")
		case arigo.StatusError:
			err = cli.RemoveDownloadResult(as.GID)
			if err == nil {
				// retry
				for _, f := range as.Files {
					var urls []string
					for _, uri := range f.URIs {
						urls = append(urls, uri.URI)
					}
					fileName := filepath.Base(f.Path)
					opts := Aria2DownloadOptions
					opts.Out = fileName
					opts.Dir = filepath.Dir(f.Path)
					_, err = a.client.AddURI(urls, &opts)
					if err != nil {
						log.Error().Err(err).Str("file", fileName).Msg("failed to retry download file")
					} else {
						log.Info().Str("file", fileName).Msg("retry download file")
					}
				}
			} else {
				log.Error().Err(err).Str("GID", as.GID).Msg("failed to remove error task")
			}
		}
		return false
	})
}

func (a *Aria2OnlineDownloader) walkTasks(taskType string, callback func(cli *arigo.Client, status arigo.Status) bool) {
	offset := 0
	var limit uint = 100
	var err error
	for {
		cli := a.client
		if cli != nil {
			var aria2Status []arigo.Status
			switch taskType {
			case TaskTypeActive:
				aria2Status, err = cli.TellActive([]string{}...)
			case TaskTypeWatting:
				aria2Status, err = cli.TellWaiting(offset, limit, []string{}...)
			case TaskTypeStopped:
				aria2Status, err = cli.TellStopped(offset, limit, []string{}...)
			}
			if err == nil {
				for _, as := range aria2Status {
					if callback(cli, as) {
						return
					}
				}

				switch taskType {
				case TaskTypeActive:
					return
				case TaskTypeWatting, TaskTypeStopped:
					if len(aria2Status) < int(limit) {
						return
					}
					offset = offset + len(aria2Status)
				}
			} else {
				log.Error().Err(err).Msg("tell stopped failed")
				time.Sleep(5 * time.Second)
				continue
			}
		}
	}

}

func (a *Aria2OnlineDownloader) AddTaskCompleteCallback(tid string, callback func(status downloader.DownloadStatus) bool) {
	go func() {
		ticker := time.NewTicker(time.Second * 30)
		for range ticker.C {
			cli := a.client
			status, err := cli.TellStatus(tid, []string{}...)
			if err == nil {
				switch status.Status {
				case arigo.StatusCompleted:
					if callback(aria2StatusToStatus(status)) {
						return
					}
				}
			}
		}
	}()
}

func aria2StatusToStatus(aria2Status arigo.Status) downloader.DownloadStatus {
	status := downloader.DownloadStatus{}
	// speed
	speed := humanize.Bytes(uint64(aria2Status.DownloadSpeed))
	status.Speed = fmt.Sprintf("%s/s", speed)

	// progress
	total := humanize.Bytes(uint64(aria2Status.TotalLength))
	downloaded := humanize.Bytes(uint64(aria2Status.CompletedLength))
	status.Progress = fmt.Sprintf("%s/%s", downloaded, total)

	// filename
	var filePaths []string
	for _, f := range aria2Status.Files {
		filePaths = append(filePaths, f.Path)
	}
	status.Filename = strings.Join(filePaths, ",")

	// status
	switch aria2Status.Status {
	case arigo.StatusActive:
		status.Status = downloader.StatusDownloading
	case arigo.StatusCompleted:
		status.Status = downloader.StatusFinished
	case arigo.StatusError:
		status.Status = downloader.StatusError
		status.ErrMessage = aria2Status.ErrorMessage
	case arigo.StatusPaused:
		status.Status = downloader.StatusPause
	case arigo.StatusWaiting:
		status.Status = downloader.StatusWaiting
	case arigo.StatusRemoved:
		status.Status = downloader.StatusRemoved
	}

	return status
}
