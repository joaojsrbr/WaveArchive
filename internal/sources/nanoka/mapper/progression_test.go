package mapper

import (
	"testing"

	"wavearchive/internal/domain"
	"wavearchive/internal/sources/nanoka/dto"
)

func TestMergeCharacterDetailMapsAscensionSkillCostsAndValues(t *testing.T) {
	source := dto.CharacterDetail{ID: 1606, Name: "Roccia", Rarity: 5}
	source.Ascensions = map[string][]dto.Cost{"1": {{Key: 41100051, Value: 4}, {Key: 2, Value: 5000}}}
	source.LevelEXP = []int{0, 400}
	source.Stats = map[string]map[string]dto.CharacterStat{"0": {"1": {Life: 980, ATK: 30, DEF: 98}}}
	source.Tags = map[string]dto.CharacterTag{"8": {Name: "Traction", Description: "Pulls targets", Color: "77adff"}}
	source.Stories = []dto.LoreEntry{{Title: "Golden Harpoon", Content: "A story."}}
	source.Weakness = dto.WeaknessStats{BuildUpMax: 10000, BreakRatio: 10000}
	source.Branches = map[string]dto.SkillBranch{"121001": {Name: "Tune Rupture", Description: "Single-target mode."}}
	source.ForteNew.Features = []string{"Deal substantial damage."}
	source.Forte.Inputs = map[string]dto.ForteInput{
		"160601": {Description: "Use the Forte action."},
	}
	source.SkillTrees = map[string]dto.SkillTreeNode{
		"1": {
			NodeType: 2,
			Consume:  []dto.Cost{{Key: 43020032, Value: 2}},
			Skill: &dto.Skill{
				Name: "Pero, Easy", Type: "Normal Attack",
				Consume: map[string][]dto.Cost{"1": {}, "2": {{Key: 43020051, Value: 2}, {Key: 2, Value: 1500}}},
				Levels:  map[string]any{"1": map[string]any{"name": "Stage 1 DMG", "param": [][]any{{"36.81%", "39.83%"}}}},
			},
		},
	}

	profile := MergeCharacterDetail(domain.CharacterProfile{Character: domain.Character{ID: 1606}}, source)
	if got := profile.Progression.Ascensions[0]; got.UnlockLevel != 20 || got.Costs[1].Quantity != 5000 {
		t.Fatalf("unexpected ascension: %#v", got)
	}
	skill := profile.Progression.Skills[0]
	if skill.MaxLevel != 2 || skill.LevelCosts[1].Costs[0].Quantity != 2 {
		t.Fatalf("unexpected skill costs: %#v", skill)
	}
	if len(skill.Values) != 1 || skill.Values[0].Values[1] != "39.83%" {
		t.Fatalf("unexpected values: %#v", skill.Values)
	}
	if len(profile.Progression.Stats) != 1 || profile.Progression.Stats[0].HP != 980 {
		t.Fatalf("unexpected stats: %#v", profile.Progression.Stats)
	}
	if profile.Extras.Tags[0].Name != "Traction" || profile.Extras.Stories[0].Title != "Golden Harpoon" {
		t.Fatalf("unexpected extras: %#v", profile.Extras)
	}
	if profile.Extras.SkillBranches[0].ID != 121001 || profile.Extras.Weakness.BreakRatio != 10000 {
		t.Fatalf("unexpected branches/weakness: %#v", profile.Extras)
	}
	if profile.Extras.Forte.Actions[0].Inputs == nil || profile.Extras.Forte.Actions[0].Images == nil {
		t.Fatalf("forte collections must be serialized as arrays: %#v", profile.Extras.Forte.Actions[0])
	}
	if profile.Extras.SkillTree[0].ParentNodes == nil || profile.Extras.SkillTree[0].BranchIDs == nil {
		t.Fatalf("skill tree collections must be serialized as arrays: %#v", profile.Extras.SkillTree[0])
	}
}
