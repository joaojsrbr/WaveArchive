package repository_test

import (
	"context"
	"path/filepath"
	"testing"

	"wavearchive/internal/database"
	"wavearchive/internal/domain"
	"wavearchive/internal/repository"
)

func TestConveneHistoryUsesCatalogIDForResourceTypeAndIcon(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "convene.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.SQL().Exec(`
		INSERT INTO characters(id,name,rarity,icon_path,game_version)
		VALUES(1610,'Yangyang',5,'/cache/characters/1610.webp','3.6.1')`); err != nil {
		t.Fatal(err)
	}

	repo := repository.NewConveneHistorySQLite(db.SQL())
	payload := domain.ConveneImportPayload{
		PlayerID:     "500000001",
		ServerID:     "6",
		Region:       "global",
		LanguageCode: "pt",
		Pools: []domain.ConvenePoolDefinition{{
			PoolType:  1,
			Name:      "Invocação de Ressonadores em Destaque",
			ShortName: "Ressonadores",
			Kind:      "featured_character",
			HardPity:  80,
		}},
		Pulls: []domain.ConvenePull{{
			PoolType:     1,
			ResourceID:   "1610",
			ResourceType: "Ressonante",
			ItemName:     "Yangyang: Xuanling",
			Rarity:       5,
			Quantity:     1,
			ObtainedAt:   "2026-07-11 15:23:00",
			Fingerprint:  "pull-1610",
		}},
	}
	_, imported, duplicates, err := repo.SaveImportedPulls(context.Background(), payload)
	if err != nil {
		t.Fatal(err)
	}
	if imported != 1 || duplicates != 0 {
		t.Fatalf("first import = %d imported, %d duplicates", imported, duplicates)
	}
	_, imported, duplicates, err = repo.SaveImportedPulls(context.Background(), payload)
	if err != nil {
		t.Fatal(err)
	}
	if imported != 0 || duplicates != 1 {
		t.Fatalf("duplicate import = %d imported, %d duplicates", imported, duplicates)
	}

	pulls, err := repo.ListConvenePulls(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(pulls) != 1 {
		t.Fatalf("pulls = %d, want 1", len(pulls))
	}
	if pulls[0].ResourceType != "character" {
		t.Fatalf("resource type = %q, want character", pulls[0].ResourceType)
	}
	if pulls[0].IconPath != "/cache/characters/1610.webp" {
		t.Fatalf("icon path = %q", pulls[0].IconPath)
	}

	if err := repo.DeleteConveneHistory(context.Background()); err != nil {
		t.Fatal(err)
	}
	var profiles, savedPulls, pools int
	if err := db.SQL().QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM convene_profiles),
			(SELECT COUNT(*) FROM convene_pulls),
			(SELECT COUNT(*) FROM convene_pool_catalog)`,
	).Scan(&profiles, &savedPulls, &pools); err != nil {
		t.Fatal(err)
	}
	if profiles != 0 || savedPulls != 0 || pools != 0 {
		t.Fatalf("remaining history: profiles=%d pulls=%d pools=%d", profiles, savedPulls, pools)
	}
}
