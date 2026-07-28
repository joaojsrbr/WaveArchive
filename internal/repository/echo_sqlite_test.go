package repository_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"wavearchive/internal/database"
	"wavearchive/internal/domain"
	"wavearchive/internal/repository"
)

func TestEchoCatalogInventoryAndFilters(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(filepath.Join(t.TempDir(), "echoes.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := repository.NewEchoSQLite(db.SQL())
	if err := repo.ReplaceSynced(ctx, "3.6.1", []domain.Echo{
		{ID: 101, Name: "Vanguard Junrock", Cost: 1, Skill: "Transform into a Junrock.", SonataIDsJSON: "[3,7]", RaritiesJSON: "[2,3,4,5]"},
		{ID: 102, Name: "Dreamless", Cost: 4, Skill: "Deal Havoc damage.", SonataIDsJSON: "[7]", RaritiesJSON: "[4,5]"},
	}, []domain.Sonata{
		{ID: 3, Name: "Moonlit Clouds", TwoPiece: "Energy Regen"},
		{ID: 7, Name: "Sun-sinking Eclipse", TwoPiece: "Havoc DMG"},
	}); err != nil {
		t.Fatal(err)
	}
	found, err := repo.List(ctx, domain.EchoFilter{Query: "Dreamless", Cost: 4, SonataID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].ID != 102 {
		t.Fatalf("unexpected filtered echoes: %#v", found)
	}
	sonataID := int64(7)
	owned, err := repo.SaveOwned(ctx, domain.OwnedEcho{
		EchoID: 102, MainStat: "CRIT Rate 22%", SubstatsJSON: `["CRIT DMG 15%"]`,
		Level: 25, SonataID: &sonataID, Locked: true, Favorite: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	items, err := repo.ListOwned(ctx)
	if err != nil || len(items) != 1 || items[0].EchoName != "Dreamless" || items[0].SonataName != "Sun-sinking Eclipse" {
		t.Fatalf("unexpected owned echoes: %#v, %v", items, err)
	}
	if err := repo.DeleteOwned(ctx, owned.ID); err != sql.ErrNoRows {
		t.Fatalf("locked echo deletion error = %v, want sql.ErrNoRows", err)
	}
	owned.Locked = false
	if _, err := repo.SaveOwned(ctx, owned); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteOwned(ctx, owned.ID); err != nil {
		t.Fatal(err)
	}
}
