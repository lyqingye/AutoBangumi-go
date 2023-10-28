package mdb

import (
	"fmt"

	tmdb "github.com/cyruzin/golang-tmdb"
)

var (
	TMDBJPLangOptions = map[string]string{
		"language": "ja-JP",
	}
	TMDBZHLangOptions = map[string]string{
		"language": "zh-CN",
	}
	TMDBENLangOptions = map[string]string{
		"language": "en-US",
	}
)

type TMDBClient struct {
	inner *tmdb.Client
}

func NewTMDBClient(token string) (*TMDBClient, error) {
	inner, err := tmdb.Init(token)
	if err != nil {
		return nil, err
	}

	client := TMDBClient{
		inner: inner,
	}
	return &client, nil
}

func (client *TMDBClient) SearchTVShowByKeyword(keyword string) (*tmdb.TVDetails, error) {
	for _, opts := range []map[string]string{TMDBZHLangOptions, TMDBJPLangOptions, TMDBENLangOptions} {
		searchResult, err := client.inner.GetSearchTVShow(keyword, opts)
		if err != nil {
			return nil, err
		}
		if len(searchResult.Results) > 0 {
			tvDetails, err := client.inner.GetTVDetails(int(searchResult.Results[0].ID), opts)
			if err == nil {
				return tvDetails, nil
			} else {
				return nil, err
			}
		}
	}
	return nil, fmt.Errorf("tmdb not found: %s", keyword)
}

func (client *TMDBClient) GetTVDetailById(tmdbId int64) (*tmdb.TVDetails, error) {
	for _, opts := range []map[string]string{TMDBZHLangOptions, TMDBJPLangOptions, TMDBENLangOptions} {
		tvDetails, err := client.inner.GetTVDetails(int(tmdbId), opts)
		if err == nil {
			return tvDetails, nil
		} else {
			return nil, err
		}
	}
	return nil, fmt.Errorf("tmdb not found: %d", tmdbId)
}
