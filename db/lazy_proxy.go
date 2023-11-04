package db

import (
	"errors"

	"autobangumi-go/bangumi"
	"gorm.io/gorm"
)

type ProxyMBangumi struct {
	MBangumi
	db *gorm.DB
}

func Proxy(inner MBangumi, db *gorm.DB) bangumi.Bangumi {
	return &ProxyMBangumi{
		MBangumi: inner,
		db:       db,
	}
}

func (bgm *ProxyMBangumi) GetSeasons() ([]bangumi.Season, error) {
	var seasons []MSeason
	if err := bgm.db.Where("bangumi_id = ?", bgm.ID).Find(&[]MSeason{}).Find(&seasons).Error; err != nil {
		return nil, err
	}
	var ret []bangumi.Season
	for _, season := range seasons {
		proxy := ProxyMSeason{
			MSeason: season,
			db:      bgm.db,
		}
		ret = append(ret, &proxy)
	}
	return ret, nil
}

func (bgm *ProxyMBangumi) GetTitle() string {
	return bgm.Title
}

func (bgm *ProxyMBangumi) GetTmDBId() int64 {
	return bgm.TMDBId
}

type ProxyMSeason struct {
	MSeason
	db *gorm.DB
}

func (s *ProxyMSeason) GetEpisodes() ([]bangumi.Episode, error) {
	var ret []bangumi.Episode
	episodes, err := s.getEpisodes()
	if err != nil {
		return nil, err
	}
	for _, episode := range episodes {
		proxy := ProxyEpisode{
			MEpisode: episode,
			db:       s.db,
		}
		ret = append(ret, &proxy)
	}
	return ret, nil
}

func (s *ProxyMSeason) getEpisodes() ([]MEpisode, error) {
	var episodes []MEpisode
	return episodes, s.db.Where("season_id = ?", s.ID).Find(&episodes).Error
}

func (s *ProxyMSeason) GetRefBangumi() (bangumi.Bangumi, error) {
	var ret MBangumi
	err := s.db.Where("id = ?", s.BangumiId).Find(&ret).Error
	return &ProxyMBangumi{
		MBangumi: ret,
		db:       s.db,
	}, err
}

type ProxyEpisode struct {
	MEpisode
	db *gorm.DB
}

func (s *ProxyEpisode) GetResources() ([]bangumi.Resource, error) {
	var resources []MEpisodeTorrent
	var ret []bangumi.Resource
	err := s.db.Select([]string{"id", "episode_id", "torrent_hash", "file_indexes", "subtitle_lang", "resolution", "resource_type", "valid"}).Where("episode_id = ?", s.MEpisode.ID).Find(&resources).Error
	if err != nil {
		return nil, err
	}
	for _, r := range resources {
		ret = append(ret, &ProxyResource{
			MEpisodeTorrent: r,
			db:              s.db,
		})
	}
	return ret, nil
}

func (s *ProxyEpisode) GetRefSeason() (bangumi.Season, error) {
	var ret MSeason
	err := s.db.Where("id = ?", s.SeasonId).Find(&ret).Error
	return &ProxyMSeason{
		MSeason: ret,
		db:      s.db,
	}, err
}

type ProxyResource struct {
	MEpisodeTorrent
	db *gorm.DB
}

func (p *ProxyResource) GetRefEpisode() (bangumi.Episode, error) {
	var episode MEpisode
	err := p.db.Where("id = ?", p.EpisodeId).First(&episode).Error
	if err != nil {
		if errors.Is(gorm.ErrRecordNotFound, err) {
			return nil, nil
		}
		return nil, err
	}
	return &ProxyEpisode{
		MEpisode: episode,
		db:       p.db,
	}, nil
}

func (p *ProxyResource) GetTorrent() []byte {
	ret := MEpisodeTorrent{}
	err := p.db.Where("id = ?", p.ID).First(&ret).Error
	if errors.Is(gorm.ErrRecordNotFound, err) {
		return nil
	}
	return ret.Bz
}

type ProxyEpisodeDownloadHistory struct {
	MEpisodeDownloadHistory
	db *gorm.DB
}

func (history *ProxyEpisodeDownloadHistory) GetState() bangumi.DownloadState {
	return bangumi.DownloadState(history.State)
}

func (history *ProxyEpisodeDownloadHistory) GetDownloader() bangumi.Downloader {
	return bangumi.Downloader(history.Downloader)
}

func (history *ProxyEpisodeDownloadHistory) GetDownloaderContext() string {
	return history.Context
}

func (history *ProxyEpisodeDownloadHistory) GetErrMsg() string {
	return history.ErrorMsg
}

func (history *ProxyEpisodeDownloadHistory) GetResourcesIds() string {
	return history.ResourcesIds
}

func (history *ProxyEpisodeDownloadHistory) GetRetryCount() int64 {
	return history.RetryCount
}

func (history *ProxyEpisodeDownloadHistory) IncRetryCount() {
	history.RetryCount = history.RetryCount + 1
}

func (history *ProxyEpisodeDownloadHistory) SetDownloader(downloader bangumi.Downloader, context string, downloadState bangumi.DownloadState, error error) {
	history.Downloader = string(downloader)
	history.Context = context
	history.State = string(downloadState)
	if error != nil {
		history.ErrorMsg = error.Error()
	}
}

func (history *ProxyEpisodeDownloadHistory) SetDownloadState(downloadState bangumi.DownloadState, error error) {
	history.State = string(downloadState)
	if error != nil {
		history.ErrorMsg = error.Error()
	}
}

func (history *ProxyEpisodeDownloadHistory) GetRefEpisode() (bangumi.Episode, error) {
	var episode MEpisode
	err := history.db.Where("id = ?", history.EpisodeId).First(&episode).Error
	if err != nil {
		if errors.Is(gorm.ErrRecordNotFound, err) {
			return nil, nil
		}
		return nil, err
	}
	return &ProxyEpisode{
		MEpisode: episode,
		db:       history.db,
	}, nil
}
