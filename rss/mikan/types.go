package mikan

import (
	"encoding/xml"
	"time"

	bangumitypes "autobangumi-go/bangumi"
	"github.com/pkg/errors"
)

type MikanRssItem struct {
	Guid struct {
		IsPermaLink string `xml:"isPermaLink,attr"`
	} `xml:"guid"`
	Link        string `xml:"link"`
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Torrent     struct {
		Xmlns         string `xml:"xmlns,attr"`
		Link          string `xml:"link"`
		ContentLength string `xml:"contentLength"`
		PubDate       string `xml:"pubDate"`
	} `xml:"torrent"`
	Enclosure struct {
		Type   string `xml:"type,attr"`
		Length string `xml:"length,attr"`
		URL    string `xml:"url,attr"`
	} `xml:"enclosure"`
}

type MikanRss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Title       string         `xml:"title"`
		Link        string         `xml:"link"`
		Description string         `xml:"description"`
		Item        []MikanRssItem `xml:"item"`
	} `xml:"channel"`
}

type Bangumi struct {
	Info    BangumiInfo     `json:"info"`
	Seasons map[uint]Season `json:"seasons"`
}

func (b *Bangumi) IsCollected() bool {
	if len(b.Seasons) == 0 {
		return false
	}

	for _, season := range b.Seasons {
		if !season.IsCollected() {
			return false
		}
	}
	return true
}

func (b *Bangumi) GetTitle() string {
	return b.Info.Title
}

func (b *Bangumi) GetSeasons() ([]bangumitypes.Season, error) {
	var ret []bangumitypes.Season
	for _, season := range b.Seasons {
		copyValue := season
		ret = append(ret, &copyValue)
	}
	return ret, nil
}

func (b *Bangumi) GetTmDBId() int64 {
	return b.Info.TmDBId
}

func (b *Bangumi) GetMikanID() string {
	return b.Info.MikanID
}

func (s *Bangumi) IsDownloaded() bool {
	panic("unsupported")
}

type Season struct {
	SubjectId      int64
	MikanBangumiId string
	Number         uint
	EpCount        uint
	Episodes       map[uint]Episode
}

func (s *Season) IsCollected() bool {
	return int(s.EpCount) == len(s.Episodes)
}

func (s *Season) GetNumber() uint {
	return s.Number
}

func (s *Season) GetEpCount() uint {
	return s.EpCount
}

func (s *Season) GetEpisodes() ([]bangumitypes.Episode, error) {
	var ret []bangumitypes.Episode
	for _, ep := range s.Episodes {
		copyValue := ep
		ret = append(ret, &copyValue)
	}
	return ret, nil
}

func (e *Season) GetRefBangumi() (bangumitypes.Bangumi, error) {
	panic("unsupported")
}

func (s *Season) IsDownloaded() bool {
	panic("unsupported")
}

func (s *Season) RemoveInvalidEpisode() {
	for k, ep := range s.Episodes {
		if ep.GetNumber() > s.GetEpCount() {
			delete(s.Episodes, k)
		}
	}
}

type BangumiInfo struct {
	Title   string
	TmDBId  int64
	MikanID string
}

type Episode struct {
	Number    uint
	Resources []TorrentResource
}

func (s *Episode) IsDownloaded() bool {
	panic("unsupported")
}

type TorrentResource struct {
	RawFilename    string
	Subgroup       string
	Magnet         string
	TorrentHash    string
	Torrent        []byte
	TorrentPubDate time.Time
	FileSize       uint64
	SubtitleLang   []bangumitypes.SubtitleLang
	Resolution     bangumitypes.Resolution
	Type           bangumitypes.ResourceType
}

func (torrent *TorrentResource) GetTorrent() []byte {
	return torrent.Torrent
}

func (torrent *TorrentResource) GetTorrentHash() string {
	return torrent.TorrentHash
}

func (torrent *TorrentResource) GetSubtitleLang() []bangumitypes.SubtitleLang {
	return torrent.SubtitleLang
}

func (torrent *TorrentResource) GetResolution() bangumitypes.Resolution {
	return torrent.Resolution
}

func (torrent *TorrentResource) GetResourceType() bangumitypes.ResourceType {
	return torrent.Type
}

func (torrent *TorrentResource) Validate() error {
	if torrent.Magnet == "" && len(torrent.Torrent) == 0 {
		return errors.New("empty download resource")
	}
	if torrent.FileSize == 0 {
		return errors.New("invalid filesize")
	}
	return nil
}

func (torrent *TorrentResource) GetRefEpisode() (bangumitypes.Episode, error) {
	panic("unsupported")
}

func (e *Episode) GetNumber() uint {
	return e.Number
}

func (e *Episode) GetResources() ([]bangumitypes.Resource, error) {
	var newResource []bangumitypes.Resource
	for _, rs := range e.Resources {
		copyValue := rs
		newResource = append(newResource, &copyValue)
	}
	return newResource, nil
}

func (e *Episode) GetRefSeason() (bangumitypes.Season, error) {
	panic("unsupported")
}

func (e *Episode) Validate() error {
	if e.Number <= 0 {
		return errors.New("invalid ep number")
	}

	return nil
}
