package repository_test

import (
	"context"
	"path/filepath"
	"testing"

	"wavearchive/internal/database"
	"wavearchive/internal/domain"
	"wavearchive/internal/repository"
)

func TestWeaponCatalogSearchAndAccountState(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "weapons.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := repository.NewWeaponSQLite(db.SQL())
	weapons := []domain.Weapon{
		{ID: 201, Name: "Emerald of Genesis", Rarity: 5, Type: 2, Description: "A sword.", EffectName: "Stormy Resolution", BaseATK: 587, SubStat: "Crit. Rate"},
		{ID: 202, Name: "Training Sword", Rarity: 1, Type: 2, Description: "For training.", BaseATK: 250, SubStat: "ATK"},
	}
	if err := repo.ReplaceSynced(context.Background(), "3.6.1", weapons); err != nil {
		t.Fatal(err)
	}
	got, err := repo.List(context.Background(), domain.WeaponFilter{Query: "Emerald", Rarity: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].BaseATK != 587 {
		t.Fatalf("unexpected weapons: %#v", got)
	}
	if err := repo.UpdateAccount(context.Background(), domain.WeaponAccountUpdate{
		WeaponID: 201, Owned: true, Level: 90, Rank: 2, Favorite: true,
	}); err != nil {
		t.Fatal(err)
	}
	weapon, err := repo.Get(context.Background(), 201)
	if err != nil {
		t.Fatal(err)
	}
	if !weapon.Owned || !weapon.Favorite || weapon.Level != 90 || weapon.Rank != 2 {
		t.Fatalf("unexpected account state: %#v", weapon)
	}
}
