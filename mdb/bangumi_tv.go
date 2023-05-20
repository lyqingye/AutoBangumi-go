package mdb

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/antlabs/strsim"
	"github.com/go-resty/resty/v2"
)

type BangumiTVClient struct {
	http     *resty.Client
	endpoint *url.URL
}

type Subjects struct {
	Date     string `json:"date"`
	Platform string `json:"platform"`
	Images   struct {
		Small  string `json:"small"`
		Grid   string `json:"grid"`
		Large  string `json:"large"`
		Medium string `json:"medium"`
		Common string `json:"common"`
	} `json:"images"`
	Summary       string `json:"summary"`
	Name          string `json:"name"`
	NameCn        string `json:"name_cn"`
	TotalEpisodes int    `json:"total_episodes"`
	Collection    struct {
		OnHold  int `json:"on_hold"`
		Dropped int `json:"dropped"`
		Wish    int `json:"wish"`
		Collect int `json:"collect"`
		Doing   int `json:"doing"`
	} `json:"collection"`
	ID      int64 `json:"id"`
	Eps     int   `json:"eps"`
	Volumes int   `json:"volumes"`
	Locked  bool  `json:"locked"`
	Nsfw    bool  `json:"nsfw"`
	Type    int   `json:"type"`
	InfoBox []any `json:"infobox"`
}

func (sub *Subjects) GetAliasNames() []string {
	var result []string
	for _, ib := range sub.InfoBox {
		if kv, ok := ib.(map[string]interface{}); ok {
			if kv["key"] != "别名" {
				continue
			}
			value := kv["value"]
			if aliasName, isStr := value.(string); isStr {
				result = append(result, aliasName)
			}
			if aliasNames, isStrArray := value.([]any); isStrArray {
				for _, entity := range aliasNames {
					if vMap, isMap := entity.(map[string]interface{}); isMap {
						if name, isStr := vMap["v"].(string); isStr {
							result = append(result, name)
						}
					}
				}
			}
		}
	}
	return result
}

func NewBangumiTVClient(apiEndpoint string) (*BangumiTVClient, error) {
	endpoint, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, err
	}
	client := BangumiTVClient{
		http:     resty.New().SetHeader("User-Agent", "czy0729/Bangumi/6.4.0 (Android)"),
		endpoint: endpoint,
	}
	return &client, nil
}

func (client *BangumiTVClient) GetSubjects(id int64) (*Subjects, error) {
	apiUrl := client.endpoint.JoinPath(fmt.Sprintf("subjects/%d", id))
	sub := Subjects{}
	_, err := client.http.R().SetResult(&sub).Get(apiUrl.String())
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

type RequestSearchSubject struct {
	Keyword string `json:"keyword"`
	Filter  struct {
		Type []int `json:"type"`
	} `json:"filter"`
}

type PageResponse struct {
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
	Data   []json.RawMessage `json:"data"`
}

const (
	SubjectTypeAnime = 2
)

func (client *BangumiTVClient) SearchAnime(keyword string) (*Subjects, error) {
	apiUrl := client.endpoint.JoinPath("search/subjects")
	params := RequestSearchSubject{
		Keyword: keyword,
	}
	params.Filter.Type = append(params.Filter.Type, SubjectTypeAnime)
	queryParams := map[string]string{
		"limit":  "10",
		"offset": "0",
	}
	var pageResponse PageResponse
	_, err := client.http.R().SetBody(params).SetQueryParams(queryParams).SetResult(&pageResponse).Post(apiUrl.String())
	if err != nil {
		return nil, err
	}
	if pageResponse.Total == 0 {
		return nil, fmt.Errorf("bangumi tv search result is empty")
	}
	var subjects []Subjects
	var names []string
	for _, entity := range pageResponse.Data {
		var subject Subjects
		err = json.Unmarshal(entity, &subject)
		if err != nil {
			return nil, err
		}
		if subject.NameCn == keyword || subject.Name == keyword {
			return client.GetSubjects(subject.ID)
		}
		subjects = append(subjects, subject)
		names = append(names, subject.Name)
		names = append(names, subject.NameCn)
	}
	matchResult := strsim.FindBestMatch(keyword, names)

	for _, subject := range subjects {
		if subject.NameCn == matchResult.Match.S || subject.Name == matchResult.Match.S {
			return client.GetSubjects(subject.ID)
		}
	}

	return client.GetSubjects(subjects[0].ID)
}
