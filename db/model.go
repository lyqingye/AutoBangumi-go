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
	ID         uint `gorm:"primaryKey"`
	Title      string
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
	BangumiId  uint `gorm:"primaryKey;autoIncrement:false"`
	Number     uint `gorm:"primaryKey;autoIncrement:false"`
	EpCount    uint
	SubjectId  int64
	MikanId    string
	SeasonType string
	State      string // ref SeasonStateComplete | SeasonStateIncomplete

	Episodes []MEpisode `gorm:"foreignKey:SeasonId"`
}

type EpisodeTorrent struct {
	TorrentId    uint   // MTorrent Id
	SubtitleLang string // split by ,
	Resolution   string
	FileIndexes  []int64 // file indexes of torrent files
}

type MEpisode struct {
	ID        uint `gorm:"primaryKey"`
	BangumiId uint `gorm:"primaryKey;autoIncrement:false"`
	SeasonId  uint `gorm:"primaryKey;autoIncrement:false"`
	Number    uint
	Type      string // ref bangumi.EpisodeTypeSpecial etc...
	Torrents  string // a json string of struct [] EpisodeTorrent
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

	BangumiId uint `gorm:"index:episode_index"`
	SeasonId  uint `gorm:"index:episode_index"`
	EpisodeId uint `gorm:"index:episode_index"`

	Downloader   string
	ResourceType string // ref ResourceTypeTorrent | ResourceTypeHttpLink | ResourceTypeMagnet
	ResourceId   string // ref MTorrent Table
	Context      string // json value use by downloader
	State        string `gorm:"index:episode_index"` // ref DownloadStateFinished | DownloadStateDownloading | DownloadStateError
	ErrorMsg     string
	RetryCount   int64
}
