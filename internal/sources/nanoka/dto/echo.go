package dto

type EchoIndexEntry struct {
	Icon      string  `json:"icon"`
	Code      string  `json:"code"`
	Ranks     []int   `json:"rank"`
	Groups    []int64 `json:"group"`
	Name      string  `json:"en"`
	Intensity int     `json:"intensity"`
}

type EchoDetail struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Intensity string `json:"intensity"`
	Place     string `json:"place"`
	Icon      string `json:"icon"`
	Skill     struct {
		Description string `json:"desc"`
		SimpleDesc  string `json:"simple_desc"`
		Params      []any  `json:"param"`
	} `json:"skill"`
	Group         map[string]any `json:"group"`
	Rarity        []int          `json:"rarity"`
	IntensityCode int            `json:"intensity_code"`
}

type SonataIndexEntry struct {
	ID   int64  `json:"id"`
	Icon string `json:"icon"`
	Name struct {
		English string `json:"en"`
	} `json:"name"`
	Set map[string]map[string]struct {
		Description string `json:"desc"`
		Params      []any  `json:"param"`
	} `json:"set"`
}
