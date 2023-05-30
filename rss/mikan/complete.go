package mikan

import (
	bangumitypes "autobangumi-go/bangumi"
	"errors"
	tmdb "github.com/cyruzin/golang-tmdb"
)

func (parser *MikanRSSParser) CompleteBangumi(bangumi *bangumitypes.Bangumi) error {
	info := bangumi.Info
	if info.Title == "" {
		return errors.New("empty bangumi title")
	}

	var tv *tmdb.TVDetails
	var err error
	if info.TmDBId == 0 {
		tv, err = parser.searchTMDB(info.Title)
		if err != nil {
			return err
		}
	} else {
		tv, err = parser.getTMDB(info.TmDBId)
		if err != nil {
			return err
		}
	}

	bangumi.Info.TmDBId = tv.ID

	// complete season info
	if len(bangumi.Seasons) == 0 {
		bangumi.Seasons = make(map[uint]bangumitypes.Season)
	}

	allSeasonCollected := true
	for _, season := range tv.Seasons {
		if season.SeasonNumber == 0 {
			continue
		}
		existsSeason := bangumi.Seasons[uint(season.SeasonNumber)]
		existsSeason.Number = uint(season.SeasonNumber)
		existsSeason.EpCount = uint(season.EpisodeCount)
		bangumi.Seasons[uint(season.SeasonNumber)] = existsSeason

		if !existsSeason.IsEpisodesCollected() {
			allSeasonCollected = false
		}
	}

	if allSeasonCollected {
		return nil
	}

	for number, season := range bangumi.Seasons {
		if season.MikanBangumiId != "" {
			if season.IsEpisodesCollected() {
				continue
			}
			err = parser.completeSeasonByMikanBangumiId(info, &season)
			if err != nil {
				continue
			}
		}
		bangumi.Seasons[number] = season
	}

	searchResult, err := parser.Search2(info.Title)
	if err != nil {
		subject, err := parser.searchBangumiTV(info.Title)
		if err != nil {
			return err
		}
		searchResult, err = parser.Search2(subject.NameCn)
		if err != nil {
			return err
		}
	}

	return parser.completeBangumiBySearchResult(bangumi, searchResult)
}

func (parser *MikanRSSParser) completeBangumiBySearchResult(bangumi *bangumitypes.Bangumi, searchResult *bangumitypes.Bangumi) error {
	if searchResult.Info.TmDBId != bangumi.Info.TmDBId {
		return errors.New("complete bangumi error, search result empty")
	}

	// complete season
	for seasonNumber, source := range searchResult.Seasons {
		dest := bangumi.Seasons[seasonNumber]
		if dest.IsEpisodesCollected() {
			continue
		}
		parser.mergeSeasonInfo(&dest, &source)
		bangumi.Seasons[seasonNumber] = dest
	}
	return nil
}

func (parser *MikanRSSParser) mergeSeasonInfo(dest *bangumitypes.Season, source *bangumitypes.Season) {
	if dest.SubjectId != 0 && source.SubjectId != 0 && dest.SubjectId != source.SubjectId {
		return
	}

	if dest.MikanBangumiId != "" && source.MikanBangumiId != "" && dest.MikanBangumiId != source.MikanBangumiId {
		return
	}

	if dest.SubjectId == 0 {
		dest.SubjectId = source.SubjectId
	}

	if dest.MikanBangumiId == "" {
		dest.MikanBangumiId = source.MikanBangumiId
	}
	if dest.EpCount == 0 {
		dest.EpCount = source.EpCount
	}
	if dest.Number == 0 {
		dest.EpCount = source.EpCount
	}

	for _, searchEpisode := range source.Episodes {
		isParsed := false
		for _, episode := range dest.Episodes {
			if episode.Number == searchEpisode.Number {
				isParsed = true
				break
			}
		}
		if !isParsed {
			dest.Episodes = append(dest.Episodes, searchEpisode)
		}

		for i, episode := range dest.Episodes {
			if episode.Number == searchEpisode.Number {
				if episode.CanReplace(&searchEpisode) {
					dest.Episodes[i] = searchEpisode
					dest.RemoveComplete(episode.Number)
				}
				break
			}
		}
	}
}

func (parser *MikanRSSParser) completeSeasonByMikanBangumiId(bangumiInfo bangumitypes.BangumiInfo, season *bangumitypes.Season) error {
	bgm, err := parser.Search3(season.MikanBangumiId)
	if err != nil {
		return err
	}
	if bgm.Info.TmDBId == bangumiInfo.TmDBId {
		if source, found := bgm.Seasons[season.Number]; found && source.MikanBangumiId == season.MikanBangumiId {
			parser.mergeSeasonInfo(season, &source)
		}
	}

	return nil
}
