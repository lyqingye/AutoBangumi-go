package mikan

import (
	"encoding/xml"
	"net/url"

	"autobangumi-go/bus"
	"autobangumi-go/utils"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

type MikanRSSParser struct {
	logger        zerolog.Logger
	mikanEndpoint *url.URL
	rssLink       string
	http          *resty.Client
	eb            *bus.EventBus
	tmdb          TMDB
	bangumiTV     BangumiTV
	cm            CacheManager
}

func NewMikanRSSParser(rss string, tmDB TMDB, bangumiTV BangumiTV, cm CacheManager) (*MikanRSSParser, error) {
	uri, err := url.Parse(rss)
	if err != nil {
		return nil, err
	}
	endpoint, err := url.Parse(uri.Hostname())
	if err != nil {
		return nil, err
	}
	endpoint.Scheme = uri.Scheme
	parser := MikanRSSParser{
		logger:        utils.GetLogger("mikanRSS"),
		mikanEndpoint: endpoint,
		rssLink:       rss,
		http:          resty.New(),
		tmdb:          tmDB,
		bangumiTV:     bangumiTV,
		cm:            cm,
	}
	return &parser, nil
}

func (parser *MikanRSSParser) RssLink() string {
	return parser.rssLink
}

func (parser *MikanRSSParser) Parse() ([]*Bangumi, error) {
	var err error
	mikan, err := parser.getRss(parser.rssLink)
	if err != nil {
		return nil, err
	}
	return parser.parseMikanRSS(mikan)
}

func (parser *MikanRSSParser) getRss(link string) (*MikanRss, error) {
	resp, err := parser.http.R().EnableTrace().Get(link)
	if err != nil {
		return nil, err
	}
	rssContent := MikanRss{}
	err = xml.Unmarshal(resp.Body(), &rssContent)
	if err != nil {
		return nil, err
	}
	return &rssContent, nil
}

func (parser *MikanRSSParser) Close() error {
	if err := parser.cm.Close(); err != nil {
		return err
	}
	return nil
}
