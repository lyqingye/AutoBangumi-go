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

	EpisodeStateParsed   = "parsed"
	EpisodeStateParseErr = "error"
	EpisodeStateDownload = "download"
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
	State          string    `json:"state"`
}

func (e *Episode) Compare(o *Episode) bool {
	langI := e.SubtitleLang[0]
	langJ := o.SubtitleLang[0]
	return ResolutionPriority[e.Resolution] > ResolutionPriority[o.Resolution] ||
		SubtitlePriority[langI] > SubtitlePriority[langJ] ||
		e.TorrentPubDate.After(o.TorrentPubDate)
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
