package domain

import "context"

type Echo struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	Type          string `json:"type"`
	Class         string `json:"class"`
	Cost          int    `json:"cost"`
	Place         string `json:"place"`
	IconPath      string `json:"iconPath"`
	Skill         string `json:"skill"`
	RaritiesJSON  string `json:"raritiesJson"`
	SonataIDsJSON string `json:"sonataIdsJson"`
	GameVersion   string `json:"gameVersion"`
	OwnedCount    int    `json:"ownedCount"`
	Favorite      bool   `json:"favorite"`
}

type Sonata struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	IconPath    string `json:"iconPath"`
	TwoPiece    string `json:"twoPiece"`
	FivePiece   string `json:"fivePiece"`
	GameVersion string `json:"gameVersion"`
}

type EchoFilter struct {
	Query     string `json:"query"`
	Cost      int    `json:"cost"`
	SonataID  int64  `json:"sonataId"`
	Class     string `json:"class"`
	Type      string `json:"type"`
	Place     string `json:"place"`
	Rarity    int    `json:"rarity"`
	MinOwned  int    `json:"minOwned"`
	OwnedOnly bool   `json:"ownedOnly"`
	Favorites bool   `json:"favorites"`
	Sort      string `json:"sort"`
}

type OwnedEcho struct {
	ID            int64  `json:"id"`
	EchoID        int64  `json:"echoId"`
	EchoName      string `json:"echoName"`
	IconPath      string `json:"iconPath"`
	Cost          int    `json:"cost"`
	MainStat      string `json:"mainStat"`
	SubstatsJSON  string `json:"substatsJson"`
	Level         int    `json:"level"`
	SonataID      *int64 `json:"sonataId,omitempty"`
	SonataName    string `json:"sonataName"`
	CharacterID   *int64 `json:"characterId,omitempty"`
	CharacterName string `json:"characterName"`
	Locked        bool   `json:"locked"`
	Favorite      bool   `json:"favorite"`
	Note          string `json:"note"`
}

type EchoRepository interface {
	List(ctx context.Context, filter EchoFilter) ([]Echo, error)
	Get(ctx context.Context, id int64) (Echo, error)
	ListSonatas(ctx context.Context) ([]Sonata, error)
	ReplaceSynced(ctx context.Context, version string, echoes []Echo, sonatas []Sonata) error
	ListOwned(ctx context.Context) ([]OwnedEcho, error)
	SaveOwned(ctx context.Context, echo OwnedEcho) (OwnedEcho, error)
	DeleteOwned(ctx context.Context, id int64) error
}
