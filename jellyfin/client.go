package jellyfin

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

var (
	ErrUnAuthorized            = errors.New("UnAuthorized")
	ErrForbidden               = errors.New("Forbidden")
	BaseAuthorizationHeaderKey = "X-Emby-Authorization"
	BaseAuthorizationHeader    = "MediaBrowser Client=\"Jellyfin Web\", Device=\"Chrome\", DeviceId=\"TW96aWxsYS81LjAgKFgxMTsgTGludXggeDg2XzY0KSBBcHBsZVdlYktpdC81MzcuMzYgKEtIVE1MLCBsaWtlIEdlY2tvKSBDaHJvbWUvMTEyLjAuMC4wIFNhZmFyaS81MzcuMzZ8MTY4NTQ1NzY3NTU1MQ11\", Version=\"10.8.11\""
)

type Client struct {
	http *resty.Client
}

func NewClient(endpoint string, username string, password string) (*Client, error) {
	ret := Client{}
	inner := resty.New()
	ret.http = inner
	inner.SetDebug(true)
	inner.SetRetryCount(1)
	inner.SetRetryWaitTime(time.Second * 1)
	inner.SetBaseURL(endpoint)
	inner.SetHeader(BaseAuthorizationHeaderKey, BaseAuthorizationHeader)

	inner.AddRetryCondition(func(response *resty.Response, err error) bool {
		if err != nil {
			return false
		}
		if strings.Contains(response.Request.URL, "/Users/AuthenticateByName") {
			return false
		}
		switch response.StatusCode() {
		case http.StatusUnauthorized:
			resp, err := ret.Login(username, password)
			if err != nil {
				return false
			}
			ret.http.SetHeader(BaseAuthorizationHeaderKey, fmt.Sprintf("%s, Token=\"%s\"", BaseAuthorizationHeader, resp.AccessToken))
			response.Request.SetHeader(BaseAuthorizationHeaderKey, fmt.Sprintf("%s, Token=\"%s\"", BaseAuthorizationHeader, resp.AccessToken))
			return true
		}
		return false
	})
	return &Client{http: inner}, nil
}

func (c *Client) Login(username string, password string) (*RespAuthenticateUserByName, error) {
	result := RespAuthenticateUserByName{}
	rawResp, err := c.http.R().EnableTrace().SetBody(ReqAuthenticateUserByName{
		Username: username,
		Pw:       password,
	}).SetResult(&result).Post("/Users/AuthenticateByName")
	if err != nil {
		return nil, err
	}

	switch rawResp.StatusCode() {
	case http.StatusOK:
		return &result, nil
	case http.StatusUnauthorized:
		return nil, ErrUnAuthorized
	case http.StatusForbidden:
		return nil, ErrForbidden
	}
	return nil, errors.Errorf("unknown response code: %d", rawResp.StatusCode())
}

func (c *Client) StartLibraryScan() error {
	rawResp, err := c.http.R().Post("/Library/Refresh")
	if err != nil {
		return err
	}
	switch rawResp.StatusCode() {
	case http.StatusNoContent:
		return nil
	case http.StatusUnauthorized:
		return ErrUnAuthorized
	case http.StatusForbidden:
		return ErrForbidden
	}
	return errors.Errorf("unknown response code: %d", rawResp.StatusCode())
}
