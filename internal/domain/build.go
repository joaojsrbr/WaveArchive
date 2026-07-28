package domain

import "context"

type Build struct {
	ID             int64       `json:"id"`
	Name           string      `json:"name"`
	CharacterID    int64       `json:"characterId"`
	CharacterName  string      `json:"characterName"`
	CharacterIcon  string      `json:"characterIcon"`
	CharacterLevel int         `json:"characterLevel"`
	Sequence       int         `json:"sequence"`
	WeaponID       *int64      `json:"weaponId,omitempty"`
	WeaponName     string      `json:"weaponName"`
	WeaponIcon     string      `json:"weaponIcon"`
	WeaponLevel    int         `json:"weaponLevel"`
	WeaponRank     int         `json:"weaponRank"`
	Echoes         []OwnedEcho `json:"echoes"`
	TargetEnemyID  *int64      `json:"targetEnemyId,omitempty"`
	RotationID     *int64      `json:"rotationId,omitempty"`
	Conditions     string      `json:"conditions"`
	Notes          string      `json:"notes"`
	Favorite       bool        `json:"favorite"`
	Locked         bool        `json:"locked"`
	GameVersion    string      `json:"gameVersion"`
	CreatedAt      string      `json:"createdAt"`
	UpdatedAt      string      `json:"updatedAt"`
}

type BuildRepository interface {
	List(ctx context.Context) ([]Build, error)
	Get(ctx context.Context, id int64) (Build, error)
	Save(ctx context.Context, build Build) (Build, error)
	Duplicate(ctx context.Context, id int64) (Build, error)
	SoftDelete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) error
	History(ctx context.Context, id int64) ([]BuildVersion, error)
}
