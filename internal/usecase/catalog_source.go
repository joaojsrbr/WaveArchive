package usecase

import (
	"context"

	"wavearchive/internal/sources/nanoka/dto"
)

// CatalogSource is the normalized boundary consumed by the catalog use cases.
// Providers keep their raw schemas and conversion rules outside the domain.
type CatalogSource interface {
	DetectVersion(ctx context.Context) (string, error)
	CharacterIndex(ctx context.Context, version string) (map[string]dto.CharacterIndexEntry, error)
	CharacterDetail(ctx context.Context, version, language string, id int64) (dto.CharacterDetail, error)
	ItemCatalog(ctx context.Context, version, language string) (map[string]dto.Item, error)
	ItemIndex(ctx context.Context, version, language string) (map[string]dto.ItemIndexEntry, error)
	WeaponIndex(ctx context.Context, version string) (map[string]dto.WeaponIndexEntry, error)
	Weapon(ctx context.Context, version, language string, id int64) (dto.Weapon, error)
	EchoIndex(ctx context.Context, version string) (map[string]dto.EchoIndexEntry, error)
	EchoDetail(ctx context.Context, version, language string, id int64) (dto.EchoDetail, error)
	SonataIndex(ctx context.Context, version string) (map[string]dto.SonataIndexEntry, error)
}
