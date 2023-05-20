package mikan

import "encoding/xml"

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
