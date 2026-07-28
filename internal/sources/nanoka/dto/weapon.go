package dto

type WeaponIndexEntry struct {
	Icon        string `json:"icon"`
	Rank        int    `json:"rank"`
	Type        int    `json:"type"`
	Name        string `json:"en"`
	Description string `json:"desc"`
	BaseATK     int    `json:"atk"`
	SubStat     string `json:"sub"`
}

type Weapon struct {
	ID          int64  `json:"id"`
	Rarity      int    `json:"rarity"`
	Type        int    `json:"type"`
	Name        string `json:"name"`
	Description string `json:"desc"`
	Icon        string `json:"icon"`
	Effect      string `json:"effect"`
	EffectName  string `json:"effect_name"`
	Params      []any  `json:"param"`
}
