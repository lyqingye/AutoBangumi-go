package db

import (
	"gorm.io/driver/postgres"
	//"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
)

type Backend struct {
	inner *gorm.DB
	mtx   sync.RWMutex
}

func NewBackend(dsn string) (*Backend, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", 0), logger.Config{LogLevel: logger.Info}),
	})
	if err != nil {
		return nil, err
	}
	backend := Backend{
		inner: db,
		mtx:   sync.RWMutex{},
	}
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&MBangumi{}, &MSeason{}, &MEpisode{}, &MEpisodeTorrent{}, &MTorrent{})
	return &backend, err
}

func (b *Backend) AddOrUpdateBangumi(bangumi *MBangumi) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	return b.inner.Create(bangumi).Error
}

func (b *Backend) ListBangumis(fn func(bgm *MBangumi) bool) error {
	b.mtx.RLock()
	b.mtx.RUnlock()
	var bangumis []MBangumi
	if err := b.inner.Preload("Seasons").Find(&bangumis).Error; err != nil {
		return err
	}
	for _, bgm := range bangumis {
		if fn(&bgm) {
			break
		}
	}
	return nil
}

func (b *Backend) ListIncompleteBangumi(fn func(bgm *MBangumi) bool) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	var bangumis []MBangumi
	result := b.inner.Joins("JOIN m_seasons ON m_bangumis.id = m_seasons.bangumi_id").
		Where("m_seasons.state = ?", SeasonStateIncomplete).
		Find(&bangumis)
	if result.Error != nil {
		return result.Error
	}
	for _, bgm := range bangumis {
		if fn(&bgm) {
			break
		}
	}
	return nil
}
