package mdb

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/antlabs/strsim"
	"github.com/go-resty/resty/v2"
)

type BangumiTVClient struct {
	http        *resty.Client
	endpoint    *url.URL
	username    string
	accessToken string
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

type ListResponse struct {
	Results int           `json:"results"`
	List    []SubjectItem `json:"list"`
}

type SubjectItem struct {
	ID         int64  `json:"id"`
	URL        string `json:"url"`
	Type       int    `json:"type"`
	Name       string `json:"name"`
	NameCn     string `json:"name_cn"`
	Summary    string `json:"summary"`
	Eps        int    `json:"eps,omitempty"`
	EpsCount   int    `json:"eps_count,omitempty"`
	AirDate    string `json:"air_date"`
	AirWeekday int    `json:"air_weekday"`
	Images     struct {
		Large  string `json:"large"`
		Common string `json:"common"`
		Medium string `json:"medium"`
		Small  string `json:"small"`
		Grid   string `json:"grid"`
	} `json:"images"`
	Collection struct {
		Wish    int `json:"wish"`
		Collect int `json:"collect"`
		Doing   int `json:"doing"`
		OnHold  int `json:"on_hold"`
		Dropped int `json:"dropped"`
	} `json:"collection"`
}

func (client *BangumiTVClient) SearchAnime2(keyword string) (*Subjects, error) {
	host := strings.ReplaceAll(client.endpoint.String(), client.endpoint.Path, "")
	apiUrl := host + "/search/subject/" + url.QueryEscape(keyword)

	queryParams := map[string]string{
		"max_results":   "10",
		"start":         "0",
		"type":          strconv.FormatInt(SubjectTypeAnime, 10),
		"responseGroup": "large",
	}
	var listResponse ListResponse
	_, err := client.http.R().SetQueryParams(queryParams).SetResult(&listResponse).Get(apiUrl)
	if err != nil {
		return nil, err
	}
	if listResponse.Results == 0 {
		return nil, fmt.Errorf("bangumi tv search result is empty")
	}
	var subjects []SubjectItem
	var names []string
	for _, item := range listResponse.List {
		if item.NameCn == keyword || item.Name == keyword {
			return client.GetSubjects(item.ID)
		}

		subjects = append(subjects, item)
		names = append(names, item.Name)
		names = append(names, item.NameCn)
	}
	matchResult := strsim.FindBestMatch(keyword, names)

	for _, subject := range subjects {
		if subject.NameCn == matchResult.Match.S || subject.Name == matchResult.Match.S {
			return client.GetSubjects(subject.ID)
		}
	}

	return client.GetSubjects(subjects[0].ID)
}

type MeInfo struct {
	Avatar struct {
		Large  string `json:"large"`
		Medium string `json:"medium"`
		Small  string `json:"small"`
	} `json:"avatar"`
	Sign      string `json:"sign"`
	URL       string `json:"url"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	ID        int    `json:"id"`
	UserGroup int    `json:"user_group"`
}

func (client *BangumiTVClient) Me() (*MeInfo, error) {
	result := MeInfo{}
	_, err := client.http.R().
		SetResult(&result).
		SetAuthToken(client.accessToken).
		Get(client.endpoint.JoinPath("me").String())
	if err != nil {
		return &result, err
	}
	return &result, nil
}

const (
	CollectionTypeWish    = 1
	CollectionTypeCollect = 2
	CollectionTypeDoing   = 3
	CollectionTypeOnHold  = 4
	CollectionTypeDropped = 5
)

type CollectionItem struct {
	UpdatedAt time.Time     `json:"updated_at"`
	Comment   interface{}   `json:"comment"`
	Tags      []interface{} `json:"tags"`
	Subject   struct {
		Date   string `json:"date"`
		Images struct {
			Small  string `json:"small"`
			Grid   string `json:"grid"`
			Large  string `json:"large"`
			Medium string `json:"medium"`
			Common string `json:"common"`
		} `json:"images"`
		Name         string `json:"name"`
		NameCn       string `json:"name_cn"`
		ShortSummary string `json:"short_summary"`
		Tags         []struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		} `json:"tags"`
		Score           float64 `json:"score"`
		Type            int     `json:"type"`
		ID              int     `json:"id"`
		Eps             int     `json:"eps"`
		Volumes         int     `json:"volumes"`
		CollectionTotal int     `json:"collection_total"`
		Rank            int     `json:"rank"`
	} `json:"subject"`
	SubjectID   int  `json:"subject_id"`
	VolStatus   int  `json:"vol_status"`
	EpStatus    int  `json:"ep_status"`
	SubjectType int  `json:"subject_type"`
	Type        int  `json:"type"`
	Rate        int  `json:"rate"`
	Private     bool `json:"private"`
}

func (client *BangumiTVClient) SetAccessToken(accessToken string) error {
	client.accessToken = accessToken
	meInfo, err := client.Me()
	if err != nil {
		return err
	}
	client.username = meInfo.Username
	return nil
}

func (client *BangumiTVClient) Collections(collectionType int, subjectType int) ([]*CollectionItem, error) {
	pageSize := 100
	offset := 0
	var result []*CollectionItem
	for {
		params := map[string]string{
			"subject_type": strconv.FormatInt(int64(subjectType), 10),
			"type":         strconv.FormatInt(int64(collectionType), 10),
			"limit":        strconv.FormatInt(int64(pageSize), 10),
			"offset":       strconv.FormatInt(int64(offset), 10),
		}
		pageResponse := PageResponse{}
		_, err := client.http.R().
			SetResult(&pageResponse).
			SetAuthToken(client.accessToken).
			SetQueryParams(params).
			Get(client.endpoint.JoinPath("users", client.username, "collections").String())
		if err != nil {
			return nil, err
		}
		for _, itemRaw := range pageResponse.Data {
			item := CollectionItem{}
			err = json.Unmarshal(itemRaw, &item)
			if err != nil {
				return nil, err
			}
			result = append(result, &item)
		}
		itemLen := len(pageResponse.Data)
		if itemLen < pageSize {
			break
		}
		offset = offset + itemLen
	}
	return result, nil
}

type Calendar []struct {
	Weekday struct {
		En string `json:"en"`
		Cn string `json:"cn"`
		Ja string `json:"ja"`
		ID int    `json:"id"`
	} `json:"weekday"`
	Items []struct {
		ID         int    `json:"id"`
		URL        string `json:"url"`
		Type       int    `json:"type"`
		Name       string `json:"name"`
		NameCn     string `json:"name_cn"`
		Summary    string `json:"summary"`
		AirDate    string `json:"air_date"`
		AirWeekday int    `json:"air_weekday"`
		Rating     struct {
			Total int `json:"total"`
			Count struct {
				Num1  int `json:"1"`
				Num2  int `json:"2"`
				Num3  int `json:"3"`
				Num4  int `json:"4"`
				Num5  int `json:"5"`
				Num6  int `json:"6"`
				Num7  int `json:"7"`
				Num8  int `json:"8"`
				Num9  int `json:"9"`
				Num10 int `json:"10"`
			} `json:"count"`
			Score float64 `json:"score"`
		} `json:"rating,omitempty"`
		Rank   int `json:"rank,omitempty"`
		Images struct {
			Large  string `json:"large"`
			Common string `json:"common"`
			Medium string `json:"medium"`
			Small  string `json:"small"`
			Grid   string `json:"grid"`
		} `json:"images"`
		Collection struct {
			Doing int `json:"doing"`
		} `json:"collection,omitempty"`
	} `json:"items"`
}

func (client *BangumiTVClient) GetCalendar() (*Calendar, error) {
	host := strings.ReplaceAll(client.endpoint.String(), client.endpoint.Path, "")
	result := Calendar{}
	_, err := client.http.R().
		SetResult(&result).
		Get(fmt.Sprintf("%s/calendar", host))
	if err != nil {
		return nil, err
	}
	return &result, nil
}
