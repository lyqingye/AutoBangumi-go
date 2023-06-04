package db

import (
	"gorm.io/gorm"
	"time"
)

type MTorrent struct {
	Hash        string `gorm:"primaryKey;autoIncrement:false"`
	Bz          []byte
	PubDate     time.Time
	Source      string
	FileSize    uint64
	TorrentType string // bangumi.TorrentTypeCollection | bangumi.TorrentTypeSingleEpisode
}

type MBangumi struct {
	ID         uint   `gorm:"primaryKey"`
	Title      string `gorm:"uniqueIndex"`
	AliasNames string // split by ,
	TMDBId     string

	Seasons []MSeason `gorm:"foreignKey:BangumiId"`
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
	SubjectId  int64
	MikanId    string
	SeasonType string
	State      string // ref SeasonStateComplete | SeasonStateIncomplete

	Episodes []MEpisode `gorm:"foreignKey:SeasonId"`
}

type MEpisodeTorrent struct {
	ID          uint   `gorm:"primaryKey"`
	EpisodeId   uint   `gorm:"uniqueIndex:episode_hash_files_idx"`
	TorrentHash string `gorm:"uniqueIndex:episode_hash_files_idx"`
	FileIndexes string `gorm:"uniqueIndex:episode_hash_files_idx"`

	SubtitleLang string // split by ,
	Resolution   string
}

type MEpisode struct {
	ID       uint              `gorm:"primaryKey"`
	SeasonId uint              `gorm:"uniqueIndex:bangumi_season_idx"`
	Number   uint              `gorm:"uniqueIndex:bangumi_season_idx"`
	Type     string            // ref bangumi.EpisodeTypeSpecial etc...
	Torrents []MEpisodeTorrent `gorm:"foreignKey:EpisodeId"` // a json string of struct [] EpisodeTorrent
}

const (
	DownloadStateFinished    = "finished"
	DownloadStateDownloading = "downloading"
	DownloadStateError       = "error"
)

const (
	ResourceTypeTorrent  = "torrent"
	ResourceTypeHttpLink = "http_link"
	ResourceTypeMagnet   = "magnet"
)

type DownloadHistory struct {
	gorm.Model

	BangumiId uint `gorm:"uniqueIndex:bangumi_season_episode_idx"`
	SeasonId  uint `gorm:"uniqueIndex:bangumi_season_episode_idx"`
	EpisodeId uint `gorm:"uniqueIndex:bangumi_season_episode_idx"`

	Downloader   string
	ResourceType string // ref ResourceTypeTorrent | ResourceTypeHttpLink | ResourceTypeMagnet
	ResourceId   string // ref MTorrent Table
	Context      string // json value use by downloader
	State        string `gorm:"index:episode_index"` // ref DownloadStateFinished | DownloadStateDownloading | DownloadStateError
	ErrorMsg     string
	RetryCount   int64
}
