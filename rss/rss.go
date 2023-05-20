package rss

import (
	"encoding/binary"
	"errors"
	"fmt"
	bangumitypes "pikpak-bot/bangumi"
	"pikpak-bot/bus"
	"pikpak-bot/db"
	"pikpak-bot/mdb"
	"pikpak-bot/utils"
	"sync"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"

	"github.com/rs/zerolog"
)

var (
	KeyRSSLink          = []byte("rss-links")
	KeyRSSSubscribeInfo = []byte("rss-subscribe")
)

var (
	ResolutionPriority = map[string]int{
		bangumitypes.Resolution1080p:   3,
		bangumitypes.Resolution720p:    2,
		bangumitypes.ResolutionUnknown: 1,
	}

	SubtitlePriority = map[string]int{
		bangumitypes.SubtitleChs:     3,
		bangumitypes.SubtitleCht:     2,
		bangumitypes.SubtitleUnknown: 1,
	}
)

type SubscribeInfo struct {
	RSSLink  string
	Bangumis []bangumitypes.Bangumi
}

type RSSInfo struct {
	Bangumis []bangumitypes.Bangumi
}

// RSSManager
// Bangumi Storage Structure
// Indexes:
// - RssLink -> BangumiInfo
// - SubjectId -> BangumiInfo
// - (SubjectId,Season,Episode) -> Episode

type RSSManager struct {
	eb              *bus.EventBus
	db              *db.DB
	ticker          *time.Ticker
	parsers         []*MikanRSSParser
	logger          zerolog.Logger
	lock            sync.Mutex
	tmdb            *tmdb.Client
	bangumiTvClient *mdb.BangumiTVClient
}

func NewRSSManager(eb *bus.EventBus, db *db.DB, period time.Duration, tmdbClient *tmdb.Client, bangumiTVClient *mdb.BangumiTVClient) (*RSSManager, error) {
	man := RSSManager{
		eb:              eb,
		db:              db,
		logger:          utils.GetLogger("RSSManager"),
		lock:            sync.Mutex{},
		tmdb:            tmdbClient,
		bangumiTvClient: bangumiTVClient,
	}
	man.ticker = time.NewTicker(period)
	err := man.initRSSLinkFromDB()
	if err != nil {
		return nil, err
	}
	return &man, nil
}

func (man *RSSManager) initRSSLinkFromDB() error {
	rssLinks, err := man.ListRSSLink()
	if err != nil {
		return err
	}
	for _, rss := range rssLinks {
		err := man.AddMikanRss(rss)
		if err != nil {
			return err
		}
		man.logger.Debug().Str("link", rss).Msg("load rss link")
	}
	return nil
}

func (man *RSSManager) saveRSSLinkToDB(newRssLink string) error {
	rssLinks, err := man.ListRSSLink()
	if err != nil {
		return err
	}
	found := false
	for _, rss := range rssLinks {
		if rss == newRssLink {
			found = true
			break
		}
	}
	if !found {
		rssLinks = append(rssLinks, newRssLink)
	}
	return man.db.Set(KeyRSSLink, &rssLinks)
}

func (man *RSSManager) removeRSSLinkFromDB(rssLink string) error {
	rssLinks, err := man.ListRSSLink()
	if err != nil {
		return err
	}
	var finalRssLinks []string
	found := false
	for _, rss := range rssLinks {
		if rss == rssLink {
			found = true
		} else {
			finalRssLinks = append(finalRssLinks, rss)
		}
	}
	if found {
		return man.db.Set(KeyRSSLink, &finalRssLinks)
	}
	return nil
}

func (man *RSSManager) ListRSSLink() ([]string, error) {
	var rssLinks []string
	found, err := man.db.Get(KeyRSSLink, &rssLinks)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return rssLinks, nil
}

func (man *RSSManager) Start() {
	man.logger.Info().Msg("start rss manager")
	man.Refresh()
	for range man.ticker.C {
		man.Refresh()
	}
}

func (man *RSSManager) AddMikanRss(mikanRss string) error {
	for _, parser := range man.parsers {
		if mikanRss == parser.rssLink {
			return nil
		}
	}
	parser, err := NewMikanRSSParser(mikanRss, man.eb, man.db, man.tmdb, man.bangumiTvClient)
	if err != nil {
		return err
	}
	man.parsers = append(man.parsers, parser)
	err = man.saveRSSLinkToDB(mikanRss)

	go man.refreshParser(parser)
	return err
}

func (man *RSSManager) RemoveMikanRss(mikanRss string) error {
	return man.removeRSSLinkFromDB(mikanRss)
}

func (man *RSSManager) Refresh() {
	man.logger.Debug().Msg("try refresh all rss")
	man.lock.Lock()
	man.logger.Debug().Msg("start refresh all rss")
	defer man.lock.Unlock()
	for _, parser := range man.parsers {
		man.refreshParser(parser)
	}
}

func (man *RSSManager) refreshParser(parser *MikanRSSParser) {
	man.logger.Info().Str("rssLink", parser.rssLink).Msg("refresh RSS")
	rssInfo, err := parser.Parse()
	if err != nil {
		man.logger.Error().Str("rssLink", parser.rssLink).Err(err).Msg("refresh RSS Failed")
		return
	}
	err = man.updateIndex(parser.rssLink, rssInfo)
	if err != nil {
		man.logger.Error().Str("rssLink", parser.rssLink).Err(err).Msg("Update index Failed")
		return
	}
	for _, bangumi := range rssInfo.Bangumis {
		for _, ep := range bangumi.Episodes {
			if !man.alreadyRead(&ep) {
				man.eb.Publish(bus.RSSTopic, bus.Event{
					EventType: bus.RSSUpdateEventType,
					Inner:     ep,
				})
				man.logger.Info().Str("bangumi", ep.BangumiTitle).Uint("season", ep.Season).Uint("ep", ep.EPNumber).Msg("rss update")
			}
		}
	}
}

func (man *RSSManager) MarkEpisodeAsRead(latestEp *bangumitypes.Episode) error {
	ep, err := man.getEpisode(latestEp.SubjectId, latestEp.Season, latestEp.EPNumber)
	if err != nil {
		return err
	}
	ep.Read = true
	return man.setEpisode(ep)
}

func (man *RSSManager) MarkEpisodeUnRead(latestEp *bangumitypes.Episode) error {
	ep, err := man.getEpisode(latestEp.SubjectId, latestEp.Season, latestEp.EPNumber)
	if err != nil {
		return err
	}
	ep.Read = false
	return man.setEpisode(ep)
}

func (man *RSSManager) GetBangumiByEpisode(ep *bangumitypes.Episode) (*bangumitypes.Bangumi, error) {
	bangumi := bangumitypes.Bangumi{}
	found, err := man.db.Get(getSubKeyBySubjectId(ep.SubjectId), &bangumi)
	if err != nil {
		return nil, err
	}
	if found {
		return &bangumi, nil
	}
	return nil, errors.New("bangumi not found")
}

func (man *RSSManager) alreadyRead(latestEp *bangumitypes.Episode) bool {
	episode, err := man.getEpisode(latestEp.SubjectId, latestEp.Season, latestEp.EPNumber)
	if err != nil {
		return false
	}
	return episode.Read
}

func (man *RSSManager) getEpisode(subjectId int64, season uint, ep uint) (*bangumitypes.Episode, error) {
	episode := bangumitypes.Episode{}
	found, err := man.db.Get(getEpisodeKey(subjectId, season, ep), &episode)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("episode: %d %d %d not found", subjectId, season, ep)
	}
	return &episode, nil
}

func (man *RSSManager) setEpisode(ep *bangumitypes.Episode) error {
	return man.db.Set(getEpisodeKey2(ep), ep)
}

func (man *RSSManager) updateIndex(rssLink string, rssInfo *RSSInfo) error {
	for _, bangumi := range rssInfo.Bangumis {
		if found, err := man.db.Has(getSubKeyByRSSLink(rssLink)); err != nil {
			return err
		} else {
			bangumiCopy := bangumi
			bangumiCopy.Episodes = nil
			err = man.db.Set(getSubKeyByRSSLink(rssLink), &bangumiCopy)
			if err != nil {
				return err
			}
			err = man.db.Set(getSubKeyBySubjectId(bangumiCopy.SubjectId), &bangumiCopy)
			if err != nil {
				return err
			}
			if !found {
				for _, ep := range bangumi.Episodes {
					err = man.db.Set(getEpisodeKey2(&ep), &ep)
					if err != nil {
						return err
					}
				}
			} else {
				for _, ep := range bangumi.Episodes {
					if existEpisode, err := man.db.Has(getEpisodeKey2(&ep)); err != nil || existEpisode {
						continue
					}
					err = man.db.Set(getEpisodeKey2(&ep), &ep)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func getEpisodeKey(subjectId int64, season uint, ep uint) []byte {
	return append(KeyRSSSubscribeInfo, []byte(fmt.Sprintf("%d-%d-%d", subjectId, season, ep))...)
}
func getEpisodeKey2(ep *bangumitypes.Episode) []byte {
	return getEpisodeKey(ep.SubjectId, ep.Season, ep.EPNumber)
}

func getSubKeyByRSSLink(rssLink string) []byte {
	return append(KeyRSSSubscribeInfo, []byte(rssLink)...)
}

func getSubKeyBySubjectId(subjectId int64) []byte {
	byteSlice := make([]byte, 8)
	binary.BigEndian.PutUint64(byteSlice, uint64(subjectId))
	return append(KeyRSSSubscribeInfo, byteSlice...)
}
