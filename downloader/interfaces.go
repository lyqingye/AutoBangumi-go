package downloader

import "autobangumi-go/bangumi"

type DownloadHistoryService interface {
	AddDownloadHistory(resource bangumi.Resource) (bangumi.DownLoadHistory, error)
	UpdateDownloadHistory(history bangumi.DownLoadHistory) error
	GetResourceDownloadHistory(resource bangumi.Resource) (bangumi.DownLoadHistory, error)
}
