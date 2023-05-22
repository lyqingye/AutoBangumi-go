package mikan

import (
	"errors"
	"fmt"
	bangumitypes "pikpak-bot/bangumi"
)

func (parser *MikanRSSParser) CompleteBangumi(bangumi *bangumitypes.Bangumi) error {
	info := bangumi.Info
	if info.Title == "" {
		return errors.New("empty bangumi title")
	}

	if info.TmDBId == 0 {
		tv, err := parser.searchTMDB(info.Title)
		if err != nil {
			return err
		}
		bangumi.Info.TmDBId = tv.ID
		if len(bangumi.Seasons) == 0 {
			bangumi.Seasons = make(map[uint]bangumitypes.Season)
		}

		for _, season := range tv.Seasons {
			if season.SeasonNumber == 0 {
				continue
			}
			existsSeason := bangumi.Seasons[uint(season.SeasonNumber)]
			existsSeason.Number = uint(season.SeasonNumber)
			existsSeason.EpCount = uint(season.EpisodeCount)
			bangumi.Seasons[uint(season.SeasonNumber)] = existsSeason
		}
	}

	for _, season := range bangumi.Seasons {
		if season.MikanBangumiId != "" {
			err := parser.completeSeasonByMikanBangumiId(info, &season)
			if err != nil {
				continue
			}
		}
	}

	searchResult, err := parser.Search(info.Title)
	if err != nil {
		return err
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
		if dest.IsComplete(searchEpisode.Number) {
			continue
		}

		isParsed := false
		for _, episode := range dest.Episodes {
			if episode.Number == searchEpisode.Number {
				isParsed = true
				// if searchEpisode.Compare(&episode) {
				// 	dest.Episodes[i] = searchEpisode
				// 	break
				// }
				break
			}
		}
		if !isParsed {
			dest.Episodes = append(dest.Episodes, searchEpisode)
		}
	}
}

func (parser *MikanRSSParser) completeSeasonByMikanBangumiId(bangumiInfo bangumitypes.BangumiInfo, season *bangumitypes.Season) error {
	mikanRss, err := parser.getRss(parser.mikanEndpoint.JoinPath(fmt.Sprintf("RSS/Bangumi?bangumiId=%s", season.MikanBangumiId)).String())
	if err != nil {
		return err
	}

	multiResults, err := parser.parseMikanRSS(mikanRss)
	if err != nil {
		return err
	}

	for _, bangumi := range multiResults {
		if bangumi.Info.TmDBId == bangumiInfo.TmDBId {
			if source, found := bangumi.Seasons[season.Number]; found {
				parser.mergeSeasonInfo(season, &source)
			}
		}
	}

	return nil
}
