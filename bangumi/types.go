package bangumi

import "errors"

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
	Title     string
	EPCount   uint
	Season    uint
	TmDBId    int64
	SubjectId int64
	Episodes  []Episode
}

type Episode struct {
	BangumiTitle   string
	SubjectId      int64 // BangumiTV SubjectId
	MikanBangumiId string
	EpisodeTitle   string
	Subgroup       string
	Season         uint
	EPNumber       uint
	Magnet         string
	TorrentHash    string
	Torrent        []byte
	Date           string
	FileSize       uint64
	Lang           []string
	Resolution     string
	Read           bool // mark as read
	EpisodeType    string
}

func (e *Episode) Validate() error {
	if e.EPNumber <= 0 {
		return errors.New("invalid ep number")
	}
	if e.Magnet == "" && len(e.Torrent) == 0 {
		return errors.New("empty download resource")
	}
	if e.FileSize == 0 {
		return errors.New("invalid filesize")
	}
	if e.BangumiTitle == "" {
		return errors.New("empty bangumi title")
	}
	return nil
}
