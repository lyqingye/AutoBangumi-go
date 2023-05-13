package downloader

const (
	StatusFinished    = "finished"
	StatusDownloading = "downloading"
	StatusWaiting     = "wanting"
	StatusPause       = "pause"
	StatusError       = "error"
	StatusRemoved     = "removed"
)

type DownloadStatus struct {
	Filename   string
	Speed      string
	Progress   string
	Status     string
	ErrMessage string
}

type OnlineDownloader interface {
	AddTask(url string, dir string, filename string) (string, error)
	QueryStatus(tid string) (DownloadStatus, error)
	RemoveTask(url string) error
	PauseTask(tid string) error
	ResumeTask(tid string) error
	ListTasks() ([]DownloadStatus, error)
}
