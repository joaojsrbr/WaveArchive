package domain

import "context"

type CharacterRepository interface {
	List(ctx context.Context, filter CharacterFilter) ([]Character, error)
	GetProfile(ctx context.Context, id int64) (CharacterProfile, error)
	ReplaceSynced(ctx context.Context, version string, characters []CharacterProfile) error
	UpdateAccount(ctx context.Context, update CharacterAccountUpdate) error
	Status(ctx context.Context) (CatalogStatus, error)
}

type WeaponRepository interface {
	List(ctx context.Context, filter WeaponFilter) ([]Weapon, error)
	Get(ctx context.Context, id int64) (Weapon, error)
	ReplaceSynced(ctx context.Context, version string, weapons []Weapon) error
	UpdateAccount(ctx context.Context, update WeaponAccountUpdate) error
}
