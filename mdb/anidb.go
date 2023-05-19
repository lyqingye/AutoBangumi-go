package mdb

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

type AnimeTitles struct {
	XMLName xml.Name `xml:"animetitles"`
	Text    string   `xml:",chardata"`
	Anime   []struct {
		Text  string `xml:",chardata"`
		Aid   string `xml:"aid,attr"`
		Title []struct {
			Text string `xml:",chardata"`
			Lang string `xml:"lang,attr"`
			Type string `xml:"type,attr"`
		} `xml:"title"`
	} `xml:"anime"`
}

type Anime struct {
	XMLName      xml.Name `xml:"anime"`
	Text         string   `xml:",chardata"`
	ID           string   `xml:"id,attr"`
	Restricted   string   `xml:"restricted,attr"`
	Type         string   `xml:"type"`
	Episodecount string   `xml:"episodecount"`
	Startdate    string   `xml:"startdate"`
	Titles       struct {
		Text  string `xml:",chardata"`
		Title []struct {
			Text string `xml:",chardata"`
			Lang string `xml:"lang,attr"`
			Type string `xml:"type,attr"`
		} `xml:"title"`
	} `xml:"titles"`
	Relatedanime struct {
		Text  string `xml:",chardata"`
		Anime struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
			Type string `xml:"type,attr"`
		} `xml:"anime"`
	} `xml:"relatedanime"`
	URL      string `xml:"url"`
	Creators struct {
		Text string `xml:",chardata"`
		Name []struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
			Type string `xml:"type,attr"`
		} `xml:"name"`
	} `xml:"creators"`
	Description string `xml:"description"`
	Ratings     struct {
		Text      string `xml:",chardata"`
		Permanent struct {
			Text  string `xml:",chardata"`
			Count string `xml:"count,attr"`
		} `xml:"permanent"`
		Temporary struct {
			Text  string `xml:",chardata"`
			Count string `xml:"count,attr"`
		} `xml:"temporary"`
	} `xml:"ratings"`
	Picture   string `xml:"picture"`
	Resources struct {
		Text     string `xml:",chardata"`
		Resource []struct {
			Text           string `xml:",chardata"`
			Type           string `xml:"type,attr"`
			Externalentity struct {
				Text       string `xml:",chardata"`
				Identifier string `xml:"identifier"`
				URL        string `xml:"url"`
			} `xml:"externalentity"`
		} `xml:"resource"`
	} `xml:"resources"`
	Tags struct {
		Text string `xml:",chardata"`
		Tag  []struct {
			Text          string `xml:",chardata"`
			ID            string `xml:"id,attr"`
			Parentid      string `xml:"parentid,attr"`
			Infobox       string `xml:"infobox,attr"`
			Weight        string `xml:"weight,attr"`
			Localspoiler  string `xml:"localspoiler,attr"`
			Globalspoiler string `xml:"globalspoiler,attr"`
			Verified      string `xml:"verified,attr"`
			Update        string `xml:"update,attr"`
			Name          string `xml:"name"`
			Description   string `xml:"description"`
			Picurl        string `xml:"picurl"`
		} `xml:"tag"`
	} `xml:"tags"`
	Characters struct {
		Text      string `xml:",chardata"`
		Character []struct {
			Text   string `xml:",chardata"`
			ID     string `xml:"id,attr"`
			Type   string `xml:"type,attr"`
			Update string `xml:"update,attr"`
			Rating struct {
				Text  string `xml:",chardata"`
				Votes string `xml:"votes,attr"`
			} `xml:"rating"`
			Name          string `xml:"name"`
			Gender        string `xml:"gender"`
			Charactertype struct {
				Text string `xml:",chardata"`
				ID   string `xml:"id,attr"`
			} `xml:"charactertype"`
			Description string `xml:"description"`
			Picture     string `xml:"picture"`
			Seiyuu      struct {
				Text    string `xml:",chardata"`
				ID      string `xml:"id,attr"`
				Picture string `xml:"picture,attr"`
			} `xml:"seiyuu"`
		} `xml:"character"`
	} `xml:"characters"`
	Episodes struct {
		Text    string `xml:",chardata"`
		Episode []struct {
			Text   string `xml:",chardata"`
			ID     string `xml:"id,attr"`
			Update string `xml:"update,attr"`
			Epno   struct {
				Text string `xml:",chardata"`
				Type string `xml:"type,attr"`
			} `xml:"epno"`
			Length  string `xml:"length"`
			Airdate string `xml:"airdate"`
			Rating  struct {
				Text  string `xml:",chardata"`
				Votes string `xml:"votes,attr"`
			} `xml:"rating"`
			Title []struct {
				Text string `xml:",chardata"`
				Lang string `xml:"lang,attr"`
			} `xml:"title"`
			Resources struct {
				Text     string `xml:",chardata"`
				Resource struct {
					Text           string `xml:",chardata"`
					Type           string `xml:"type,attr"`
					Externalentity struct {
						Text       string `xml:",chardata"`
						Identifier string `xml:"identifier"`
					} `xml:"externalentity"`
				} `xml:"resource"`
			} `xml:"resources"`
			Summary string `xml:"summary"`
		} `xml:"episode"`
	} `xml:"episodes"`
}

const (
	CacheFileName = "anidb.cache"
)

type AniDBClient struct {
	clientName string
	clientVer  string
	cacheDir   string

	memCache  *AnimeTitles
	cacheLock sync.Mutex
	http      *resty.Client
	endpoint  *url.URL
}

func NewAniDBClient(cacheDir string, clientName string, clientVer string) (*AniDBClient, error) {
	_, err := os.Stat(cacheDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(cacheDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	endpoint, err := url.Parse("http://api.anidb.net:9001/")
	if err != nil {
		return nil, err
	}
	client := AniDBClient{
		cacheDir:   cacheDir,
		memCache:   &AnimeTitles{},
		cacheLock:  sync.Mutex{},
		http:       resty.New(),
		endpoint:   endpoint,
		clientName: clientName,
		clientVer:  clientVer,
	}

	return &client, nil
}

func (ani *AniDBClient) SearchTitle(targetTitle string) (string, error) {
	for _, ani := range ani.memCache.Anime {
		for _, title := range ani.Title {
			if strings.Contains(title.Text, targetTitle) {
				return ani.Aid, nil
			}
		}
	}
	return "", errors.New("not found")
}

func (ani *AniDBClient) GetAnime(aniId string) (*Anime, error) {
	params := map[string]string{
		"request":   "anime",
		"aid":       aniId,
		"client":    ani.clientName,
		"clientver": ani.clientVer,
		"protover":  "1",
	}
	anime := Anime{}
	resp, err := ani.http.R().SetQueryParams(params).Get(ani.endpoint.JoinPath("httpapi").String())
	if err != nil {
		return nil, err
	}
	return &anime, xml.Unmarshal(resp.Body(), &anime)
}

func (ani *AniDBClient) InitDumpData() error {
	return ani.refreshDumpData(false)
}

func (ani *AniDBClient) AutoRefreshDumpData() {
	ticker := time.NewTicker(time.Hour * 24)
	for range ticker.C {
		_ = ani.refreshDumpData(true)
	}
}

func (ani *AniDBClient) refreshDumpData(forced bool) error {
	ani.cacheLock.Lock()
	defer ani.cacheLock.Unlock()
	cacheFilePath := filepath.Join(ani.cacheDir, CacheFileName)
	_, err := os.Stat(cacheFilePath)
	var xmlBz []byte
	if os.IsNotExist(err) || forced {
		xmlBz, err = ani.downloadDumpData()
		if err != nil {
			return err
		}
		err = os.WriteFile(cacheFilePath, xmlBz, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		xmlBz, err = os.ReadFile(cacheFilePath)
		if err != nil {
			return err
		}
	}
	newData := AnimeTitles{}
	err = xml.Unmarshal(xmlBz, &newData)
	if err != nil {
		return err
	}
	ani.memCache = &newData
	return nil
}

func (ani *AniDBClient) downloadDumpData() ([]byte, error) {
	resp, err := ani.http.R().Get("https://anidb.net/api/animetitles.xml.gz")
	if err != nil {
		return nil, err
	}
	bz := resp.Body()
	reader, err := gzip.NewReader(bytes.NewBuffer(bz))
	return io.ReadAll(reader)
}
