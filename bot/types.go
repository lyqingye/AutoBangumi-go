package bot

import (
	"autobangumi-go/bangumi"
)

type BangumiFromFs struct {
	title   string
	tmdbID  int64
	seasons map[uint]SeasonFromFs
}

func NewBangumiFromFs(title string, tmdbID int64) BangumiFromFs {
	return BangumiFromFs{
		title:   title,
		tmdbID:  tmdbID,
		seasons: make(map[uint]SeasonFromFs),
	}
}

func (b *BangumiFromFs) AddDownloadEpisode(seasonNum uint, epCount uint, epNum uint) {
	season, found := b.seasons[seasonNum]
	if !found {
		season = SeasonFromFs{
			number:   seasonNum,
			epCount:  epCount,
			episodes: make(map[uint]EpisodeFromFs),
		}
	}
	season.episodes[epNum] = EpisodeFromFs{epNum}
	b.seasons[seasonNum] = season
}

func (b *BangumiFromFs) GetTitle() string {
	return b.title
}

func (b *BangumiFromFs) GetTmDBId() int64 {
	return b.tmdbID
}

func (b *BangumiFromFs) GetSeasons() ([]bangumi.Season, error) {
	var ret []bangumi.Season
	for _, season := range b.seasons {
		copyValue := season
		ret = append(ret, &copyValue)
	}
	return ret, nil
}

func (b *BangumiFromFs) IsDownloaded() bool {
	for _, season := range b.seasons {
		if !season.IsDownloaded() {
			return false
		}
	}
	return true
}

type SeasonFromFs struct {
	number   uint
	epCount  uint
	episodes map[uint]EpisodeFromFs
}

func (s SeasonFromFs) GetNumber() uint {
	return s.number
}

func (s SeasonFromFs) GetEpCount() uint {
	return s.epCount
}

func (s SeasonFromFs) GetEpisodes() ([]bangumi.Episode, error) {
	var ret []bangumi.Episode
	for _, ep := range s.episodes {
		copyValue := ep
		ret = append(ret, &copyValue)
	}
	return ret, nil
}

func (s SeasonFromFs) GetRefBangumi() (bangumi.Bangumi, error) {
	panic("implement me")
}

func (s SeasonFromFs) IsDownloaded() bool {
	return s.epCount <= uint(len(s.episodes))
}

type EpisodeFromFs struct {
	number uint
}

func (e EpisodeFromFs) GetNumber() uint {
	return e.number
}

func (e EpisodeFromFs) GetResources() ([]bangumi.Resource, error) {
	return nil, nil
}

func (e EpisodeFromFs) GetRefSeason() (bangumi.Season, error) {
	panic("implement me")
}

func (e EpisodeFromFs) IsDownloaded() bool {
	return true
}
