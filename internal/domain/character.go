package domain

import "time"

type Element int

const (
	ElementGlacio  Element = 1
	ElementFusion  Element = 2
	ElementElectro Element = 3
	ElementAero    Element = 4
	ElementSpectro Element = 5
	ElementHavoc   Element = 6
)

func (e Element) String() string {
	switch e {
	case ElementGlacio:
		return "Glacio"
	case ElementFusion:
		return "Fusion"
	case ElementElectro:
		return "Electro"
	case ElementAero:
		return "Aero"
	case ElementSpectro:
		return "Spectro"
	case ElementHavoc:
		return "Havoc"
	default:
		return "Unknown"
	}
}

type Character struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Nickname       string  `json:"nickname"`
	Rarity         int     `json:"rarity"`
	Element        Element `json:"elementCode"`
	ElementName    string  `json:"element"`
	WeaponType     int     `json:"weaponTypeCode"`
	WeaponTypeName string  `json:"weaponType"`
	IconPath       string  `json:"iconPath"`
	BackgroundPath string  `json:"backgroundPath"`
	Owned          bool    `json:"owned"`
	Level          int     `json:"level"`
	Sequence       int     `json:"sequence"`
	Favorite       bool    `json:"favorite"`
	GameVersion    string  `json:"gameVersion"`
	APIOrder       int     `json:"apiOrder"`
}

type CharacterAccountUpdate struct {
	CharacterID int64 `json:"characterId"`
	Owned       bool  `json:"owned"`
	Level       int   `json:"level"`
	Sequence    int   `json:"sequence"`
	Favorite    bool  `json:"favorite"`
}

type CharacterProfile struct {
	Character         Character            `json:"character"`
	Description       string               `json:"description"`
	Birthday          string               `json:"birthday"`
	Gender            string               `json:"gender"`
	Region            string               `json:"region"`
	Faction           string               `json:"faction"`
	TalentName        string               `json:"talentName"`
	TalentDescription string               `json:"talentDescription"`
	SignatureWeapon   *Weapon              `json:"signatureWeapon,omitempty"`
	Skills            []Skill              `json:"skills"`
	Chains            []ResonanceChain     `json:"chains"`
	Progression       CharacterProgression `json:"progression"`
	Extras            CharacterExtras      `json:"extras"`
	DetailLoaded      bool                 `json:"-"`
}

type Skill struct {
	NodeID      string `json:"nodeId"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconPath    string `json:"iconPath"`
	LevelsJSON  string `json:"levelsJson"`
	SortOrder   int    `json:"sortOrder"`
}

type ResonanceChain struct {
	Sequence    int    `json:"sequence"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconPath    string `json:"iconPath"`
}

type Material struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Rarity      int      `json:"rarity"`
	Type        int      `json:"type"`
	Description string   `json:"description"`
	IconPath    string   `json:"iconPath"`
	Sources     []string `json:"sources"`
	GameVersion string   `json:"gameVersion"`
}

type MaterialCost struct {
	Material Material `json:"material"`
	Quantity int      `json:"quantity"`
}

type AscensionStage struct {
	Stage       int            `json:"stage"`
	UnlockLevel int            `json:"unlockLevel"`
	Costs       []MaterialCost `json:"costs"`
}

type SkillLevelCost struct {
	Level int            `json:"level"`
	Costs []MaterialCost `json:"costs"`
}

type SkillValueRow struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type SkillProgression struct {
	NodeID      string           `json:"nodeId"`
	NodeType    int              `json:"nodeType"`
	Type        string           `json:"type"`
	Name        string           `json:"name"`
	IconPath    string           `json:"iconPath"`
	MaxLevel    int              `json:"maxLevel"`
	UnlockCosts []MaterialCost   `json:"unlockCosts"`
	LevelCosts  []SkillLevelCost `json:"levelCosts"`
	Values      []SkillValueRow  `json:"values"`
}

type CharacterStat struct {
	Ascension int     `json:"ascension"`
	Level     int     `json:"level"`
	HP        float64 `json:"hp"`
	ATK       float64 `json:"atk"`
	DEF       float64 `json:"def"`
}

type CharacterProgression struct {
	Ascensions []AscensionStage   `json:"ascensions"`
	Skills     []SkillProgression `json:"skills"`
	LevelEXP   []int              `json:"levelExp"`
	Stats      []CharacterStat    `json:"stats"`
}

type ProgressionPlanRequest struct {
	CharacterID    int64          `json:"characterId"`
	CurrentLevel   int            `json:"currentLevel"`
	TargetLevel    int            `json:"targetLevel"`
	CurrentSkills  map[string]int `json:"currentSkills"`
	TargetSkills   map[string]int `json:"targetSkills"`
	IncludeUnlocks bool           `json:"includeUnlocks"`
}

type ProgressionPlan struct {
	CharacterID int64          `json:"characterId"`
	Ascensions  []MaterialCost `json:"ascensions"`
	Skills      []MaterialCost `json:"skills"`
	Total       []MaterialCost `json:"total"`
}

type CharacterTag struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconPath    string `json:"iconPath"`
	Color       string `json:"color"`
}

type LoreEntry struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	IconPath string `json:"iconPath"`
}

type ForteAction struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Inputs      []string `json:"inputs"`
	Images      []string `json:"images"`
}

type ForteGuide struct {
	IconPath     string        `json:"iconPath"`
	Descriptions []string      `json:"descriptions"`
	Features     []string      `json:"features"`
	Actions      []ForteAction `json:"actions"`
}

type WeaknessStats struct {
	BuildUp    float64 `json:"buildUp"`
	BuildUpMax float64 `json:"buildUpMax"`
	TotalBonus float64 `json:"totalBonus"`
	BreakRatio float64 `json:"breakRatio"`
	Mastery    float64 `json:"mastery"`
}

type SkillBranch struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconPath    string `json:"iconPath"`
}

type SkillTreeNodeInfo struct {
	NodeID          string  `json:"nodeId"`
	NodeType        int     `json:"nodeType"`
	Coordinate      int     `json:"coordinate"`
	ParentNodes     []int64 `json:"parentNodes"`
	BranchIDs       []int64 `json:"branchIds"`
	UnlockCondition int     `json:"unlockCondition"`
}

type CharacterExtras struct {
	Tags          []CharacterTag      `json:"tags"`
	Stories       []LoreEntry         `json:"stories"`
	Goods         []LoreEntry         `json:"goods"`
	Forte         ForteGuide          `json:"forte"`
	Weakness      WeaknessStats       `json:"weakness"`
	SkillBranches []SkillBranch       `json:"skillBranches"`
	SkillTree     []SkillTreeNodeInfo `json:"skillTree"`
}

type Weapon struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Rarity      int    `json:"rarity"`
	Type        int    `json:"typeCode"`
	TypeName    string `json:"type"`
	Description string `json:"description"`
	EffectName  string `json:"effectName"`
	Effect      string `json:"effect"`
	IconPath    string `json:"iconPath"`
	ParamsJSON  string `json:"paramsJson"`
	BaseATK     int    `json:"baseAtk"`
	SubStat     string `json:"subStat"`
	GameVersion string `json:"gameVersion"`
	Owned       bool   `json:"owned"`
	Level       int    `json:"level"`
	Rank        int    `json:"rank"`
	Favorite    bool   `json:"favorite"`
}

type WeaponFilter struct {
	Query     string `json:"query"`
	Type      int    `json:"type"`
	Rarity    int    `json:"rarity"`
	OwnedOnly bool   `json:"ownedOnly"`
	Favorites bool   `json:"favorites"`
	Sort      string `json:"sort"`
}

type WeaponAccountUpdate struct {
	WeaponID int64 `json:"weaponId"`
	Owned    bool  `json:"owned"`
	Level    int   `json:"level"`
	Rank     int   `json:"rank"`
	Favorite bool  `json:"favorite"`
}

type CharacterFilter struct {
	Query     string `json:"query"`
	Element   int    `json:"element"`
	Rarity    int    `json:"rarity"`
	OwnedOnly bool   `json:"ownedOnly"`
	Favorites bool   `json:"favorites"`
	Sort      string `json:"sort"`
}

type CatalogStatus struct {
	Count      int        `json:"count"`
	Version    string     `json:"version"`
	LastSyncAt *time.Time `json:"lastSyncAt"`
}

type SyncResult struct {
	Version string    `json:"version"`
	Count   int       `json:"count"`
	Synced  time.Time `json:"syncedAt"`
}
