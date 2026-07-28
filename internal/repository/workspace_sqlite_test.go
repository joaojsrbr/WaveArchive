package repository

import (
	"context"
	"path/filepath"
	"testing"

	"wavearchive/internal/database"
	"wavearchive/internal/domain"
)

func TestWorkspaceSettingsGoalsAndConvenes(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "workspace.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewWorkspaceSQLite(db.SQL())
	ctx := context.Background()
	settings, err := repo.GetSettings(ctx)
	if err != nil || settings.AIProvider != "ollama" {
		t.Fatalf("settings = %+v, %v", settings, err)
	}
	settings.AIProvider, settings.AIEndpoint, settings.AIModel = "lmstudio", "http://127.0.0.1:1234", "test"
	if _, err := repo.SaveSettings(ctx, settings); err != nil {
		t.Fatal(err)
	}
	goal, err := repo.SaveGoal(ctx, domain.PlannerGoal{Title: "Elevar personagem", GoalType: "character", RequiredAmount: 10, OwnedAmount: 4, Priority: 1})
	if err != nil || goal.ID == 0 {
		t.Fatalf("goal = %+v, %v", goal, err)
	}
	record, err := repo.SaveConvene(ctx, domain.ConveneRecord{Banner: "Teste", ItemName: "Sanhua", Rarity: 4, PullNumber: 9, ObtainedAt: "2026-07-26"})
	if err != nil || record.ID == 0 {
		t.Fatalf("record = %+v, %v", record, err)
	}
	summary, err := repo.Dashboard(ctx)
	if err != nil || len(summary.Goals) != 1 {
		t.Fatalf("summary = %+v, %v", summary, err)
	}
}
