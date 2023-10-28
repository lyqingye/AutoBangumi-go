package pikpak

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"autobangumi-go/config"
	"autobangumi-go/utils"

	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog"

	pikpakgo "github.com/lyqingye/pikpak-go"
)

const (
	StateRestricted  = "RestrictedDailyCreateLimit"
	StateNoSpaceLeft = "NoSpaceLeft"
	StateNormal      = "Normal"
)

var ErrNoAvailableAccount = errors.New("no available account")

type File struct {
	Id          string
	Name        string
	DownloadUrl string
	RefAcc      string
}

type Account struct {
	Username       string
	Password       string
	State          string
	RestrictedTime int64
}

func (acc *Account) GetRestrictedTime() time.Time {
	return time.Unix(acc.RestrictedTime, 0)
}

type AccountStorage interface {
	ListAccounts() ([]Account, error)
	ListAccountsByState(state string) ([]Account, error)
	UpdateAccount(acc Account) error
	AddAccount(acc Account) error
	GetAccount(username string) (Account, error)
}

type Pool struct {
	logger        zerolog.Logger
	lock          sync.RWMutex
	clients       map[string]*pikpakgo.PikPakClient
	fileUrlToFile map[string]*File
	tempDirectory string
	storage       AccountStorage
	cfg           config.PikpakConfig
}

func NewPool(storage AccountStorage, cfg config.PikpakConfig) (*Pool, error) {
	pool := Pool{
		logger:        utils.GetLogger("pikpak-pool"),
		lock:          sync.RWMutex{},
		clients:       map[string]*pikpakgo.PikPakClient{},
		fileUrlToFile: map[string]*File{},
		tempDirectory: "/temp",
		storage:       storage,
		cfg:           cfg,
	}
	// recycle storage on startup
	go pool.recycleAllAccStorage()
	return &pool, nil
}

func (pool *Pool) OfflineDownAndWait(name, magnet string) ([]*File, error) {
	pool.lock.Lock()
	pool.logger.Info().Str("name", name).Msg("try to add offline task")

Retry:
	client, acc, err := pool.getClient()
	if err != nil {
		pool.lock.Unlock()
		return nil, err
	}

	pool.logger.Debug().Str("name", name).Msg("try to find exists offline task")
	var task *pikpakgo.Task
	err = client.OfflineListIterator(func(t *pikpakgo.Task) bool {
		sameTask := strings.Contains(magnet, t.Params.URL) ||
			t.Params.URL == magnet ||
			strings.Contains(t.Params.URL, magnet)
		if sameTask {
			if t.Phase == pikpakgo.PhaseTypeError {
				_ = client.OfflineRemove([]string{t.ID}, true)
			}
			if t.Phase == pikpakgo.PhaseTypeComplete {
				task = t
				return true
			}
		}
		return false
	})

	if err != nil {
		pool.lock.Unlock()
		return nil, err
	}

	if task == nil {
		// offline download and wait
		pool.logger.Debug().Str("name", name).Msg("add offline task")
		pid, err := client.FolderPathToID(pool.tempDirectory, true)
		if err != nil {
			goto Retry
		}
		newTask, err := client.OfflineDownload(name, magnet, pid)
		if err != nil {
			if errors.Is(err, pikpakgo.ErrDailyCreateLimit) {
				pool.logger.Warn().Str("username", acc.Username).Str("reason", err.Error()).Msg("restrict account")
				pool.setAccountRestricted(acc, StateRestricted)
			} else if errors.Is(err, pikpakgo.ErrSpaceNotEnough) {
				pool.logger.Warn().Str("username", acc.Username).Str("reason", err.Error()).Msg("restrict account")
				pool.setAccountRestricted(acc, StateNoSpaceLeft)
			}
			if newTask != nil && newTask.Task != nil && newTask.Task.ID != "" {
				_ = client.OfflineRemove([]string{newTask.Task.ID}, true)
			}
			if acc.State == StateNormal {
				pool.lock.Unlock()
				return nil, err
			}
			pool.logger.Error().Err(err).Msg("offline download error")
			goto Retry
		}
		task = newTask.Task
	} else {
		pool.logger.Debug().Str("name", name).Msg("find exists offline task")
	}

	pool.lock.Unlock()

	pool.logger.Info().Str("name", name).Msg("wait for offline task finished")
	finishedTask, err := client.WaitForOfflineDownloadComplete(task.ID, pool.cfg.OfflineDownloadTimeout, func(t *pikpakgo.Task) {
		pool.logger.Debug().Int("progress", t.Progress).Str("name", t.FileName).Msg("task update")
	})

	if err != nil {
		if errors.Is(err, pikpakgo.ErrWaitForOfflineDownloadTimeout) {
			_ = client.OfflineRemove([]string{task.ID}, true)
		}
		return nil, err
	}

	pool.logger.Info().Str("name", name).Msg("offline task finished")
	if finishedTask.Phase == pikpakgo.PhaseTypeError {
		_ = client.OfflineRemove([]string{task.ID}, true)
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
	defer pool.lock.Unlock()
	var file *File
	if f, found := pool.fileUrlToFile[downloadUrl]; found {
		file = f
	} else {
		return errors.New("file not found")
	}

	acc, err := pool.storage.GetAccount(file.RefAcc)
	if err != nil {
		return fmt.Errorf("acc not found: %s", file.RefAcc)
	} else {
		client, err := pool.getClientByAcc(&acc)
		if err != nil {
			return err
		}
		err = client.BatchDeleteFiles([]string{file.Id})
		if err != nil {
			return err
		}
		delete(pool.fileUrlToFile, downloadUrl)

	}
	pool.logger.Info().Str("username", acc.Username).Str("name", file.Name).Msg("recycle account storage")
	return pool.refreshAccount(&acc)
}

func (pool *Pool) chooseAccount() (*Account, error) {
	accounts, err := pool.storage.ListAccountsByState(StateNormal)
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 {
		err := pool.refreshAccounts()
		if err != nil {
			return nil, err
		}
	}

	accounts, err = pool.storage.ListAccountsByState(StateNormal)
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, ErrNoAvailableAccount
	}
	for _, acc := range accounts {
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
	accounts, err := pool.storage.ListAccounts()
	if err != nil {
		return err
	}
	for _, acc := range accounts {
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
		restrictedTime := utils.TimeToMidnightTime(acc.GetRestrictedTime())
		if now.Sub(restrictedTime).Hours() < 24 {
			return nil
		}
		pool.logger.Info().Str("username", acc.Username).Msg("account recover from Restricted state")
	} else {
		return nil
	}
	acc.State = StateNormal
	return pool.storage.UpdateAccount(*acc)
}

func (pool *Pool) setAccountRestricted(acc *Account, newState string) {
	updateAcc := *acc
	updateAcc.State = newState
	updateAcc.RestrictedTime = time.Now().Unix()
	if err := pool.storage.UpdateAccount(updateAcc); err != nil {
		pool.logger.Error().Err(err).Msg("update account error")
	}
}

func (pool *Pool) recycleAllAccStorage() {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	accounts, err := pool.storage.ListAccounts()
	if err != nil {
		pool.logger.Error().Err(err).Msg("recycle all accounts")
		return
	}
	for _, acc := range accounts {
		client, err := pool.getClientByAcc(&acc)
		if err == nil {
			_ = pool.recycleAccStorage(client)
		}
	}
}

func (pool *Pool) recycleAccStorage(client *pikpakgo.PikPakClient) error {
	err := client.OfflineRemoveAll([]string{pikpakgo.PhaseTypeError}, true)
	if err != nil {
		return err
	}
	fileId, err := client.FolderPathToID(pool.tempDirectory, true)
	if err != nil {
		return err
	}
	files, err := client.FileListAll(fileId)
	if err != nil {
		return err
	}
	var ids []string
	for _, fi := range files {
		if fi.FolderType == pikpakgo.FolderTypeDownload {
			continue
		}
		ids = append(ids, fi.ID)
	}
	return client.BatchDeleteFiles(ids)
}
