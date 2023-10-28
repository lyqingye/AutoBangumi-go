package cache

import (
	"autobangumi-go/mdb"
	"autobangumi-go/rss/mikan"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/pkg/errors"
)

type InMemoryCacheManager struct {
	cacheParseItemResult          map[string]mikan.ParseItemResult
	cacheMikanBangumiToBangumiTV  map[string]int64
	cacheTMDBByID                 map[int64]tmdb.TVDetails
	cacheTMDBByKeyWord            map[string]tmdb.TVDetails
	cacheBangumiTVSubjectsByID    map[int64]mdb.Subjects
	cacheBangumiTVSubjectsByTitle map[string]mdb.Subjects
}

func (imc *InMemoryCacheManager) GetBangumiTVCache(title string) (mdb.Subjects, error) {
	rs, cached := imc.cacheBangumiTVSubjectsByTitle[title]
	if cached {
		return rs, nil
	}
	return mdb.Subjects{}, errors.New("not found")
}

func (imc *InMemoryCacheManager) StoreBangumiTVCache(title string, subjects mdb.Subjects) error {
	imc.cacheBangumiTVSubjectsByTitle[title] = subjects
	return nil
}

func (imc *InMemoryCacheManager) GetBangumiTVSubjectsCache(id int64) (mdb.Subjects, error) {
	rs, cached := imc.cacheBangumiTVSubjectsByID[id]
	if cached {
		return rs, nil
	}
	return mdb.Subjects{}, errors.New("not found")
}

func (imc *InMemoryCacheManager) StoreBangumiTVSubjectsCache(id int64, subjects mdb.Subjects) error {
	imc.cacheBangumiTVSubjectsByID[id] = subjects
	return nil
}

func (imc *InMemoryCacheManager) GetTMDBCache(keyword string) (tmdb.TVDetails, error) {
	rs, cached := imc.cacheTMDBByKeyWord[keyword]
	if cached {
		return rs, nil
	}
	return tmdb.TVDetails{}, errors.New("not found")
}

func (imc *InMemoryCacheManager) StoreTMDBCache(keyword string, detail tmdb.TVDetails) error {
	imc.cacheTMDBByKeyWord[keyword] = detail
	return nil
}

func (imc *InMemoryCacheManager) GetTMDBCacheByID(id int64) (tmdb.TVDetails, error) {
	rs, cached := imc.cacheTMDBByID[id]
	if cached {
		return rs, nil
	}
	return tmdb.TVDetails{}, errors.New("not found")
}

func (imc *InMemoryCacheManager) StoreTMDBCacheById(id int64, detail tmdb.TVDetails) error {
	imc.cacheTMDBByID[id] = detail
	return nil
}

func NewInMemoryCacheManager() mikan.CacheManager {
	return &InMemoryCacheManager{
		cacheParseItemResult:          make(map[string]mikan.ParseItemResult),
		cacheMikanBangumiToBangumiTV:  make(map[string]int64),
		cacheBangumiTVSubjectsByTitle: make(map[string]mdb.Subjects),
		cacheBangumiTVSubjectsByID:    make(map[int64]mdb.Subjects),
		cacheTMDBByKeyWord:            make(map[string]tmdb.TVDetails),
		cacheTMDBByID:                 make(map[int64]tmdb.TVDetails),
	}
}

func (imc *InMemoryCacheManager) GetParseCache(itemLink string) (mikan.ParseItemResult, error) {
	rs, cached := imc.cacheParseItemResult[itemLink]
	if cached {
		return rs, nil
	}
	return mikan.ParseItemResult{}, errors.New("not found")
}

func (imc *InMemoryCacheManager) StoreParseCache(itemLink string, cache mikan.ParseItemResult) error {
	imc.cacheParseItemResult[itemLink] = cache
	return nil
}

func (imc *InMemoryCacheManager) StoreMikanBangumiToBangumiTV(mikanBangumiID string, bangumiTVID int64) error {
	imc.cacheMikanBangumiToBangumiTV[mikanBangumiID] = bangumiTVID
	return nil
}

func (imc *InMemoryCacheManager) GetMikanBangumiToBangumiTV(mikanBangumiID string) (int64, error) {
	rs, cached := imc.cacheMikanBangumiToBangumiTV[mikanBangumiID]
	if cached {
		return rs, nil
	}
	return 0, errors.New("not found")
}

func (imc *InMemoryCacheManager) Close() error {
	return nil
}
