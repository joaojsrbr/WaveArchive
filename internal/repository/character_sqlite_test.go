package repository_test

import (
	"context"
	"path/filepath"
	"testing"

	"wavearchive/internal/database"
	"wavearchive/internal/domain"
	"wavearchive/internal/repository"
)

func TestCharacterCatalogPersistsAndSearchesWithFTS(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewCharacterSQLite(db.SQL())
	characters := []domain.CharacterProfile{
		{
			Character:       domain.Character{ID: 101, Name: "Cantarella", Nickname: "Bane", Rarity: 5, Element: domain.ElementHavoc, APIOrder: 2},
			Description:     "A mysterious resonator.",
			DetailLoaded:    true,
			Skills:          []domain.Skill{{NodeID: "1", Type: "Normal Attack", Name: "Basic Attack", Description: "Deals damage."}},
			Chains:          []domain.ResonanceChain{{Sequence: 1, Name: "First Chain", Description: "Improves the kit."}},
			SignatureWeapon: &domain.Weapon{ID: 2001, Name: "Whispers", Rarity: 5, Type: 5, Effect: "Increases damage."},
			Progression: domain.CharacterProgression{
				Ascensions: []domain.AscensionStage{{Stage: 1, UnlockLevel: 20, Costs: []domain.MaterialCost{{Material: domain.Material{ID: 2, Name: "Shell Credit", Rarity: 3, Sources: []string{"Simulation Training"}}, Quantity: 5000}}}},
				Skills:     []domain.SkillProgression{{NodeID: "1", NodeType: 2, Type: "Normal Attack", Name: "Basic Attack", MaxLevel: 2, LevelCosts: []domain.SkillLevelCost{{Level: 2, Costs: []domain.MaterialCost{{Material: domain.Material{ID: 2, Name: "Shell Credit", Rarity: 3}, Quantity: 1500}}}}, Values: []domain.SkillValueRow{{Name: "Damage", Values: []string{"10%", "12%"}}}}},
				LevelEXP:   []int{0, 400},
				Stats:      []domain.CharacterStat{{Ascension: 0, Level: 1, HP: 100, ATK: 20, DEF: 30}},
			},
			Extras: domain.CharacterExtras{Tags: []domain.CharacterTag{{ID: 8, Name: "Traction"}}, Stories: []domain.LoreEntry{{Title: "Story", Content: "Lore"}}},
		},
		{Character: domain.Character{ID: 102, Name: "Baizhi", Nickname: "Researcher", Rarity: 4, Element: domain.ElementGlacio, APIOrder: 1}},
	}
	if err := repo.ReplaceSynced(context.Background(), "3.5.3", characters); err != nil {
		t.Fatal(err)
	}

	got, err := repo.List(context.Background(), domain.CharacterFilter{Query: "Cant", Element: int(domain.ElementHavoc)})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Cantarella" {
		t.Fatalf("unexpected search result: %#v", got)
	}
	apiOrder, err := repo.List(context.Background(), domain.CharacterFilter{Sort: "api"})
	if err != nil {
		t.Fatal(err)
	}
	if len(apiOrder) != 2 || apiOrder[0].Name != "Baizhi" || apiOrder[1].Name != "Cantarella" {
		t.Fatalf("unexpected API order: %#v", apiOrder)
	}

	status, err := repo.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Count != 2 || status.Version != "3.5.3" || status.LastSyncAt == nil {
		t.Fatalf("unexpected status: %#v", status)
	}

	profile, err := repo.GetProfile(context.Background(), 101)
	if err != nil {
		t.Fatal(err)
	}
	if profile.Description == "" || len(profile.Skills) != 1 || len(profile.Chains) != 1 {
		t.Fatalf("unexpected profile: %#v", profile)
	}
	if profile.SignatureWeapon == nil || profile.SignatureWeapon.Name != "Whispers" {
		t.Fatalf("unexpected signature weapon: %#v", profile.SignatureWeapon)
	}
	if len(profile.Progression.Ascensions) != 1 || profile.Progression.Ascensions[0].Costs[0].Quantity != 5000 {
		t.Fatalf("unexpected ascension data: %#v", profile.Progression.Ascensions)
	}
	if len(profile.Progression.Skills) != 1 || profile.Progression.Skills[0].Values[0].Values[1] != "12%" {
		t.Fatalf("unexpected skill progression: %#v", profile.Progression.Skills)
	}
	if len(profile.Extras.Tags) != 1 || profile.Extras.Stories[0].Content != "Lore" {
		t.Fatalf("unexpected extras: %#v", profile.Extras)
	}

	if err := repo.UpdateAccount(context.Background(), domain.CharacterAccountUpdate{
		CharacterID: 101,
		Owned:       false,
		Level:       70,
		Sequence:    2,
		Favorite:    true,
	}); err != nil {
		t.Fatal(err)
	}
	favorites, err := repo.List(context.Background(), domain.CharacterFilter{Favorites: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(favorites) != 1 || favorites[0].Owned || !favorites[0].Favorite {
		t.Fatalf("favorite should not imply ownership: %#v", favorites)
	}
	owned, err := repo.List(context.Background(), domain.CharacterFilter{OwnedOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(owned) != 0 {
		t.Fatalf("unexpected owned characters: %#v", owned)
	}
}
