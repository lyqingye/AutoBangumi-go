package moe

import (
	bangumitypes "autobangumi-go/bangumi"
)

func (parser *BangumiMoe) Search(keyword string) (*bangumitypes.Bangumi, error) {
	return nil, nil
}

func (parser *BangumiMoe) SearchTorrentByTag(tags []string, page int) (*ResponseSearchTorrent, error) {
	req := RequestSearchTorrentByTags{
		TagID:   tags,
		PageNum: page,
	}
	result := ResponseSearchTorrent{}
	_, err := parser.http.R().SetResult(&result).SetBody(&req).Post(parser.endpoint.JoinPath("/api/torrent/search").String())
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (parser *BangumiMoe) SearchTagByKeyword(keyword string) (*ResponseSearchTag, error) {
	req := RequestSearchTagByKeyword{
		Name:     keyword,
		Keywords: true,
		Multi:    true,
	}
	result := ResponseSearchTag{}
	_, err := parser.http.R().SetResult(&result).SetBody(&req).Post(parser.endpoint.JoinPath("/api/tag/search").String())
	if err != nil {
		return nil, err
	}
	return &result, nil
}
