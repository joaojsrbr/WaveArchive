package dto

type Manifest struct {
	WW struct {
		Latest string `json:"latest"`
	} `json:"ww"`
}

type CharacterIndexEntry struct {
	Name       string `json:"en"`
	Nickname   string `json:"nickname"`
	Rank       int    `json:"rank"`
	Element    int    `json:"element"`
	Weapon     int    `json:"weapon"`
	Icon       string `json:"icon"`
	Background string `json:"background"`
	APIOrder   int    `json:"-"`
}

type CharacterDetail struct {
	ID          int64  `json:"id"`
	Rarity      int    `json:"rarity"`
	Weapon      int    `json:"weapon"`
	Element     int    `json:"element"`
	Name        string `json:"name"`
	Nickname    string `json:"nick_name"`
	Description string `json:"desc"`
	Icon        string `json:"icon"`
	Background  string `json:"background"`
	CharaInfo   struct {
		Birth      string `json:"birth"`
		Sex        string `json:"sex"`
		Country    string `json:"country"`
		Influence  string `json:"influence"`
		TalentName string `json:"talent_name"`
		TalentDoc  string `json:"talent_doc"`
	} `json:"chara_info"`
	SkillTrees map[string]SkillTreeNode            `json:"skill_trees"`
	Tags       map[string]CharacterTag             `json:"tag"`
	Stories    []LoreEntry                         `json:"stories"`
	Goods      []LoreEntry                         `json:"goods"`
	Weakness   WeaknessStats                       `json:"stats_weakness"`
	Branches   map[string]SkillBranch              `json:"skill_branches"`
	Forte      ForteGuide                          `json:"forte"`
	ForteNew   ForteNew                            `json:"forte_new"`
	Chains     map[string]Chain                    `json:"chains"`
	Ascensions map[string][]Cost                   `json:"ascensions"`
	LevelEXP   []int                               `json:"level_exp"`
	Stats      map[string]map[string]CharacterStat `json:"stats"`
	Recommend  struct {
		Weapons []int64 `json:"weapon"`
	} `json:"recommend"`
}

type SkillTreeNode struct {
	Coordinate      int     `json:"coordinate"`
	NodeType        int     `json:"node_type"`
	Consume         []Cost  `json:"consume"`
	ParentNodes     []int64 `json:"parent_nodes"`
	BranchIDs       []int64 `json:"skill_branch_ids"`
	UnlockCondition int     `json:"un_lock_condition"`
	Skill           *Skill  `json:"skill"`
}

type Skill struct {
	Name         string            `json:"name"`
	Description  string            `json:"desc"`
	Params       []any             `json:"param"`
	SimpleDesc   string            `json:"simple_desc"`
	SimpleParams []any             `json:"simple_param"`
	Icon         string            `json:"icon"`
	Type         string            `json:"type"`
	Levels       map[string]any    `json:"level"`
	Consume      map[string][]Cost `json:"consume"`
}

type Cost struct {
	Key   int64 `json:"key"`
	Value int   `json:"value"`
}

type CharacterStat struct {
	Life float64 `json:"life"`
	ATK  float64 `json:"atk"`
	DEF  float64 `json:"def"`
}

type Item struct {
	ID          int64    `json:"id"`
	Rarity      int      `json:"rarity"`
	Type        int      `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"desc"`
	Icon        string   `json:"icon"`
	Source      []string `json:"source"`
}

type ItemIndexEntry struct {
	Rank int    `json:"rank"`
	Type int    `json:"type"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type CharacterTag struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

type LoreEntry struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Icon    string `json:"icon"`
}

type WeaknessStats struct {
	BuildUp    float64 `json:"weakness_build_up"`
	BuildUpMax float64 `json:"weakness_build_up_max"`
	TotalBonus float64 `json:"weakness_total_bonus"`
	BreakRatio float64 `json:"break_weakness_ratio"`
	Mastery    float64 `json:"weakness_mastery"`
}

type SkillBranch struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
	Icon        string `json:"icon"`
}

type ForteInput struct {
	Description string   `json:"desc"`
	Inputs      []string `json:"input_list"`
}

type ForteGuide struct {
	Inputs       map[string]ForteInput `json:"skill_input_list"`
	Icon         string                `json:"icon"`
	Descriptions []string              `json:"desc_list"`
}

type ForteNewStep struct {
	Description string   `json:"desc"`
	Inputs      []string `json:"input_list"`
	Images      []string `json:"image_list"`
}

type ForteNewInstruction struct {
	Name  string                  `json:"name"`
	Steps map[string]ForteNewStep `json:"desc"`
}

type ForteNew struct {
	Features     []string                       `json:"features"`
	Instructions map[string]ForteNewInstruction `json:"instructions"`
}

type Chain struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
	Params      []any  `json:"param"`
	Icon        string `json:"icon"`
}
