package mapper

import (
	"encoding/json"
	"strconv"

	"wavearchive/internal/domain"
	"wavearchive/internal/normalizer"
	"wavearchive/internal/sources/nanoka/dto"
)

func EchoIndex(id string, source dto.EchoIndexEntry, version string) (domain.Echo, bool) {
	numericID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return domain.Echo{}, false
	}
	rarities, _ := json.Marshal(source.Ranks)
	sonatas, _ := json.Marshal(source.Groups)
	return domain.Echo{
		ID: numericID, Name: source.Name, Code: source.Code,
		Cost: echoCost(source.Intensity), IconPath: source.Icon,
		RaritiesJSON: string(rarities), SonataIDsJSON: string(sonatas),
		GameVersion: version,
	}, true
}

func MergeEchoDetail(base domain.Echo, source dto.EchoDetail) domain.Echo {
	base.Name = source.Name
	base.Code = source.Code
	base.Type = source.Type
	base.Class = source.Intensity
	base.Cost = echoCost(source.IntensityCode)
	base.Place = source.Place
	base.IconPath = source.Icon
	description := source.Skill.Description
	if description == "" {
		description = source.Skill.SimpleDesc
	}
	params := source.Skill.Params
	if len(params) > 0 {
		if row, ok := params[0].([]any); ok {
			params = row
		}
	}
	description, _ = normalizer.ApplyParams(description, params)
	base.Skill = normalizer.CleanText(description)
	rarities, _ := json.Marshal(source.Rarity)
	base.RaritiesJSON = string(rarities)
	ids := make([]int64, 0, len(source.Group))
	for id := range source.Group {
		if numeric, err := strconv.ParseInt(id, 10, 64); err == nil {
			ids = append(ids, numeric)
		}
	}
	sonatas, _ := json.Marshal(ids)
	base.SonataIDsJSON = string(sonatas)
	return base
}

func Sonata(source dto.SonataIndexEntry, version string) domain.Sonata {
	var twoPiece, fivePiece string
	if languages, ok := source.Set["2"]; ok {
		if effect, ok := languages["en"]; ok {
			twoPiece, _ = normalizer.ApplyParams(effect.Description, effect.Params)
		}
	}
	if languages, ok := source.Set["5"]; ok {
		if effect, ok := languages["en"]; ok {
			fivePiece, _ = normalizer.ApplyParams(effect.Description, effect.Params)
		}
	}
	return domain.Sonata{
		ID: source.ID, Name: source.Name.English, IconPath: source.Icon,
		TwoPiece: normalizer.CleanText(twoPiece), FivePiece: normalizer.CleanText(fivePiece),
		GameVersion: version,
	}
}

func echoCost(intensity int) int {
	switch intensity {
	case 0:
		return 1
	case 1:
		return 3
	default:
		return 4
	}
}
