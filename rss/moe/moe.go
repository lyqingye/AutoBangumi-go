package moe

import (
	"autobangumi-go/bus"
	"autobangumi-go/utils"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"net/url"
)

type BangumiMoe struct {
	endpoint *url.URL
	http     *resty.Client
	eb       *bus.EventBus
	logger   zerolog.Logger
}

func NewBangumiMoe() (*BangumiMoe, error) {
	endpoint, err := url.Parse("https://bangumi.moe")
	if err != nil {
		return nil, err
	}
	moe := BangumiMoe{
		endpoint: endpoint,
		http:     resty.New(),
		logger:   utils.GetLogger("bangumi-moe"),
	}
	return &moe, nil
}
