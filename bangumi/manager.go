package bangumi

import (
	"autobangumi-go/bus"
	"autobangumi-go/utils"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
)

type BangumiManager struct {
	home        string
	inComplete  map[string]Bangumi
	complete    map[string]Bangumi
	rwLock      sync.RWMutex
	logger      zerolog.Logger
	eb          *bus.EventBus
	episodeLock map[string]sync.Mutex
}

func NewBangumiManager(home string, eb *bus.EventBus) (*BangumiManager, error) {
	_, err := os.Stat(home)
	if os.IsNotExist(err) {
		err = os.MkdirAll(home, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	man := BangumiManager{
		home:       home,
		complete:   make(map[string]Bangumi),
		inComplete: make(map[string]Bangumi),
		rwLock:     sync.RWMutex{},
		logger:     utils.GetLogger("bangumiMan"),
		eb:         eb,
	}
	err = man.init()
	if err != nil {
		return nil, err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = watcher.Add(home)
	if err != nil {
		return nil, err
	}
	go man.watchNewBangumiFile(watcher)
	return &man, nil
}

func (man *BangumiManager) init() error {
	man.rwLock.Lock()
	defer man.rwLock.Unlock()
	return filepath.WalkDir(man.home, func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(path, ".json") {
			bz, err := os.ReadFile(path)
			var bangumi Bangumi
			if err == nil {
				err = json.Unmarshal(bz, &bangumi)
				if err == nil {
					if bangumi.IsComplete() {
						man.complete[bangumi.Info.Title] = bangumi
						man.logger.Debug().Str("title", bangumi.Info.Title).Msg("load complete bangumi")
					} else {
						man.inComplete[bangumi.Info.Title] = bangumi
						man.logger.Debug().Str("title", bangumi.Info.Title).Msg("load inComplete bangumi")
					}
				} else {
					man.logger.Error().Err(err).Str("file", path).Msg("failed to load bangumi")
				}
			} else {
				man.logger.Error().Err(err).Str("file", path).Msg("failed to load bangumi")
			}
		}
		return nil
	})
}

func (man *BangumiManager) IsBangumiExist(title string) bool {
	man.rwLock.RLock()
	defer man.rwLock.RUnlock()
	_, found := man.inComplete[title]
	if found {
		return true
	}
	_, found = man.complete[title]
	return found
}

func (man *BangumiManager) AddBangumiIfNotExist(bangumi Bangumi) {
	man.rwLock.Lock()
	defer man.rwLock.Unlock()
	if _, found := man.inComplete[bangumi.Info.Title]; found {
		return
	}
	man.inComplete[bangumi.Info.Title] = bangumi

	// publish event
	man.eb.Publish(bus.BangumiManTopic, bus.Event{
		EventType: bus.BangumiManAddNewEvent,
		Inner:     bangumi,
	})
	man.logger.Info().Str("title", bangumi.Info.Title).Msg("load new bangumi")
}

func (man *BangumiManager) watchNewBangumiFile(watcher *fsnotify.Watcher) {
	man.logger.Info().Msg("start bangumi file watcher")
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op == fsnotify.Create {
				if !strings.HasSuffix(event.Name, ".json") {
					continue
				}
				filename := strings.ReplaceAll(event.Name, ".json", "")
				if _, found := man.inComplete[filename]; found {
					continue
				}
				time.Sleep(5 * time.Second)
				bz, err := os.ReadFile(event.Name)
				if err != nil {
					man.logger.Error().Err(err).Msg("can't not be load bangumi file")
				} else {
					bangumi := Bangumi{}
					if err = json.Unmarshal(bz, &bangumi); err == nil {
						man.AddBangumiIfNotExist(bangumi)
					} else {
						man.logger.Error().Err(err).Msg("can't not be load bangumi file")
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				_ = watcher.Close()
				return
			}
			man.logger.Error().Err(err).Msg("watcher err")
		}
	}
}

func (man *BangumiManager) MarkEpisodeComplete(info *BangumiInfo, seasonNum uint, episode Episode) {
	man.rwLock.Lock()
	defer man.rwLock.Unlock()
	if bangumi, found := man.inComplete[info.Title]; found {
		if season, foundSeason := bangumi.Seasons[seasonNum]; foundSeason {
			if !season.IsComplete(episode.Number) {
				season.Complete = append(season.Complete, episode.Number)
				bangumi.Seasons[seasonNum] = season
				_ = man.Flush(&bangumi)
			}
		}
		if bangumi.IsComplete() {
			delete(man.inComplete, info.Title)
			man.complete[info.Title] = bangumi
		} else {
			man.inComplete[info.Title] = bangumi
		}
	}
}

func (man *BangumiManager) DownloaderTouchEpisode(info *BangumiInfo, seasonNum uint, episode Episode, downloader DownloadState) {
	man.rwLock.Lock()
	defer man.rwLock.Unlock()
	if bangumi, found := man.inComplete[info.Title]; found {
		if season, foundSeason := bangumi.Seasons[seasonNum]; foundSeason {
			for i, ep := range season.Episodes {
				if ep.Number == episode.Number {
					season.Episodes[i].DownloadState = downloader
				}
			}
			bangumi.Seasons[seasonNum] = season
			_ = man.Flush(&bangumi)
		}
		man.inComplete[info.Title] = bangumi
	}
}

func (man *BangumiManager) IterInCompleteBangumi(fn func(man *BangumiManager, bangumi *Bangumi) bool) {
	man.rwLock.Lock()
	defer man.rwLock.Unlock()
	for title, bangumi := range man.inComplete {
		result := fn(man, &bangumi)
		man.inComplete[title] = bangumi
		if result {
			break
		}
	}
}

func (man *BangumiManager) GetAndLockInCompleteBangumi(title string, fn func(man *BangumiManager, bangumi *Bangumi)) {
	man.rwLock.Lock()
	defer man.rwLock.Unlock()
	if bangumi, ok := man.inComplete[title]; ok {
		fn(man, &bangumi)
		man.inComplete[title] = bangumi
	}
}

func (man *BangumiManager) Flush(bangumi *Bangumi) error {
	bz, err := json.MarshalIndent(bangumi, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(man.home, fmt.Sprintf("%s.json", bangumi.Info.Title)), bz, os.ModePerm)
}
