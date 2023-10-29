package mdb

import (
	"fmt"

	"github.com/antlabs/strsim"
	tmdb "github.com/cyruzin/golang-tmdb"
	"golang.org/x/exp/maps"
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

		resultMap := map[string]*tmdb.TVDetails{}
		if len(searchResult.Results) > 0 {
			for _, rs := range searchResult.Results {
				tvDetails, err := client.inner.GetTVDetails(int(rs.ID), opts)
				if err != nil {
					return nil, err
				}
				resultMap[tvDetails.Name] = tvDetails
			}
			if len(resultMap) > 0 {
				bestMatch := strsim.FindBestMatch(keyword, maps.Keys(resultMap))
				return resultMap[bestMatch.Match.S], nil
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
