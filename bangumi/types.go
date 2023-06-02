package bangumi

import (
	"sort"
	"time"
)

// Resolution 分辨率
type Resolution string

const (
	Resolution2160p   Resolution = "2160p"
	Resolution1080p   Resolution = "1080P"
	Resolution720p    Resolution = "720P"
	ResolutionUnknown Resolution = "unknown"
)

// SubtitleLang 字幕语言
type SubtitleLang string

const (
	SubtitleChs     SubtitleLang = "CHS"
	SubtitleCht     SubtitleLang = "CHT"
	SubtitleUnknown SubtitleLang = "unknown"
)

// SeasonType 动漫季度类型
type SeasonType string

const (
	SeasonTypeSpecial SeasonType = "Special"
	SeasonTypeOVA     SeasonType = "OVA"
	SeasonTypeTV      SeasonType = "TV"
	SeasonTypeNone    SeasonType = "None"
)

// ResourceType 动漫每一集的类型
type ResourceType string

const (
	ResourceTypeSP         ResourceType = "SP"
	ResourceTypeOVA        ResourceType = "OVA"
	ResourceTypeSpecial    ResourceType = "Special"
	ResourceTypeNone       ResourceType = "None"
	ResourceTypeCollection ResourceType = "Collection"
	ResourceTypeUnknown    ResourceType = "unknown"
)

type DownloadState string

const (
	TryDownload DownloadState = "try download"
	Downloading DownloadState = "downloading"
	Downloaded  DownloadState = "downloaded"
	DownloadErr DownloadState = "download err"
)

type Downloader string

const (
	QBDownloader     Downloader = "qb"
	PikpakDownloader Downloader = "pikpak + aria2"
)

var (
	ResolutionPriority = map[Resolution]int{
		Resolution2160p:   4,
		Resolution1080p:   3,
		Resolution720p:    2,
		ResolutionUnknown: 1,
	}

	SubtitlePriority = map[SubtitleLang]int{
		SubtitleChs:     3,
		SubtitleCht:     2,
		SubtitleUnknown: 1,
	}
)

func SelectBestResource(resources []Resource) Resource {
	if len(resources) == 0 {
		return nil
	}
	if len(resources) == 1 {
		return resources[0]
	}

	// priority: resolution | subtitle
	sort.Slice(resources, func(i, j int) bool {
		return ResolutionPriority[resources[i].GetResolution()] > ResolutionPriority[resources[j].GetResolution()] ||
			SubtitlePriority[getResourceBestLang(resources[i])] > SubtitlePriority[getResourceBestLang(resources[j])]
	})

	return resources[0]
}

func getResourceBestLang(resource Resource) SubtitleLang {
	langs := resource.GetSubtitleLang()
	sort.Slice(langs, func(i, j int) bool {
		return SubtitlePriority[langs[i]] > SubtitlePriority[langs[j]]
	})
	return langs[0]
}

type Episode interface {
	GetNumber() uint
	GetResources() ([]Resource, error)
	GetRefSeason() (Season, error)
	IsDownloaded() bool
}

type Resource interface {
	GetRefEpisode() (Episode, error)
	GetTorrent() []byte
	GetTorrentHash() string
	GetSubtitleLang() []SubtitleLang
	GetResolution() Resolution
	GetResourceType() ResourceType
}

type Season interface {
	GetNumber() uint
	GetEpCount() uint
	GetEpisodes() ([]Episode, error)
	GetRefBangumi() (Bangumi, error)
	IsDownloaded() bool
}

type Bangumi interface {
	GetTitle() string
	GetTmDBId() int64
	GetSeasons() ([]Season, error)
	IsDownloaded() bool
}

type DownLoadHistory interface {
	GetState() DownloadState
	GetDownloader() Downloader
	GetDownloaderContext() string
	GetErrMsg() string
	GetTorrent() []byte
	GetTorrentHash() string
	GetRetryCount() int64
	IncRetryCount()
	SetDownloader(downloader Downloader, context string, downloadState DownloadState, error error)
	SetDownloadState(downloadState DownloadState, error error)
	LastUpdatedTime() time.Time
}
