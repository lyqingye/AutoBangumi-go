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
	AssResource = map[string]interface{}{
		".ass":    nil,
		".srt":    nil,
		".stl":    nil,
		".sbv":    nil,
		".webvtt": nil,
		".dfxp":   nil,
		".ttml":   nil,
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

func DirNaming(b *Bangumi) string {
	return filepath.Join(b.Title, fmt.Sprintf("Season %02d", b.Season))
}

func RenamingEpisodeFileName(ep *Episode, filename string) string {
	newName := fmt.Sprintf("[%s] S%02dE%02d", ep.BangumiTitle, ep.Season, ep.EPNumber)
	ext := filepath.Ext(filename)
	if ext == "" {
		return newName
	}
	if isBangumiMediaResource(ext) {
		return fmt.Sprintf("%s%s", newName, ext)
	}
	if isASSResource(ext) {
		for keyword, lang := range SubTitleLangKeyword {
			if strings.Contains(filename, keyword) {
				return fmt.Sprintf("%s.%s%s", newName, strings.ToLower(lang), ext)
			}
		}
	}
	return ""
}

func isBangumiMediaResource(extension string) bool {
	_, found := MediaResource[strings.ToLower(extension)]
	return found
}

func isASSResource(extension string) bool {
	_, found := AssResource[strings.ToLower(extension)]
	return found
}
