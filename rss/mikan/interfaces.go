package mikan

import (
	"autobangumi-go/mdb"
	tmdb "github.com/cyruzin/golang-tmdb"
)

type TMDB interface {
	SearchTVShowByKeyword(keyword string) (*tmdb.TVDetails, error)
	GetTVDetailById(tmdbId int64) (*tmdb.TVDetails, error)
}

type BangumiTV interface {
	SearchAnime2(keyword string) (*mdb.Subjects, error)
	GetSubjects(id int64) (*mdb.Subjects, error)
}

type CacheManager interface {
	GetParseCache(itemLink string) (ParseItemResult, error)
	StoreParseCache(itemLink string, cache ParseItemResult) error
	StoreMikanBangumiToBangumiTV(mikanBangumiID string, bangumiTVID int64) error
	GetMikanBangumiToBangumiTV(mikanBangumiID string) (int64, error)

	GetBangumiTVCache(title string) (mdb.Subjects, error)
	StoreBangumiTVCache(title string, subjects mdb.Subjects) error
	GetBangumiTVSubjectsCache(id int64) (mdb.Subjects, error)
	StoreBangumiTVSubjectsCache(id int64, subjects mdb.Subjects) error

	GetTMDBCache(keyword string) (tmdb.TVDetails, error)
	StoreTMDBCache(keyword string, detail tmdb.TVDetails) error

	GetTMDBCacheByID(id int64) (tmdb.TVDetails, error)
	StoreTMDBCacheById(id int64, detail tmdb.TVDetails) error

	Close() error
}
