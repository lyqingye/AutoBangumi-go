package bangumi

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
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

	SubTitleLangKeyword = map[string]SubtitleLang{
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
		"简日内嵌":    SubtitleChs,
		"简体内嵌":    SubtitleChs,
		"繁中内嵌":    SubtitleCht,
		"繁体内嵌":    SubtitleCht,
		"繁日内嵌":    SubtitleCht,
		"简繁日内封字幕": SubtitleChs,
		"BIG5":    SubtitleCht,
		"GB":      SubtitleChs,
		"BIG5_GB": SubtitleChs,
	}
)

func DirNaming(info Bangumi, seasonNum uint) string {
	return filepath.Join(info.GetTitle(), fmt.Sprintf("SeasonNum %02d", seasonNum))
}

func ParseDirName(dirname string) (uint, error) {
	season := strings.ReplaceAll(dirname, "SeasonNum", "")
	season = strings.TrimSpace(season)
	seasonNum, err := strconv.ParseUint(season, 10, 32)
	return uint(seasonNum), err
}

func RenamingEpisodeFileName(info Bangumi, seasonNum uint, epNum uint, filename string) string {
	newName := fmt.Sprintf("[%s] S%02dE%02d", info.GetTitle(), seasonNum, epNum)
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
					return fmt.Sprintf("%s.%s%s", newName, strings.ToLower(string(lang)), extension)
				}
			}
			return fmt.Sprintf("%s%s", newName, extension)
		}
	}
	return ""
}

func ParseEpisodeFilename(filename string) (season uint, episode uint, err error) {
	pattern := `S(\d{2})E(\d{2})`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(filename)
	var number uint64
	if len(matches) == 3 {
		number, err = strconv.ParseUint(matches[1], 10, 32)
		if err != nil {
			return
		}
		season = uint(number)
		number, err = strconv.ParseUint(matches[2], 10, 32)
		if err != nil {
			return
		}
		episode = uint(number)
	} else {
		err = fmt.Errorf("faild to parse episode file name: %s", filename)
	}
	return
}
