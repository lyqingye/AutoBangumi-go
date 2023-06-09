package bangumi

import (
	"errors"
	"time"
)

const (
	Resolution2160p   = "2160p"
	Resolution1080p   = "1080P"
	Resolution720p    = "720P"
	ResolutionUnknown = "unknown"

	SubtitleChs     = "CHS"
	SubtitleCht     = "CHT"
	SubtitleUnknown = "unknown"

	EpisodeTypeSP         = "SP"
	EpisodeTypeOVA        = "OVA"
	EpisodeTypeSpecial    = "Special"
	EpisodeTypeNone       = "None"
	EpisodeTypeCollection = "Collection"
	EpisodeTypeUnknown    = "unknown"

	NoDownloader = "None"
)

var (
	ResolutionPriority = map[string]int{
		Resolution2160p:   4,
		Resolution1080p:   3,
		Resolution720p:    2,
		ResolutionUnknown: 1,
	}

	SubtitlePriority = map[string]int{
		SubtitleChs:     3,
		SubtitleCht:     2,
		SubtitleUnknown: 1,
	}
)

type Bangumi struct {
	Info    BangumiInfo     `json:"info"`
	Seasons map[uint]Season `json:"seasons"`
}

func (bangumi *Bangumi) IsComplete() bool {
	if len(bangumi.Seasons) == 0 {
		return false
	}

	result := true
	for _, season := range bangumi.Seasons {
		if len(season.Complete) != int(season.EpCount) {
			result = false
			break
		}
	}
	return result
}

type Season struct {
	SubjectId      int64     `json:"subjectId"`
	MikanBangumiId string    `json:"mikanBangumiId"`
	Number         uint      `json:"number"`
	EpCount        uint      `json:"epcount"`
	Episodes       []Episode `json:"episodes"`
	Complete       []uint    `json:"complete"`
}

func (season *Season) ListIncompleteEpisodes() []Episode {
	var result []Episode
	for _, ep := range season.Episodes {
		if !season.IsComplete(ep.Number) {
			result = append(result, ep)
		}
	}
	return result
}

func (season *Season) IsComplete(epNum uint) bool {
	for _, number := range season.Complete {
		if epNum == number {
			return true
		}
	}
	return false
}

func (season *Season) RemoveComplete(epNum uint) {
	var newArray []uint
	for _, number := range season.Complete {
		if epNum != number {
			newArray = append(newArray, number)
		}
	}
	season.Complete = newArray
}

func (season *Season) IsEpisodesCollected() bool {
	return int(season.EpCount) == len(season.Episodes)
}

type BangumiInfo struct {
	Title  string `json:"title"`
	TmDBId int64  `json:"tmdbId"`
}

type Episode struct {
	Number         uint      `json:"number"`
	RawFilename    string    `json:"rawFilename"`
	Subgroup       string    `json:"subgroup"`
	Magnet         string    `json:"magnet"`
	TorrentHash    string    `json:"torrentHash"`
	Torrent        []byte    `json:"torrent"`
	TorrentPubDate time.Time `json:"torrentPubDate"`
	FileSize       uint64    `json:"fileSize"`
	SubtitleLang   []string  `json:"subtitleLang"`
	Resolution     string    `json:"resolution"`
	Type           string    `json:"episodeType"`

	// modify my downloader
	DownloadState DownloadState `json:"downloadState"`
}

func (e *Episode) IsNeedToDownload() bool {
	return e.DownloadState.Downloader == ""
}

func (e *Episode) IsCHSubtitle() bool {
	for _, sub := range e.SubtitleLang {
		switch sub {
		case SubtitleCht, SubtitleChs:
			return true
		}
	}
	return false
}

type DownloadState struct {
	Downloader string `json:"downloader"`
	TaskId     string `json:"taskId"`
}

func (e *Episode) Compare(o *Episode) bool {
	langI := e.SubtitleLang[0]
	langJ := o.SubtitleLang[0]
	return ResolutionPriority[e.Resolution] > ResolutionPriority[o.Resolution] ||
		SubtitlePriority[langI] > SubtitlePriority[langJ] ||
		e.TorrentPubDate.After(o.TorrentPubDate)
}

func (e *Episode) CanReplace(replacement *Episode) bool {
	if !e.IsCHSubtitle() && replacement.IsCHSubtitle() {
		return true
	}
	if ResolutionPriority[replacement.Resolution] > ResolutionPriority[e.Resolution] {
		return true
	}
	return false
}

func (e *Episode) Validate() error {
	if e.Number <= 0 {
		return errors.New("invalid ep number")
	}
	if e.Magnet == "" && len(e.Torrent) == 0 {
		return errors.New("empty download resource")
	}
	if e.FileSize == 0 {
		return errors.New("invalid filesize")
	}
	return nil
}
