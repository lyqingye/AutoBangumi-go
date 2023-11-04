package db

import (
	"strings"
	"time"

	"autobangumi-go/bangumi"
)

type MBangumi struct {
	ID         uint   `gorm:"primaryKey"`
	Title      string `gorm:"uniqueIndex"`
	TMDBId     int64  `gorm:"uniqueIndex:bangumi_tmdb_idx"`
	Downloaded bool
	UpdatedAt  time.Time
}

func (bgm *MBangumi) GetTitle() string {
	return bgm.Title
}

func (bgm *MBangumi) GetTmDBId() int64 {
	return bgm.TMDBId
}

func (bgm *MBangumi) GetSeasons() ([]bangumi.Season, error) {
	panic("unsupported")
}

func (s *MBangumi) IsDownloaded() bool {
	return s.Downloaded
}

const (
	SeasonStateIncomplete = "incomplete"
	SeasonStateComplete   = "complete"
)

type MSeason struct {
	ID         uint `gorm:"primaryKey"`
	BangumiId  uint `gorm:"uniqueIndex:bangumi_number_idx"`
	Number     uint `gorm:"uniqueIndex:bangumi_number_idx"`
	EpCount    uint
	Downloaded bool
}

func (s *MSeason) GetNumber() uint {
	return s.Number
}

func (s *MSeason) GetEpCount() uint {
	return s.EpCount
}

func (s *MSeason) GetEpisodes() ([]bangumi.Episode, error) {
	panic("unsupported")
}

func (s *MSeason) GetRefBangumi() (bangumi.Bangumi, error) {
	panic("unsupported")
}

func (s *MSeason) IsDownloaded() bool {
	return s.Downloaded
}

type MEpisodeTorrent struct {
	ID           uint   `gorm:"primaryKey"`
	TorrentHash  string `gorm:"uniqueIndex:torrent_hash"`
	EpisodeId    uint
	FileIndexes  string
	Bz           []byte
	SubtitleLang string
	Resolution   bangumi.Resolution
	ResourceType bangumi.ResourceType
	Valid        bool `gorm:"default:true"`
}

func (t *MEpisodeTorrent) GetTorrent() []byte {
	return t.Bz
}

func (t *MEpisodeTorrent) GetTorrentHash() string {
	return t.TorrentHash
}

func (t *MEpisodeTorrent) GetSubtitleLang() []bangumi.SubtitleLang {
	var ret []bangumi.SubtitleLang
	for _, l := range strings.Split(t.SubtitleLang, ",") {
		ret = append(ret, bangumi.SubtitleLang(l))
	}
	return ret
}

func (t *MEpisodeTorrent) GetResolution() bangumi.Resolution {
	return t.Resolution
}

func (t *MEpisodeTorrent) GetResourceType() bangumi.ResourceType {
	return t.ResourceType
}

func (t *MEpisodeTorrent) SetSubtitleLang(lang []bangumi.SubtitleLang) {
	var newLang []string
	for _, l := range lang {
		newLang = append(newLang, string(l))
	}
	t.SubtitleLang = strings.Join(newLang, ",")
}

func (t *MEpisodeTorrent) GetRefEpisode() (bangumi.Episode, error) {
	panic("unsupported")
}

type MEpisode struct {
	ID         uint `gorm:"primaryKey"`
	SeasonId   uint `gorm:"uniqueIndex:bangumi_season_idx"`
	Number     uint `gorm:"uniqueIndex:bangumi_season_idx"`
	Downloaded bool
}

func (ep *MEpisode) GetNumber() uint {
	return ep.Number
}

func (ep *MEpisode) GetResources() ([]bangumi.Resource, error) {
	panic("unsupported")
}

func (s *MEpisode) GetRefSeason() (bangumi.Season, error) {
	panic("unsupported")
}

func (s *MEpisode) IsDownloaded() bool {
	return s.Downloaded
}

const (
	DownloadStateFinished    = "finished"
	DownloadStateDownloading = "downloading"
	DownloadStateError       = "error"
)

type MEpisodeDownloadHistory struct {
	EpisodeId    uint `gorm:"primarykey"`
	ResourcesIds string
	Downloader   string
	Context      string // json value use by downloader
	State        string `gorm:"index:episode_index"` // ref DownloadStateFinished | DownloadStateDownloading | DownloadStateError
	ErrorMsg     string
	RetryCount   int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (history *MEpisodeDownloadHistory) GetState() bangumi.DownloadState {
	return bangumi.DownloadState(history.State)
}

func (history *MEpisodeDownloadHistory) GetDownloader() bangumi.Downloader {
	return bangumi.Downloader(history.Downloader)
}

func (history *MEpisodeDownloadHistory) GetDownloaderContext() string {
	return history.Context
}

func (history *MEpisodeDownloadHistory) GetErrMsg() string {
	return history.ErrorMsg
}

func (history *MEpisodeDownloadHistory) GetResourcesIds() string {
	return history.ResourcesIds
}

func (history *MEpisodeDownloadHistory) GetRetryCount() int64 {
	return history.RetryCount
}

func (history *MEpisodeDownloadHistory) IncRetryCount() {
	history.RetryCount = history.RetryCount + 1
}

func (history *MEpisodeDownloadHistory) SetDownloader(downloader bangumi.Downloader, context string, downloadState bangumi.DownloadState, error error) {
	history.Downloader = string(downloader)
	history.Context = context
	history.State = string(downloadState)
	history.ErrorMsg = error.Error()
}

func (history *MEpisodeDownloadHistory) SetDownloadState(downloadState bangumi.DownloadState, error error) {
	history.State = string(downloadState)
	if error != nil {
		history.ErrorMsg = error.Error()
	}
}

func (history *MEpisodeDownloadHistory) LastUpdatedTime() time.Time {
	return history.UpdatedAt
}

func (history *MEpisodeDownloadHistory) GetRefEpisode() (bangumi.Episode, error) {
	panic("unsupported")
}

type MAccount struct {
	Username       string `gorm:"primaryKey"`
	Password       string
	State          string `gorm:"index:state_index"`
	RestrictedTime int64
}
