package mapper

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"wavearchive/internal/domain"
	"wavearchive/internal/normalizer"
	"wavearchive/internal/sources/nanoka/dto"
)

func CharacterIndex(id string, source dto.CharacterIndexEntry, version string) (domain.CharacterProfile, bool) {
	numericID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return domain.CharacterProfile{}, false
	}
	name := source.Name
	if name == "" {
		name = source.Nickname
	}
	if name == "" {
		name = "Rover Variant"
	}
	return domain.CharacterProfile{
		Character: domain.Character{
			ID:             numericID,
			Name:           name,
			Nickname:       source.Nickname,
			Rarity:         source.Rank,
			Element:        domain.Element(source.Element),
			ElementName:    domain.Element(source.Element).String(),
			WeaponType:     source.Weapon,
			WeaponTypeName: weaponTypeName(source.Weapon),
			IconPath:       source.Icon,
			BackgroundPath: source.Background,
			Gender:         source.Gender,
			GameVersion:    version,
			APIOrder:       source.APIOrder,
		},
		Skills: []domain.Skill{},
		Chains: []domain.ResonanceChain{},
	}, true
}

func MergeCharacterDetail(profile domain.CharacterProfile, source dto.CharacterDetail) domain.CharacterProfile {
	if source.Name != "" {
		profile.Character.Name = source.Name
	}
	profile.Character.Nickname = source.Nickname
	profile.Character.Rarity = source.Rarity
	profile.Character.Element = domain.Element(source.Element)
	profile.Character.ElementName = profile.Character.Element.String()
	profile.Character.WeaponType = source.Weapon
	profile.Character.WeaponTypeName = weaponTypeName(source.Weapon)
	profile.Character.IconPath = source.Icon
	profile.Character.BackgroundPath = source.Background
	profile.Description = normalizer.CleanText(source.Description)
	profile.Birthday = source.CharaInfo.Birth
	profile.Gender = source.CharaInfo.Sex
	profile.Character.Gender = source.CharaInfo.Sex
	profile.Region = source.CharaInfo.Country
	profile.Faction = source.CharaInfo.Influence
	profile.TalentName = source.CharaInfo.TalentName
	profile.TalentDescription = normalizer.CleanText(source.CharaInfo.TalentDoc)
	profile.DetailLoaded = true

	profile.Skills = make([]domain.Skill, 0, len(source.SkillTrees))
	profile.Progression = domain.CharacterProgression{
		Ascensions: make([]domain.AscensionStage, 0, len(source.Ascensions)),
		Skills:     make([]domain.SkillProgression, 0, len(source.SkillTrees)),
		LevelEXP:   append([]int(nil), source.LevelEXP...),
		Stats:      []domain.CharacterStat{},
	}
	profile.Extras = mapCharacterExtras(source)
	for nodeID, node := range source.SkillTrees {
		if node.Skill == nil || node.Skill.Name == "" {
			continue
		}
		description := node.Skill.Description
		params := node.Skill.Params
		if node.Skill.SimpleDesc != "" {
			description = node.Skill.SimpleDesc
			params = node.Skill.SimpleParams
		}
		description, _ = normalizer.ApplyParams(description, params)
		levels, _ := json.Marshal(node.Skill.Levels)
		order := node.Coordinate
		if order == 0 {
			order, _ = strconv.Atoi(nodeID)
		}
		profile.Skills = append(profile.Skills, domain.Skill{
			NodeID:      nodeID,
			Type:        node.Skill.Type,
			Name:        node.Skill.Name,
			Description: normalizer.CleanText(description),
			IconPath:    node.Skill.Icon,
			LevelsJSON:  string(levels),
			SortOrder:   order,
		})
		progression := domain.SkillProgression{
			NodeID: nodeID, NodeType: node.NodeType, Type: node.Skill.Type,
			Name: node.Skill.Name, IconPath: node.Skill.Icon,
			MaxLevel: 1, UnlockCosts: mapCosts(node.Consume),
			LevelCosts: []domain.SkillLevelCost{}, Values: skillValues(node.Skill.Levels),
		}
		for levelText, costs := range node.Skill.Consume {
			level, err := strconv.Atoi(levelText)
			if err != nil {
				continue
			}
			if level > progression.MaxLevel {
				progression.MaxLevel = level
			}
			progression.LevelCosts = append(progression.LevelCosts, domain.SkillLevelCost{Level: level, Costs: mapCosts(costs)})
		}
		sort.Slice(progression.LevelCosts, func(i, j int) bool { return progression.LevelCosts[i].Level < progression.LevelCosts[j].Level })
		profile.Progression.Skills = append(profile.Progression.Skills, progression)
	}
	sort.Slice(profile.Skills, func(i, j int) bool {
		if profile.Skills[i].SortOrder == profile.Skills[j].SortOrder {
			return profile.Skills[i].NodeID < profile.Skills[j].NodeID
		}
		return profile.Skills[i].SortOrder < profile.Skills[j].SortOrder
	})
	sort.Slice(profile.Progression.Skills, func(i, j int) bool {
		return profile.Progression.Skills[i].NodeID < profile.Progression.Skills[j].NodeID
	})

	for stageText, costs := range source.Ascensions {
		stage, err := strconv.Atoi(stageText)
		if err != nil {
			continue
		}
		profile.Progression.Ascensions = append(profile.Progression.Ascensions, domain.AscensionStage{
			Stage: stage, UnlockLevel: ascensionUnlockLevel(stage), Costs: mapCosts(costs),
		})
	}
	sort.Slice(profile.Progression.Ascensions, func(i, j int) bool {
		return profile.Progression.Ascensions[i].Stage < profile.Progression.Ascensions[j].Stage
	})
	for ascensionText, levels := range source.Stats {
		ascension, err := strconv.Atoi(ascensionText)
		if err != nil {
			continue
		}
		for levelText, stat := range levels {
			level, err := strconv.Atoi(levelText)
			if err == nil {
				profile.Progression.Stats = append(profile.Progression.Stats, domain.CharacterStat{
					Ascension: ascension, Level: level, HP: stat.Life, ATK: stat.ATK, DEF: stat.DEF,
				})
			}
		}
	}
	sort.Slice(profile.Progression.Stats, func(i, j int) bool {
		if profile.Progression.Stats[i].Ascension == profile.Progression.Stats[j].Ascension {
			return profile.Progression.Stats[i].Level < profile.Progression.Stats[j].Level
		}
		return profile.Progression.Stats[i].Ascension < profile.Progression.Stats[j].Ascension
	})

	profile.Chains = make([]domain.ResonanceChain, 0, len(source.Chains))
	for id, chain := range source.Chains {
		sequence, err := strconv.Atoi(id)
		if err != nil {
			continue
		}
		description, _ := normalizer.ApplyParams(chain.Description, chain.Params)
		profile.Chains = append(profile.Chains, domain.ResonanceChain{
			Sequence:    sequence,
			Name:        chain.Name,
			Description: normalizer.CleanText(description),
			IconPath:    chain.Icon,
		})
	}
	sort.Slice(profile.Chains, func(i, j int) bool {
		return profile.Chains[i].Sequence < profile.Chains[j].Sequence
	})

	if len(source.Recommend.Weapons) > 0 {
		profile.SignatureWeapon = &domain.Weapon{ID: source.Recommend.Weapons[0]}
	}
	return profile
}

func mapCharacterExtras(source dto.CharacterDetail) domain.CharacterExtras {
	extras := domain.CharacterExtras{
		Tags: []domain.CharacterTag{}, Stories: []domain.LoreEntry{}, Goods: []domain.LoreEntry{},
		Forte: domain.ForteGuide{IconPath: source.Forte.Icon, Descriptions: []string{}, Features: []string{}, Actions: []domain.ForteAction{}},
		Weakness: domain.WeaknessStats{
			BuildUp: source.Weakness.BuildUp, BuildUpMax: source.Weakness.BuildUpMax,
			TotalBonus: source.Weakness.TotalBonus, BreakRatio: source.Weakness.BreakRatio, Mastery: source.Weakness.Mastery,
		},
		SkillBranches: []domain.SkillBranch{}, SkillTree: []domain.SkillTreeNodeInfo{},
	}
	tagIDs := numericKeys(source.Tags)
	for _, id := range tagIDs {
		tag := source.Tags[strconv.Itoa(id)]
		extras.Tags = append(extras.Tags, domain.CharacterTag{ID: id, Name: tag.Name, Description: normalizer.CleanText(tag.Description), IconPath: tag.Icon, Color: tag.Color})
	}
	for _, story := range source.Stories {
		extras.Stories = append(extras.Stories, domain.LoreEntry{Title: story.Title, Content: normalizer.CleanText(story.Content), IconPath: story.Icon})
	}
	for _, good := range source.Goods {
		extras.Goods = append(extras.Goods, domain.LoreEntry{Title: good.Title, Content: normalizer.CleanText(good.Content), IconPath: good.Icon})
	}
	for _, description := range source.Forte.Descriptions {
		extras.Forte.Descriptions = append(extras.Forte.Descriptions, normalizer.CleanText(description))
	}
	for _, feature := range source.ForteNew.Features {
		extras.Forte.Features = append(extras.Forte.Features, normalizer.CleanText(feature))
	}
	for _, id := range stringKeys(source.Forte.Inputs) {
		action := source.Forte.Inputs[id]
		extras.Forte.Actions = append(extras.Forte.Actions, domain.ForteAction{Name: id, Description: normalizer.CleanText(action.Description), Inputs: stringSlice(action.Inputs), Images: []string{}})
	}
	for _, instructionID := range stringKeys(source.ForteNew.Instructions) {
		instruction := source.ForteNew.Instructions[instructionID]
		for _, stepID := range stringKeys(instruction.Steps) {
			step := instruction.Steps[stepID]
			extras.Forte.Actions = append(extras.Forte.Actions, domain.ForteAction{Name: instruction.Name, Description: normalizer.CleanText(step.Description), Inputs: stringSlice(step.Inputs), Images: stringSlice(step.Images)})
		}
	}
	for _, id := range numericKeys(source.Branches) {
		branch := source.Branches[strconv.Itoa(id)]
		extras.SkillBranches = append(extras.SkillBranches, domain.SkillBranch{ID: int64(id), Name: branch.Name, Description: normalizer.CleanText(branch.Description), IconPath: branch.Icon})
	}
	for _, nodeID := range stringKeys(source.SkillTrees) {
		node := source.SkillTrees[nodeID]
		extras.SkillTree = append(extras.SkillTree, domain.SkillTreeNodeInfo{
			NodeID: nodeID, NodeType: node.NodeType, Coordinate: node.Coordinate,
			ParentNodes: int64Slice(node.ParentNodes), BranchIDs: int64Slice(node.BranchIDs),
			UnlockCondition: node.UnlockCondition,
		})
	}
	return extras
}

func stringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string{}, values...)
}

func int64Slice(values []int64) []int64 {
	if len(values) == 0 {
		return []int64{}
	}
	return append([]int64{}, values...)
}

func numericKeys[T any](values map[string]T) []int {
	keys := make([]int, 0, len(values))
	for key := range values {
		if number, err := strconv.Atoi(key); err == nil {
			keys = append(keys, number)
		}
	}
	sort.Ints(keys)
	return keys
}

func stringKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left, leftErr := strconv.Atoi(keys[i])
		right, rightErr := strconv.Atoi(keys[j])
		if leftErr == nil && rightErr == nil {
			return left < right
		}
		return keys[i] < keys[j]
	})
	return keys
}

func mapCosts(costs []dto.Cost) []domain.MaterialCost {
	result := make([]domain.MaterialCost, 0, len(costs))
	for _, cost := range costs {
		if cost.Key > 0 && cost.Value > 0 {
			result = append(result, domain.MaterialCost{Material: domain.Material{ID: cost.Key}, Quantity: cost.Value})
		}
	}
	return result
}

func skillValues(levels map[string]any) []domain.SkillValueRow {
	if len(levels) == 0 {
		return []domain.SkillValueRow{}
	}
	raw, _ := json.Marshal(levels)
	var rows map[string]struct {
		Name  string  `json:"name"`
		Param [][]any `json:"param"`
	}
	if json.Unmarshal(raw, &rows) != nil {
		return []domain.SkillValueRow{}
	}
	keys := make([]int, 0, len(rows))
	for key := range rows {
		if number, err := strconv.Atoi(key); err == nil {
			keys = append(keys, number)
		}
	}
	sort.Ints(keys)
	result := make([]domain.SkillValueRow, 0, len(keys))
	for _, key := range keys {
		row := rows[strconv.Itoa(key)]
		values := []string{}
		for _, group := range row.Param {
			for _, value := range group {
				if value != nil {
					values = append(values, fmt.Sprint(value))
				}
			}
		}
		result = append(result, domain.SkillValueRow{Name: row.Name, Values: values})
	}
	return result
}

func ascensionUnlockLevel(stage int) int {
	levels := []int{0, 20, 40, 50, 60, 70, 80}
	if stage >= 1 && stage < len(levels) {
		return levels[stage]
	}
	return 0
}

func Weapon(source dto.Weapon) domain.Weapon {
	effect, _ := normalizer.ApplyParams(source.Effect, source.Params)
	params, _ := json.Marshal(source.Params)
	return domain.Weapon{
		ID:          source.ID,
		Name:        source.Name,
		Rarity:      source.Rarity,
		Type:        source.Type,
		TypeName:    weaponTypeName(source.Type),
		Description: normalizer.CleanText(source.Description),
		EffectName:  source.EffectName,
		Effect:      normalizer.CleanText(effect),
		IconPath:    source.Icon,
		ParamsJSON:  string(params),
	}
}

func WeaponIndex(id string, source dto.WeaponIndexEntry, version string) (domain.Weapon, bool) {
	numericID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return domain.Weapon{}, false
	}
	return domain.Weapon{
		ID:          numericID,
		Name:        source.Name,
		Rarity:      source.Rank,
		Type:        source.Type,
		TypeName:    weaponTypeName(source.Type),
		Description: normalizer.CleanText(source.Description),
		IconPath:    source.Icon,
		BaseATK:     source.BaseATK,
		SubStat:     source.SubStat,
		GameVersion: version,
		Level:       1,
		Rank:        1,
	}, true
}

func MergeWeaponDetail(base domain.Weapon, source dto.Weapon) domain.Weapon {
	detail := Weapon(source)
	detail.BaseATK = base.BaseATK
	detail.SubStat = base.SubStat
	detail.GameVersion = base.GameVersion
	detail.Level = base.Level
	detail.Rank = base.Rank
	return detail
}

func weaponTypeName(code int) string {
	switch code {
	case 1:
		return "Broadblade"
	case 2:
		return "Sword"
	case 3:
		return "Pistols"
	case 4:
		return "Gauntlets"
	case 5:
		return "Rectifier"
	default:
		return "Unknown"
	}
}
