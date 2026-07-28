package arikatsu

type textEntry struct {
	ID      string `json:"Id"`
	Content string `json:"Content"`
}

type roleInfo struct {
	ID                int64  `json:"Id"`
	QualityID         int    `json:"QualityId"`
	RoleType          int    `json:"RoleType"`
	Name              string `json:"Name"`
	Nickname          string `json:"NickName"`
	Introduction      string `json:"Introduction"`
	ElementID         int    `json:"ElementId"`
	WeaponType        int    `json:"WeaponType"`
	RoleBody          string `json:"RoleBody"`
	RoleHeadIconLarge string `json:"RoleHeadIconLarge"`
	FormationRoleCard string `json:"FormationRoleCard"`
	ShowInBag         bool   `json:"ShowInBag"`
}

type propertyValue struct {
	ID      int64   `json:"Id"`
	Value   float64 `json:"Value"`
	IsRatio bool    `json:"IsRatio"`
}

type stringArray struct {
	Values []string `json:"ArrayString"`
}

type weaponConf struct {
	ItemID                int64         `json:"ItemId"`
	IsShow                bool          `json:"IsShow"`
	WeaponName            string        `json:"WeaponName"`
	QualityID             int           `json:"QualityId"`
	WeaponType            int           `json:"WeaponType"`
	FirstProp             propertyValue `json:"FirstPropId"`
	FirstCurve            int           `json:"FirstCurve"`
	SecondProp            propertyValue `json:"SecondPropId"`
	SecondCurve           int           `json:"SecondCurve"`
	ResonID               int64         `json:"ResonId"`
	Desc                  string        `json:"Desc"`
	DescParams            []stringArray `json:"DescParams"`
	AttributesDescription string        `json:"AttributesDescription"`
	Icon                  string        `json:"Icon"`
	ShowInBag             bool          `json:"ShowInBag"`
}

type weaponPropertyGrowth struct {
	CurveID     int     `json:"CurveId"`
	Level       int     `json:"Level"`
	BreachLevel int     `json:"BreachLevel"`
	CurveValue  float64 `json:"CurveValue"`
}

type propertyIndex struct {
	ID               int64  `json:"Id"`
	Key              string `json:"Key"`
	Name             string `json:"Name"`
	AnotherName      string `json:"AnotherName"`
	ConvertToWhiteID int64  `json:"ConvertToWhiteId"`
}

type weaponReson struct {
	ResonID int64  `json:"ResonId"`
	Level   int    `json:"Level"`
	Name    string `json:"Name"`
}

type phantomItem struct {
	ItemID      int64   `json:"ItemId"`
	MonsterID   int64   `json:"MonsterId"`
	MonsterName string  `json:"MonsterName"`
	SkillID     int64   `json:"SkillId"`
	Rarity      int     `json:"Rarity"`
	QualityID   int     `json:"QualityId"`
	Icon        string  `json:"Icon"`
	FetterGroup []int64 `json:"FetterGroup"`
	PhantomType int     `json:"PhantomType"`
}

type phantomSkill struct {
	ID                int64         `json:"Id"`
	PhantomSkillID    int64         `json:"PhantomSkillId"`
	Description       string        `json:"DescriptionEx"`
	SimpleDescription string        `json:"SimplyDescription"`
	LevelDesc         []stringArray `json:"LevelDescStrArray"`
}

type keyValue struct {
	Key   int   `json:"Key"`
	Value int64 `json:"Value"`
}

type phantomFetterGroup struct {
	ID        int64      `json:"Id"`
	Name      string     `json:"FetterGroupName"`
	Icon      string     `json:"FetterElementPath"`
	FetterMap []keyValue `json:"FetterMap"`
}

type phantomFetter struct {
	ID          int64    `json:"Id"`
	Description string   `json:"EffectDescription"`
	Params      []string `json:"EffectDescriptionParam"`
}
