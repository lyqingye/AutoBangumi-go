package moe

import "time"

type Torrent struct {
	ID            string     `json:"_id"`
	CategoryTagID string     `json:"category_tag_id"`
	Title         string     `json:"title"`
	Introduction  string     `json:"introduction"`
	TagIds        []string   `json:"tag_ids"`
	Comments      int        `json:"comments"`
	Downloads     int        `json:"downloads"`
	Finished      int        `json:"finished"`
	Leechers      int        `json:"leechers"`
	Seeders       int        `json:"seeders"`
	UploaderID    string     `json:"uploader_id"`
	TeamID        string     `json:"team_id"`
	PublishTime   time.Time  `json:"publish_time"`
	Magnet        string     `json:"magnet"`
	InfoHash      string     `json:"infoHash"`
	FileID        string     `json:"file_id"`
	Teamsync      bool       `json:"teamsync"`
	Content       [][]string `json:"content"`
	Size          string     `json:"size"`
	Btskey        string     `json:"btskey"`
	Sync          struct {
		Dmhy   string `json:"dmhy"`
		Acgrip string `json:"acgrip"`
		Acgnx  string `json:"acgnx"`
	} `json:"sync"`
}

type RequestSearchTorrentByTags struct {
	TagID   []string `json:"tag_id"`
	PageNum int      `json:"p"`
}

type ResponseSearchTorrent struct {
	Torrents  []Torrent `json:"torrents"`
	Count     int       `json:"count"`
	PageCount int       `json:"page_count"`
}

type RequestSearchTagByKeyword struct {
	Name     string `json:"name"`
	Keywords bool   `json:"keywords"`
	Multi    bool   `json:"multi"`
}

type ResponseSearchTag struct {
	Success bool `json:"success"`
	Found   bool `json:"found"`
	Tag     []struct {
		ID       string   `json:"_id"`
		Name     string   `json:"name"`
		Type     string   `json:"type"`
		Synonyms []string `json:"synonyms"`
		Locale   struct {
			Ja   string `json:"ja"`
			ZhTw string `json:"zh_tw"`
			En   string `json:"en"`
			ZhCn string `json:"zh_cn"`
		} `json:"locale"`
		SynLowercase []string `json:"syn_lowercase"`
		Activity     int      `json:"activity"`
	} `json:"tag"`
}
