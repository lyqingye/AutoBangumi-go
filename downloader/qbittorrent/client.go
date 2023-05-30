package qbittorrent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	torrent "github.com/anacrolix/torrent/metainfo"
	"github.com/go-resty/resty/v2"
)

type QbittorrentClient struct {
	endpoint *url.URL
	username string
	password string
	client   *resty.Client
	dir      string
	cookies  []*http.Cookie
}

func NewQbittorrentClient(apiEndpoint, username, password string, dir string) (*QbittorrentClient, error) {
	endpoint, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, err
	}
	qb := QbittorrentClient{
		endpoint: endpoint,
		username: username,
		password: password,
		dir:      dir,
		client:   resty.New().EnableTrace(),
	}
	return &qb, nil
}

func (qb *QbittorrentClient) Login() error {
	params := map[string]string{
		"username": qb.username,
		"password": qb.password,
	}
	resp, err := qb.client.R().
		SetFormData(params).
		Post(qb.endpoint.JoinPath("/api/v2/auth/login").String())
	if err != nil {
		return err
	}
	qb.cookies = resp.Cookies()
	return respToErr(resp.Body())
}

func (qb *QbittorrentClient) Logout() error {
	resp, err := qb.client.R().
		SetCookies(qb.cookies).
		Post(qb.endpoint.JoinPath("/api/v2/auth/logout").String())
	if err != nil {
		return err
	}
	if len(resp.Body()) == 0 {
		return nil
	}
	return respToErr(resp.Body())
}

func (qb *QbittorrentClient) AddTorrent(torrentName string, bz []byte, dir string) (string, error) {
	hash := ""
	metaInfo, err := torrent.Load(bytes.NewBuffer(bz))
	if err != nil {
		return hash, err
	}
	hash = metaInfo.HashInfoBytes().HexString()
	req := qb.client.R()
	options := AddTorrentOptions{
		Paused:   true,
		SavePath: filepath.Join(qb.dir, dir),
		Rename:   torrentName,
	}
	req = req.SetFormData(options.toMap())
	req = req.SetMultipartField("", fmt.Sprintf("%s.torrent", hash), "application/x-bittorrent", bytes.NewBuffer(bz))
	req = req.SetCookies(qb.cookies)
	resp, err := req.Post(qb.endpoint.JoinPath("/api/v2/torrents/add").String())
	if err != nil {
		return hash, nil
	}
	return hash, respToErr(resp.Body())
}

func (qb *QbittorrentClient) AddTorrentEx(options *AddTorrentOptions, bz []byte, dir string) (string, error) {
	hash := ""
	metaInfo, err := torrent.Load(bytes.NewBuffer(bz))
	if err != nil {
		return hash, err
	}
	hash = metaInfo.HashInfoBytes().HexString()

	req := qb.client.R()
	options.SavePath = filepath.Join(qb.dir, dir)
	req = req.SetFormData(options.toMap())
	req = req.SetMultipartField("", fmt.Sprintf("%s.torrent", hash), "application/x-bittorrent", bytes.NewBuffer(bz))
	req = req.SetCookies(qb.cookies)
	resp, err := req.Post(qb.endpoint.JoinPath("/api/v2/torrents/add").String())
	if err != nil {
		return hash, nil
	}
	return hash, respToErr(resp.Body())
}

func (qb *QbittorrentClient) ListAllTorrent(filter string) ([]Torrent, error) {
	limit := 100
	offset := 0
	var torrentList []Torrent
	for {
		list, err := qb.ListTorrent(&RequestTorrentList{
			Filter: filter,
			Offset: offset,
			Limit:  limit,
		})
		if err != nil {
			return nil, err
		}
		torrentList = append(torrentList, list...)
		if len(list) < limit {
			break
		}
		offset = offset + len(list)
	}
	return torrentList, nil
}

func (qb *QbittorrentClient) ListTorrent(req *RequestTorrentList) ([]Torrent, error) {
	resp, err := qb.client.R().SetCookies(qb.cookies).
		SetQueryParams(req.toMap()).
		Get(qb.endpoint.JoinPath("/api/v2/torrents/info").String())
	if err != nil {
		return nil, err
	}
	var torrentList []Torrent
	err = json.Unmarshal(resp.Body(), &torrentList)
	return torrentList, err
}

func (qb *QbittorrentClient) GetTorrent(hash string) (*Torrent, error) {
	req := RequestTorrentList{
		Offset: 0,
		Limit:  1,
		Hashes: []string{hash},
	}

	torrents, err := qb.ListTorrent(&req)
	if err != nil {
		return nil, err
	}
	if len(torrents) == 0 {
		return nil, ErrTorrentNotFound
	}
	return &torrents[0], nil
}

func (qb *QbittorrentClient) GetTorrentProperties(hash string) (*TorrentProperties, error) {
	resp, err := qb.client.R().SetQueryParam("hash", hash).Get(qb.endpoint.JoinPath("/api/v2/torrents/properties").String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrTorrentNotFound
	}
	var torr TorrentProperties
	err = json.Unmarshal(resp.Body(), &torr)
	return &torr, err
}

func (qb *QbittorrentClient) PauseTorrents(hashes []string) error {
	params := map[string]string{
		"hashes": strings.Join(hashes, "|"),
	}
	_, err := qb.client.R().SetFormData(params).Post(qb.endpoint.JoinPath("/api/v2/torrents/pause").String())
	if err != nil {
		return err
	}
	return nil
}

func (qb *QbittorrentClient) PauseAll() error {
	return qb.PauseTorrents([]string{"all"})
}

func (qb *QbittorrentClient) ResumeTorrents(hashes []string) error {
	params := map[string]string{
		"hashes": strings.Join(hashes, "|"),
	}
	_, err := qb.client.R().SetFormData(params).Post(qb.endpoint.JoinPath("/api/v2/torrents/resume").String())
	if err != nil {
		return err
	}
	return nil
}

func (qb *QbittorrentClient) DeleteTorrents(hashes []string, deleteFiles bool) error {
	params := map[string]string{
		"hashes":      strings.Join(hashes, "|"),
		"deleteFiles": strconv.FormatBool(deleteFiles),
	}
	_, err := qb.client.R().
		SetFormData(params).
		Post(qb.endpoint.JoinPath("/api/v2/torrents/delete").String())
	if err != nil {
		return err
	}
	return nil
}

func (qb *QbittorrentClient) RenameFile(hash string, oldPath string, newPath string) error {
	params := map[string]string{
		"hash":    hash,
		"oldPath": oldPath,
		"newPath": newPath,
	}
	resp, err := qb.client.R().SetFormData(params).Post(qb.endpoint.JoinPath("/api/v2/torrents/renameFile").String())
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusBadRequest {
		return ErrMissingParameter
	}
	if resp.StatusCode() == http.StatusConflict {
		return ErrPath
	}
	return nil
}

func (qb *QbittorrentClient) RenameFolder(hash string, oldPath string, newPath string) error {
	params := map[string]string{
		"hash":    hash,
		"oldPath": oldPath,
		"newPath": newPath,
	}
	resp, err := qb.client.R().SetFormData(params).Post(qb.endpoint.JoinPath("/api/v2/torrents/renameFolder").String())
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusBadRequest {
		return ErrMissingParameter
	}
	if resp.StatusCode() == http.StatusConflict {
		return ErrPath
	}
	return nil
}

func (qb *QbittorrentClient) GetTorrentContent(hash string, indexes []int64) ([]TorrentContent, error) {
	var indexArray []string
	for _, idx := range indexes {
		indexArray = append(indexArray, strconv.FormatInt(idx, 10))
	}
	params := map[string]string{
		"hash": hash,
	}
	var contents []TorrentContent
	resp, err := qb.client.R().SetQueryParams(params).SetResult(&contents).Get(qb.endpoint.JoinPath("/api/v2/torrents/files").String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusBadRequest {
		return nil, ErrTorrentNotFound
	}
	return contents, nil
}

func (qb *QbittorrentClient) SetFilePriority(hash string, indexes []int, priority int64) error {
	params := map[string]string{
		"hash":     hash,
		"priority": strconv.FormatInt(priority, 10),
	}
	var indexArray []string
	for _, idx := range indexes {
		indexArray = append(indexArray, strconv.FormatInt(int64(idx), 10))
	}
	if len(indexArray) != 0 {
		params["id"] = strings.Join(indexArray, "|")
	}
	resp, err := qb.client.R().SetFormData(params).Post(qb.endpoint.JoinPath("/api/v2/torrents/filePrio").String())
	if err != nil {
		return err
	}
	switch resp.StatusCode() {
	case http.StatusBadRequest:
		return errors.New("invalid priority")
	case http.StatusNotFound:
		return ErrTorrentNotFound
	case http.StatusConflict:
		return errors.New("torrent metadata hasn't downloaded yet")
	}
	return nil
}

func (qb *QbittorrentClient) WaitForDownloadComplete(hash string, period time.Duration, callback func() bool) error {
	return qb.WatchTorrent(hash, period, func(torr *Torrent) bool {
		if torr.CompletionOn != 0 {
			return callback()
		}
		return false
	})
}

func (qb *QbittorrentClient) WatchTorrent(hash string, period time.Duration, callback func(torr *Torrent) bool) error {
	ticker := time.NewTicker(period)
	for range ticker.C {
		torr, err := qb.GetTorrent(hash)
		if err == nil {
			if callback(torr) {
				break
			}
		} else if err == ErrTorrentNotFound {
			return err
		}
	}
	return nil
}

func respToErr(resp []byte) error {
	if string(resp) == "Ok." {
		return nil
	}
	return errors.New(string(resp))
}
