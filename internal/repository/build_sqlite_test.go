package repository_test

import (
	"context"
	"path/filepath"
	"testing"

	"wavearchive/internal/database"
	"wavearchive/internal/domain"
	"wavearchive/internal/repository"
)

func TestBuildLifecycleWithUndo(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(filepath.Join(t.TempDir(), "builds.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	characters := repository.NewCharacterSQLite(db.SQL())
	weapons := repository.NewWeaponSQLite(db.SQL())
	if err := characters.ReplaceSynced(ctx, "3.6.1", []domain.CharacterProfile{{
		Character: domain.Character{ID: 1102, Name: "Sanhua", Rarity: 4, Element: domain.ElementGlacio, WeaponType: 2},
	}}); err != nil {
		t.Fatal(err)
	}
	if err := weapons.ReplaceSynced(ctx, "3.6.1", []domain.Weapon{{ID: 21020015, Name: "Emerald of Genesis", Rarity: 5, Type: 2}}); err != nil {
		t.Fatal(err)
	}
	echoes := repository.NewEchoSQLite(db.SQL())
	if err := echoes.ReplaceSynced(ctx, "3.6.1", []domain.Echo{{ID: 390070051, Name: "Vanguard Junrock", Cost: 1}}, nil); err != nil {
		t.Fatal(err)
	}
	ownedEcho, err := echoes.SaveOwned(ctx, domain.OwnedEcho{EchoID: 390070051, Level: 25, SubstatsJSON: "[]"})
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewBuildSQLite(db.SQL())
	weaponID := int64(21020015)
	build, err := repo.Save(ctx, domain.Build{
		Name: "Sanhua DPS", CharacterID: 1102, CharacterLevel: 90, Sequence: 6,
		WeaponID: &weaponID, WeaponLevel: 90, WeaponRank: 1, Echoes: []domain.OwnedEcho{ownedEcho}, GameVersion: "3.6.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	copy, err := repo.Duplicate(ctx, build.ID)
	if err != nil {
		t.Fatal(err)
	}
	if copy.ID == build.ID || copy.Name != "Sanhua DPS — cópia" || len(copy.Echoes) != 1 {
		t.Fatalf("unexpected duplicate: %#v", copy)
	}
	if err := repo.SoftDelete(ctx, build.ID); err != nil {
		t.Fatal(err)
	}
	list, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != copy.ID {
		t.Fatalf("soft-deleted build should be hidden: %#v", list)
	}
	if err := repo.Restore(ctx, build.ID); err != nil {
		t.Fatal(err)
	}
	list, err = repo.List(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("restore failed: %#v, %v", list, err)
	}
}
