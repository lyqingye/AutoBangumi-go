package cache

import (
	"strconv"

	"autobangumi-go/mdb"
	"autobangumi-go/rss/mikan"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/pkg/errors"
)

var ErrCacheNotFound = errors.New("cache not found")

type KVStorage interface {
	Get(key []byte, value interface{}) (bool, error)
	Set(key []byte, value interface{}) error
	Close() error
}

type KVCacheManager struct {
	db KVStorage
}

func NewKVCacheManager(store KVStorage) mikan.CacheManager {
	return &KVCacheManager{db: store}
}

var (
	KeyParseCacheByLink             = []byte{0x1}
	KeyBangumiTVCacheBySubjectId    = []byte{0x2}
	KeyBangumiTVCacheByKeyword      = []byte{0x3}
	KeyTMDBCacheByKeyword           = []byte{0x4}
	KeyTMDBCacheByID                = []byte{0x5}
	KeyMikanBangumiToBangumiTvCache = []byte{0x6}
)

func getParseCacheKeyByLink(link string) []byte {
	return append(KeyParseCacheByLink, []byte(link)...)
}

func getBangumiTVCacheKeyBySubjectId(subject int64) []byte {
	return append(KeyBangumiTVCacheBySubjectId, []byte(strconv.FormatInt(subject, 10))...)
}

func getBangumiTVCacheKeyByKeyword(keyword string) []byte {
	return append(KeyBangumiTVCacheByKeyword, []byte(keyword)...)
}

func getTMDBCacheByKeyword(keyword string) []byte {
	return append(KeyTMDBCacheByKeyword, []byte(keyword)...)
}

func getTMDBCacheByID(id int64) []byte {
	return append(KeyTMDBCacheByID, []byte(strconv.FormatInt(id, 10))...)
}

func getMikanBangumiToBangumiTVCache(mikanBangumiId string) []byte {
	return append(KeyMikanBangumiToBangumiTvCache, []byte(mikanBangumiId)...)
}

func (cm *KVCacheManager) GetParseCache(itemLink string) (mikan.ParseItemResult, error) {
	cache := mikan.ParseItemResult{}
	found, err := cm.db.Get(getParseCacheKeyByLink(itemLink), &cache)
	if err != nil {
		return cache, err
	}
	if !found {
		return cache, ErrCacheNotFound
	}
	return cache, nil
}

func (cm *KVCacheManager) StoreParseCache(itemLink string, cache mikan.ParseItemResult) error {
	return cm.db.Set(getParseCacheKeyByLink(itemLink), &cache)
}

func (cm *KVCacheManager) StoreMikanBangumiToBangumiTV(mikanBangumiID string, bangumiTVID int64) error {
	return cm.db.Set(getMikanBangumiToBangumiTVCache(mikanBangumiID), &bangumiTVID)
}

func (cm *KVCacheManager) GetMikanBangumiToBangumiTV(mikanBangumiID string) (int64, error) {
	var ret int64
	found, err := cm.db.Get(getMikanBangumiToBangumiTVCache(mikanBangumiID), &ret)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, ErrCacheNotFound
	}
	return ret, nil
}

func (cm *KVCacheManager) GetBangumiTVCache(title string) (mdb.Subjects, error) {
	var ret mdb.Subjects
	found, err := cm.db.Get(getBangumiTVCacheKeyByKeyword(title), &ret)
	if err != nil {
		return ret, err
	}
	if !found {
		return ret, ErrCacheNotFound
	}
	return ret, nil
}

func (cm *KVCacheManager) StoreBangumiTVCache(title string, subjects mdb.Subjects) error {
	return cm.db.Set(getBangumiTVCacheKeyByKeyword(title), &subjects)
}

func (cm *KVCacheManager) GetBangumiTVSubjectsCache(id int64) (mdb.Subjects, error) {
	var ret mdb.Subjects
	found, err := cm.db.Get(getBangumiTVCacheKeyBySubjectId(id), &ret)
	if err != nil {
		return ret, err
	}
	if !found {
		return ret, ErrCacheNotFound
	}
	return ret, nil
}

func (cm *KVCacheManager) StoreBangumiTVSubjectsCache(id int64, subjects mdb.Subjects) error {
	return cm.db.Set(getBangumiTVCacheKeyBySubjectId(id), &subjects)
}

func (cm *KVCacheManager) GetTMDBCache(keyword string) (tmdb.TVDetails, error) {
	var ret tmdb.TVDetails
	found, err := cm.db.Get(getTMDBCacheByKeyword(keyword), &ret)
	if err != nil {
		return ret, err
	}
	if !found {
		return ret, ErrCacheNotFound
	}
	return ret, nil
}

func (cm *KVCacheManager) StoreTMDBCache(keyword string, detail tmdb.TVDetails) error {
	return cm.db.Set(getTMDBCacheByKeyword(keyword), &detail)
}

func (cm *KVCacheManager) GetTMDBCacheByID(id int64) (tmdb.TVDetails, error) {
	var ret tmdb.TVDetails
	found, err := cm.db.Get(getTMDBCacheByID(id), &ret)
	if err != nil {
		return ret, err
	}
	if !found {
		return ret, ErrCacheNotFound
	}
	return ret, nil
}

func (cm *KVCacheManager) StoreTMDBCacheById(id int64, detail tmdb.TVDetails) error {
	return cm.db.Set(getTMDBCacheByID(id), &detail)
}

func (cm *KVCacheManager) Close() error {
	return cm.db.Close()
}
