package domain

import "context"

type Team struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Members     []TeamMember `json:"members"`
	Notes       string       `json:"notes"`
	Favorite    bool         `json:"favorite"`
	Locked      bool         `json:"locked"`
	GameVersion string       `json:"gameVersion"`
	CreatedAt   string       `json:"createdAt"`
	UpdatedAt   string       `json:"updatedAt"`
}

type TeamMember struct {
	Slot          int    `json:"slot"`
	CharacterID   int64  `json:"characterId"`
	CharacterName string `json:"characterName"`
	CharacterIcon string `json:"characterIcon"`
	BuildID       *int64 `json:"buildId,omitempty"`
	BuildName     string `json:"buildName"`
	Role          string `json:"role"`
	CustomRole    string `json:"customRole"`
}

type TeamRepository interface {
	List(ctx context.Context) ([]Team, error)
	Get(ctx context.Context, id int64) (Team, error)
	Save(ctx context.Context, team Team) (Team, error)
	Duplicate(ctx context.Context, id int64) (Team, error)
	SoftDelete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) error
}
