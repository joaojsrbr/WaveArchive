package usecase

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"wavearchive/internal/database"
	"wavearchive/internal/domain"
	"wavearchive/internal/repository"
)

func TestCalculateProgressionAggregatesAscensionSkillsAndPassiveNodes(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "progression.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := repository.NewCharacterSQLite(db.SQL())
	credit := domain.Material{ID: 2, Name: "Shell Credit"}
	profile := domain.CharacterProfile{
		Character:    domain.Character{ID: 1606, Name: "Roccia"},
		DetailLoaded: true,
		Skills: []domain.Skill{
			{NodeID: "1", Name: "Normal Attack"},
			{NodeID: "4", Name: "Inherent Skill"},
			{NodeID: "5", Name: "Second Inherent Skill"},
			{NodeID: "8", Name: "Outro Skill"},
		},
		Progression: domain.CharacterProgression{
			Ascensions: []domain.AscensionStage{{Stage: 1, UnlockLevel: 20, Costs: []domain.MaterialCost{{Material: credit, Quantity: 5000}}}},
			Skills: []domain.SkillProgression{
				{NodeID: "1", Name: "Normal Attack", MaxLevel: 2, UnlockCosts: []domain.MaterialCost{{Material: credit, Quantity: 999}}, LevelCosts: []domain.SkillLevelCost{{Level: 2, Costs: []domain.MaterialCost{{Material: credit, Quantity: 1500}}}}},
				{NodeID: "4", Name: "Stat Node", NodeType: 4, MaxLevel: 1, UnlockCosts: []domain.MaterialCost{{Material: credit, Quantity: 3900}}},
				{NodeID: "5", Name: "Inherent Skill", NodeType: 3, MaxLevel: 1, UnlockCosts: []domain.MaterialCost{{Material: credit, Quantity: 3900}}, LevelCosts: []domain.SkillLevelCost{{Level: 1, Costs: []domain.MaterialCost{{Material: credit, Quantity: 10000}}}}},
				{NodeID: "8", Name: "Outro Skill", NodeType: 3, MaxLevel: 1, UnlockCosts: []domain.MaterialCost{{Material: credit, Quantity: 3900}}},
			},
		},
	}
	if err := repo.ReplaceSynced(context.Background(), "3.6.1", []domain.CharacterProfile{profile}); err != nil {
		t.Fatal(err)
	}
	catalog := NewCharacterCatalog(repo, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	plan, err := catalog.CalculateProgression(context.Background(), domain.ProgressionPlanRequest{
		CharacterID: 1606, CurrentLevel: 1, TargetLevel: 40,
		CurrentSkills: map[string]int{"1": 1}, TargetSkills: map[string]int{"1": 2},
		IncludeUnlocks: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Total) != 1 || plan.Total[0].Quantity != 20400 {
		t.Fatalf("total = %#v, want 20,400 Shell Credits", plan.Total)
	}
}
