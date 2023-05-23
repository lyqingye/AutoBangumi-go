package mikan

import (
	"autobangumi-go/mdb"
	"strconv"
)

var (
	KeyParseCacheByLink = []byte("parse-cache-link")
)

var (
	KeyBangumiTVCacheBySubjectId    = []byte("bangumiTV-cache-subject-id")
	KeyBangumiTVCacheByKeyword      = []byte("bangumiTV-cache-keyword")
	KeyTMDBCacheByKeyword           = []byte("TMDB-cache-keyword")
	KeyMikanBangumiToBangumiTvCache = []byte("mikan-bangumi-to-bangumi-tv-cache")
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

func getMikanBangumiToBangumiTVCache(mikanBangumiId string) []byte {
	return append(KeyMikanBangumiToBangumiTvCache, []byte(mikanBangumiId)...)
}

func (parser *MikanRSSParser) getParseCache(itemLink string) (*ParseItemResult, bool) {
	cache := ParseItemResult{}
	found, err := parser.db.Get(getParseCacheKeyByLink(itemLink), &cache)
	if err != nil {
		return nil, false
	}
	return &cache, found
}

func (parser *MikanRSSParser) storeParseCache(itemLink string, cache *ParseItemResult) {
	err := parser.db.Set(getParseCacheKeyByLink(itemLink), cache)
	if err != nil {
		parser.logger.Err(err).Msg("store parse cache error")
	}
}

func (parser *MikanRSSParser) getBangumiTVSubject(subjectId int64) (*mdb.Subjects, error) {
	cachedSubject := mdb.Subjects{}
	key := getBangumiTVCacheKeyBySubjectId(subjectId)
	cached, err := parser.db.Get(key, &cachedSubject)
	if err != nil || !cached {
		subject, err := parser.bangumiTvClient.GetSubjects(subjectId)
		if err != nil {
			return nil, err
		}
		return subject, parser.db.Set(key, subject)
	} else {
		return &cachedSubject, nil
	}
}
