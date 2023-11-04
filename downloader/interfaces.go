package downloader

import "autobangumi-go/bangumi"

type DownloadService interface {
	AddEpisodeDownloadHistory(episode bangumi.Episode, resourcesId string) (bangumi.EpisodeDownLoadHistory, error)
	UpdateDownloadHistory(history bangumi.EpisodeDownLoadHistory) error
	GetEpisodeDownloadHistory(episode bangumi.Episode) (bangumi.EpisodeDownLoadHistory, error)
	RemoveEpisodeDownloadHistory(episode bangumi.Episode) error
	MarkResourceIsInvalid(resource bangumi.Resource) error
}
