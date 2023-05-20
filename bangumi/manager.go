package bangumi

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"pikpak-bot/utils"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

type BangumiManager struct {
	home       string
	inComplete map[string]Bangumi
	complete   map[string]Bangumi
	rwLock     sync.RWMutex
	logger     zerolog.Logger
}

func NewBangumiManager(home string) (*BangumiManager, error) {
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
	}
	return &man, man.init()
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

func (man *BangumiManager) Flush(bangumi *Bangumi) error {
	bz, err := json.MarshalIndent(bangumi, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(man.home, fmt.Sprintf("%s.json", bangumi.Info.Title)), bz, os.ModePerm)
}
