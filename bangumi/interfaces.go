package bangumi

import "context"

type Storage interface {
	AddBangumi(ctx context.Context, newOrUpdate Bangumi) error
	GetBgmByTitle(ctx context.Context, title string) (Bangumi, error)
	GetBgmByTmDBId(ctx context.Context, tmdbId int64) (Bangumi, error)
	ListBangumis(ctx context.Context, fn func(bgm Bangumi) bool) error
	ListUnDownloadedBangumis(ctx context.Context) ([]Bangumi, error)
	ListDownloadedBangumis(ctx context.Context) ([]Bangumi, error)
	MarkEpisodeDownloaded(ctx context.Context, episode Episode) error
	MarkSeasonDownloaded(ctx context.Context, season Season, download bool) error
	MarkBangumiDownloaded(ctx context.Context, bangumi Bangumi, download bool) error
	GetValidEpisodeResources(ctx context.Context, episode Episode) ([]Resource, error)
	GetResource(ctx context.Context, hash string) (Resource, error)
	UpdateDownloadHistory(ctx context.Context, history EpisodeDownLoadHistory) error
	AddEpisodeDownloadHistory(ctx context.Context, episode Episode, resourcesId string) (EpisodeDownLoadHistory, error)
	MarkResourceIsInvalid(ctx context.Context, resource Resource) error
	GetEpisodeDownloadHistory(ctx context.Context, episode Episode) (EpisodeDownLoadHistory, error)
	RemoveEpisodeDownloadHistory(ctx context.Context, episode Episode) error

	Commit(ctx context.Context) error
	Begin() (context.Context, error)
	Rollback(ctx context.Context) error
}
