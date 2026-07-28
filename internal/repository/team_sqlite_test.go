package repository_test

import (
	"context"
	"path/filepath"
	"testing"

	"wavearchive/internal/database"
	"wavearchive/internal/domain"
	"wavearchive/internal/repository"
	"wavearchive/internal/usecase"
)

func TestTeamLifecycle(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(filepath.Join(t.TempDir(), "teams.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	characters := repository.NewCharacterSQLite(db.SQL())
	profiles := []domain.CharacterProfile{
		{Character: domain.Character{ID: 1, Name: "Jinhsi", Rarity: 5}},
		{Character: domain.Character{ID: 2, Name: "Zhezhi", Rarity: 5}},
		{Character: domain.Character{ID: 3, Name: "Shorekeeper", Rarity: 5}},
	}
	if err := characters.ReplaceSynced(ctx, "3.6.1", profiles); err != nil {
		t.Fatal(err)
	}
	manager := usecase.NewTeamManager(repository.NewTeamSQLite(db.SQL()))
	team, err := manager.Save(ctx, domain.Team{
		Name: "Spectro coordinated", GameVersion: "3.6.1",
		Members: []domain.TeamMember{
			{CharacterID: 1, Role: "main_dps"},
			{CharacterID: 2, Role: "sub_dps"},
			{CharacterID: 3, Role: "support"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(team.Members) != 3 || team.Members[1].CharacterName != "Zhezhi" {
		t.Fatalf("unexpected team: %#v", team)
	}
	copy, err := manager.Duplicate(ctx, team.ID)
	if err != nil || copy.ID == team.ID || copy.Name != "Spectro coordinated — cópia" {
		t.Fatalf("unexpected duplicate: %#v, %v", copy, err)
	}
	if err := manager.Delete(ctx, team.ID); err != nil {
		t.Fatal(err)
	}
	list, err := manager.List(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("soft delete failed: %#v, %v", list, err)
	}
	if err := manager.Restore(ctx, team.ID); err != nil {
		t.Fatal(err)
	}
	list, err = manager.List(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("restore failed: %#v, %v", list, err)
	}
}
