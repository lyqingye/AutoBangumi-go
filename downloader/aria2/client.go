package aria2

import (
	"autobangumi-go/utils"
	"context"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/siku2/arigo"
)

var Aria2DownloadOptions arigo.Options = arigo.Options{
	AutoFileRenaming: false,
	MaxFileNotFound:  100,
	MaxTries:         100,
	AllowOverwrite:   true,
}

type Client struct {
	url    string
	secret string
	inner  *atomic.Pointer[arigo.Client]
	logger zerolog.Logger
	dir    string
}

func NewClient(url string, secret string, dir string) (*Client, error) {
	inner, err := arigo.NewClient2(context.Background(), url, secret)
	if err != nil {
		return nil, err
	}
	cli := Client{
		logger: utils.GetLogger("aria2"),
		url:    url,
		secret: secret,
		dir:    dir,
		inner:  &atomic.Pointer[arigo.Client]{},
	}
	cli.inner.Store(inner)
	go cli.run()
	return &cli, nil
}

func (cli *Client) run() {
	for {
		reconCount := 1
		cli.inner.Load().Run()
		cli.logger.Warn().Int("count", reconCount).Msg("connection close, try reconn")
		// recon
		for {
			newClient, err := arigo.NewClient2(context.Background(), cli.url, cli.secret)
			if err == nil {
				cli.inner.Store(newClient)
				cli.logger.Info().Msg("reconn success")
				break
			} else {
				cli.logger.Error().Err(err).Msg("reconn error")
				reconCount += 1
				time.Sleep(time.Second * 1)
			}
		}
	}
}

func (cli *Client) AddTask(url string, dir string, filename string) (arigo.GID, error) {
	aria2 := cli.inner.Load()

	opts := Aria2DownloadOptions
	opts.Out = filename
	opts.Dir = filepath.Join(cli.dir, dir)

	status, err := aria2.AddURI([]string{url}, &opts)

	if err != nil {
		return arigo.GID{}, err
	}

	return status, nil
}

func (cli *Client) RemoveTask(gid string) error {
	aria2 := cli.inner.Load()
	return aria2.Remove(gid)
}

func (cli *Client) ForceRemove(gid string) error {
	aria2 := cli.inner.Load()
	return aria2.ForceRemove(gid)
}

func (cli *Client) PauseTask(gid string) error {
	aria2 := cli.inner.Load()
	return aria2.Pause(gid)
}

func (cli *Client) ForcePauseTask(gid string) error {
	aria2 := cli.inner.Load()
	return aria2.ForcePause(gid)
}

func (cli *Client) ResumeTask(gid string) error {
	aria2 := cli.inner.Load()
	return aria2.Unpause(gid)
}

func (cli *Client) ResumeAllTasks() error {
	aria2 := cli.inner.Load()
	return aria2.UnpauseAll()
}

func (cli *Client) QueryStatus(gid string) (arigo.Status, error) {
	aria2 := cli.inner.Load()
	return aria2.TellStatus(gid, []string{}...)
}

func (cli *Client) IterAllTasks(fn func(status arigo.Status) bool) error {
	activeTasks, err := cli.ListActiveTasks()
	if err != nil {
		return err
	}

	for _, task := range activeTasks {
		if fn(task) {
			return nil
		}
	}

	err = cli.IterStoppedTasks(fn)
	if err != nil {
		return err
	}

	err = cli.IterWaitingTasks(fn)
	if err != nil {
		return err
	}
	return nil
}

func (cli *Client) ListActiveTasks() ([]arigo.Status, error) {
	aria2 := cli.inner.Load()
	return aria2.TellActive([]string{}...)
}

func (cli *Client) IterStoppedTasks(fn func(status arigo.Status) bool) error {
	offset := 0
	var limit uint = 100
	for {
		aria2 := cli.inner.Load()
		statusList, err := aria2.TellStopped(offset, limit, []string{}...)
		if err != nil {
			return err
		}

		for _, status := range statusList {
			if fn(status) {
				break
			}
		}
		if len(statusList) < int(limit) {
			break
		}
		offset = offset + len(statusList)
	}
	return nil
}

func (cli *Client) IterWaitingTasks(fn func(status arigo.Status) bool) error {
	offset := 0
	var limit uint = 100
	for {
		aria2 := cli.inner.Load()
		statusList, err := aria2.TellWaiting(offset, limit, []string{}...)
		if err != nil {
			return err
		}

		for _, status := range statusList {
			if fn(status) {
				break
			}
		}
		if len(statusList) < int(limit) {
			break
		}
		offset = offset + len(statusList)
	}
	return nil
}

func (cli *Client) WaitDownloadComplete(gid string, fn func(status arigo.Status)) {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		status, err := cli.QueryStatus(gid)
		if err == nil {
			if status.Status == arigo.StatusCompleted || status.Status == arigo.StatusError {
				fn(status)
				return
			}
		}
	}
}

func (cli *Client) WatchTask(gid string, period time.Duration, fn func(status arigo.Status) bool) {
	ticker := time.NewTicker(period)
	for range ticker.C {
		status, err := cli.QueryStatus(gid)
		if err == nil {
			if fn(status) {
				return
			}
		}
	}
}

func (cli *Client) RemoveDownloadResult(gid string) error {
	aria2 := cli.inner.Load()
	return aria2.RemoveDownloadResult(gid)
}
