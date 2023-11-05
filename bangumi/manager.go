package bangumi

import (
	"context"

	"github.com/pkg/errors"
)

type Manager struct {
	store Storage
}

func (m *Manager) AddEpisodeDownloadHistory(episode Episode, resourcesId string) (EpisodeDownLoadHistory, error) {
	return m.store.AddEpisodeDownloadHistory(nil, episode, resourcesId)
}

func (m *Manager) MarkResourceIsInvalid(resource Resource) error {
	return m.store.MarkResourceIsInvalid(nil, resource)
}

func (m *Manager) GetEpisodeDownloadHistory(episode Episode) (EpisodeDownLoadHistory, error) {
	return m.store.GetEpisodeDownloadHistory(nil, episode)
}

func (m *Manager) RemoveEpisodeDownloadHistory(episode Episode) error {
	return m.store.RemoveEpisodeDownloadHistory(nil, episode)
}

func NewManager(store Storage) *Manager {
	return &Manager{
		store: store,
	}
}

func (m *Manager) UpdateDownloadHistory(history EpisodeDownLoadHistory) error {
	ctx, err := m.store.Begin()
	if err != nil {
		return err
	}
	err = m.updateDownloadHistoryInternal(ctx, history)
	if err != nil {
		_ = m.store.Rollback(ctx)
		return err
	}
	return m.store.Commit(ctx)
}

func (m *Manager) AddBangumi(newOrUpdate Bangumi) error {
	return m.store.AddBangumi(nil, newOrUpdate)
}

func (m *Manager) GetValidEpisodeResources(episode Episode) ([]Resource, error) {
	return m.store.GetValidEpisodeResources(nil, episode)
}

func (m *Manager) ListUnDownloadedBangumis() ([]Bangumi, error) {
	return m.store.ListUnDownloadedBangumis(nil)
}

func (m *Manager) ListDownloadedBangumis(ctx context.Context) ([]Bangumi, error) {
	return m.store.ListDownloadedBangumis(ctx)
}

func (m *Manager) GetBgmByTitle(title string) (Bangumi, error) {
	return m.store.GetBgmByTitle(nil, title)
}

func (m *Manager) GetBgmByTmDBId(tmdbId int64) (Bangumi, error) {
	return m.store.GetBgmByTmDBId(nil, tmdbId)
}

func (m *Manager) updateDownloadHistoryInternal(ctx context.Context, history EpisodeDownLoadHistory) error {
	if history.GetState() == Downloaded {
		history.SetDownloadState(history.GetState(), nil)
		episode, err := history.GetRefEpisode()
		if err != nil {
			return err
		}
		if episode == nil {
			return errors.Errorf("episode not found for resource %s", history.GetResourcesIds())
		}
		if err := m.store.MarkEpisodeDownloaded(ctx, episode); err != nil {
			return err
		}

		season, err := episode.GetRefSeason()
		if err != nil {
			return err
		}

		episodes, err := season.GetEpisodes()
		if err != nil {
			return err
		}

		allEpisodeDownloaded := true
		for _, ep := range episodes {
			if !ep.IsDownloaded() {
				allEpisodeDownloaded = false
			}
		}
		allEpisodeDownloaded = uint(len(episodes)) == season.GetEpCount() && allEpisodeDownloaded
		if err := m.store.MarkSeasonDownloaded(ctx, season, allEpisodeDownloaded); err != nil {
			return err
		}

		bgm, err := season.GetRefBangumi()
		if err != nil {
			return err
		}
		seasons, err := bgm.GetSeasons()
		allSeasonDownloaded := true
		if err != nil {
			return err
		}
		for _, s := range seasons {
			if !s.IsDownloaded() {
				allSeasonDownloaded = false
			}
		}

		if err := m.store.MarkBangumiDownloaded(ctx, bgm, allSeasonDownloaded); err != nil {
			return err
		}
	}
	return m.store.UpdateDownloadHistory(ctx, history)
}
