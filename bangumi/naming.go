package bangumi

import (
	"fmt"
	"path/filepath"
	"strings"
)

var (
	MediaResource = map[string]interface{}{
		".mp4":  nil,
		".mkv":  nil,
		".avi":  nil,
		".flv":  nil,
		".rmvb": nil,
		".wmv":  nil,
	}
	AssResource = []string{
		".sc.ass",
		".tc.ass",
		".zh.ass",
		".cn.ass",
		".ass",
		".srt",
		".stl",
		".sbv",
		".webvtt",
		".dfxp",
		".ttml",
	}

	SubTitleLangKeyword = map[string]string{
		"CHT":     SubtitleCht,
		"CHS":     SubtitleChs,
		"简繁":      SubtitleCht,
		"简体":      SubtitleChs,
		"繁体":      SubtitleCht,
		"简中":      SubtitleChs,
		"繁中":      SubtitleCht,
		"简日双语":    SubtitleChs,
		"繁日双语":    SubtitleCht,
		"简繁内封字幕":  SubtitleChs,
		"简中内嵌":    SubtitleChs,
		"繁中内嵌":    SubtitleCht,
		"简繁日内封字幕": SubtitleChs,
		"BIG5":    SubtitleCht,
		"GB":      SubtitleChs,
		"BIG5_GB": SubtitleChs,
	}
)

func DirNaming(info *BangumiInfo, seasonNum uint) string {
	return filepath.Join(info.Title, fmt.Sprintf("Season %02d", seasonNum))
}

func RenamingEpisodeFileName(info *BangumiInfo, seasonNum uint, ep *Episode, filename string) string {
	newName := fmt.Sprintf("[%s] S%02dE%02d", info.Title, seasonNum, ep.Number)
	ext := filepath.Ext(filename)
	if ext == "" {
		return newName
	}

	for keyword := range MediaResource {
		if strings.HasSuffix(filename, keyword) {
			return fmt.Sprintf("%s%s", newName, keyword)
		}
	}

	for _, extension := range AssResource {
		if strings.HasSuffix(filename, extension) {
			// Subtitle Resource, try predict lang
			for keyword, lang := range SubTitleLangKeyword {
				if strings.Contains(filename, keyword) {
					return fmt.Sprintf("%s.%s%s", newName, strings.ToLower(lang), extension)
				}
			}
			return fmt.Sprintf("%s%s", newName, extension)
		}
	}
	return ""
}
