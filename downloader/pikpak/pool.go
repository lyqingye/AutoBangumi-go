package pikpak

import (
	"autobangumi-go/utils"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog"

	pikpakgo "github.com/lyqingye/pikpak-go"
)

const (
	StateRestricted  = "RestrictedDailyCreateLimit"
	StateNoSpaceLeft = "NoSpaceLeft"
	StateNormal      = "Normal"
)

type File struct {
	Id          string
	Name        string
	DownloadUrl string
	RefAcc      string
}

type Account struct {
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	State          string    `json:"state"`
	RestrictedTime time.Time `json:"restricted_time"`
}

type Pool struct {
	logger             zerolog.Logger
	lock               sync.RWMutex
	accounts           map[string]Account
	restrictedAccounts map[string]Account
	clients            map[string]*pikpakgo.PikPakClient
	fileUrlToFile      map[string]*File
	configPath         string
}

func NewPool(configPath string) (*Pool, error) {
	pool := Pool{
		logger:             utils.GetLogger("pikpak-pool"),
		lock:               sync.RWMutex{},
		accounts:           map[string]Account{},
		restrictedAccounts: map[string]Account{},
		clients:            map[string]*pikpakgo.PikPakClient{},
		fileUrlToFile:      map[string]*File{},
		configPath:         configPath,
	}
	err := pool.loadAccountsFromConfigFile()
	if err != nil {
		return nil, err
	}
	go pool.autoRefreshAccounts()
	return &pool, nil
}

func (pool *Pool) loadAccountsFromConfigFile() error {
	bz, err := os.ReadFile(pool.configPath)
	if os.IsNotExist(err) {
		return nil
	}
	var accounts []Account
	err = json.Unmarshal(bz, &accounts)
	if err != nil {
		return err
	}
	for _, acc := range accounts {
		pool.logger.Info().Str("username", acc.Username).Msg("load pikpak account")
		pool.accounts[acc.Username] = acc
	}
	return nil
}

func (pool *Pool) AddAccount(username, password string) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	pool.accounts[username] = Account{
		Username:       username,
		Password:       password,
		State:          StateNormal,
		RestrictedTime: time.Time{},
	}
}

func (pool *Pool) OfflineDownAndWait(name, magnet string, timeout time.Duration) ([]*File, error) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.logger.Info().Str("name", name).Msg("try to add offline task")

Retry:
	client, acc, err := pool.getClient()
	if err != nil {
		return nil, err
	}

	pool.logger.Debug().Str("name", name).Msg("try to find exists offline task")
	var task *pikpakgo.Task
	err = client.OfflineListIterator(func(t *pikpakgo.Task) bool {
		sameTask := strings.Contains(magnet, t.Params.URL) ||
			t.Params.URL == magnet ||
			strings.Contains(t.Params.URL, magnet)
		if sameTask {
			if t.Phase == pikpakgo.PhaseTypeComplete {
				task = t
				return true
			}
		}
		return false
	})

	if err != nil {
		return nil, err
	}

	if task == nil {
		// offline download and wait
		pool.logger.Debug().Str("name", name).Msg("add offline task")
		newTask, err := client.OfflineDownload(name, magnet, "")
		if err != nil {
			if errors.Is(err, pikpakgo.ErrDailyCreateLimit) {
				pool.logger.Warn().Str("username", acc.Username).Str("reason", err.Error()).Msg("restrict account")
				pool.setAccountRestricted(acc, StateRestricted)
			} else if errors.Is(err, pikpakgo.ErrSpaceNotEnough) {
				pool.logger.Warn().Str("username", acc.Username).Str("reason", err.Error()).Msg("restrict account")
				pool.setAccountRestricted(acc, StateNoSpaceLeft)
			}
			if acc.State == StateNormal {
				return nil, err
			}
			pool.logger.Error().Err(err).Msg("offline download error")
			goto Retry
		}
		task = newTask.Task
	} else {
		pool.logger.Debug().Str("name", name).Msg("find exists offline task")
	}

	pool.logger.Info().Str("name", name).Msg("wait for offline task finished")
	finishedTask, err := client.WaitForOfflineDownloadComplete(task.ID, timeout, func(t *pikpakgo.Task) {
		pool.logger.Debug().Int("progress", t.Progress).Str("name", t.FileName).Msg("task update")
	})

	if err != nil {
		return nil, err
	}

	pool.logger.Info().Str("name", name).Msg("offline task finished")
	if finishedTask.Phase == pikpakgo.PhaseTypeError {
		pool.logger.Error().Str("name", name).Str("detail", finishedTask.Message).Msg("offline task error")
		return nil, fmt.Errorf("offline task error: %s", finishedTask.Message)
	}

	pool.logger.Info().Str("name", name).Msg("offline task success")

	// walk dir and get download url
	var pikpakFiles []*pikpakgo.File
	err = client.WalkDir(finishedTask.FileID, func(file *pikpakgo.File) bool {
		if file.Kind == pikpakgo.KindOfFile {
			pikpakFiles = append(pikpakFiles, file)
		}
		return false
	})

	if err != nil {
		return nil, err
	}

	var files []*File
	for _, f := range pikpakFiles {
		downloadUrl, err := client.GetDownloadUrl(f.ID)
		if err != nil {
			pool.logger.Error().Err(err).Str("name", f.Name).Msg("get file download url err")
			return nil, err
		}
		file := File{
			Id:          f.ID,
			Name:        f.Name,
			DownloadUrl: downloadUrl,
			RefAcc:      acc.Username,
		}
		pool.logger.Debug().Str("name", f.Name).Msg("get file download url success")
		files = append(files, &file)
		pool.fileUrlToFile[downloadUrl] = &file
	}
	return files, nil
}

func (pool *Pool) RemoveFile(downloadUrl string) error {
	pool.lock.Lock()
	defer pool.lock.Lock()
	var file *File
	if f, found := pool.fileUrlToFile[downloadUrl]; found {
		file = f
	} else {
		return errors.New("file not found")
	}

	var acc *Account
	if a, found := pool.accounts[file.RefAcc]; found {
		acc = &a
	} else if b, found := pool.restrictedAccounts[file.RefAcc]; found {
		acc = &b
	}
	if acc != nil {
		client, err := pool.getClientByAcc(acc)
		if err != nil {
			return err
		}
		err = client.BatchTrashFiles([]string{file.Id})
		if err != nil {
			return err
		}

		delete(pool.fileUrlToFile, downloadUrl)

		err = client.BatchDeleteFiles([]string{file.Id})
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("acc not found: %s", file.RefAcc)
	}
	pool.logger.Info().Str("username", acc.Username).Str("name", file.Name).Msg("recycle account storage")
	return pool.refreshAccount(acc)
}

func (pool *Pool) chooseAccount() (*Account, error) {
	if len(pool.accounts) == 0 {
		err := pool.refreshAccounts()
		if err != nil {
			return nil, err
		}
	}
	if len(pool.accounts) == 0 {
		return nil, errors.New("no available pikpak accounts")
	}
	for _, acc := range pool.accounts {
		pool.logger.Debug().Str("username", acc.Username).Msg("select pikpak account")
		return &acc, nil
	}
	panic("unreachable code")
}

func (pool *Pool) getClient() (*pikpakgo.PikPakClient, *Account, error) {
	acc, err := pool.chooseAccount()
	if err != nil {
		return nil, nil, err
	}
	client, err := pool.getClientByAcc(acc)
	return client, acc, err
}

func (pool *Pool) getClientByAcc(acc *Account) (*pikpakgo.PikPakClient, error) {
	if cli, found := pool.clients[acc.Username]; found {
		pool.clients[acc.Username] = cli
		return cli, nil
	} else {
		cli, err := pikpakgo.NewPikPakClient(acc.Username, acc.Password)
		if err != nil {
			return nil, err
		}
		err = cli.Login()
		if err != nil {
			return nil, err
		}
		pool.clients[acc.Username] = cli
		return cli, nil
	}
}

func (pool *Pool) refreshAccounts() error {
	for _, acc := range pool.restrictedAccounts {
		err := pool.refreshAccount(&acc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pool *Pool) refreshAccount(acc *Account) error {
	if acc.State == StateNoSpaceLeft {
		client, err := pool.getClientByAcc(acc)
		if err != nil {
			return err
		}
		about, err := client.About()
		if err != nil {
			return err
		}
		if about.Quota != nil {
			free := about.Quota.Limit - about.Quota.Usage
			// 500MB
			if free >= 524288000 {
				if about.Quota.UsageInTrash > 0 {
					err = client.EmptyTrash()
					if err != nil {
						return err
					}
				}
				pool.logger.Info().Str("username", acc.Username).Str("free", humanize.Bytes(uint64(free))).Msg("account recover from NoSpaceLeft state")
			} else {
				return nil
			}
		}
	} else if acc.State == StateRestricted {
		now := utils.GetMidnightTime()
		restrictedTime := utils.TimeToMidnightTime(acc.RestrictedTime)
		if now.Sub(restrictedTime).Hours() < 24 {
			return nil
		}
		pool.logger.Info().Str("username", acc.Username).Msg("account recover from Restricted state")
	}
	acc.State = StateNormal
	delete(pool.restrictedAccounts, acc.Username)
	pool.accounts[acc.Username] = *acc
	return nil
}

func (pool *Pool) setAccountRestricted(acc *Account, newState string) {
	acc.State = newState
	acc.RestrictedTime = time.Now()
	delete(pool.accounts, acc.Username)
	pool.restrictedAccounts[acc.Username] = *acc
}

func (pool *Pool) autoRefreshAccounts() {
	ticker := time.NewTicker(time.Hour * 1)
	for range ticker.C {
		pool.logger.Info().Msg("auto refresh accounts")
		err := pool.refreshAccounts()
		if err != nil {
			pool.logger.Error().Err(err).Msg("auto refresh accounts error")
		}
	}
}
