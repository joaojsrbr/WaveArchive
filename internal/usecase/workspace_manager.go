package usecase

import (
	"context"
	"errors"
	"strings"

	"wavearchive/internal/domain"
)

type WorkspaceManager struct{ repository domain.WorkspaceRepository }

func NewWorkspaceManager(repository domain.WorkspaceRepository) *WorkspaceManager {
	return &WorkspaceManager{repository: repository}
}
func (m *WorkspaceManager) Settings(ctx context.Context) (domain.AppSettings, error) {
	return m.repository.GetSettings(ctx)
}
func (m *WorkspaceManager) SaveSettings(ctx context.Context, s domain.AppSettings) (domain.AppSettings, error) {
	validDensity := map[string]bool{"compact": true, "comfortable": true, "spacious": true}
	validProvider := map[string]bool{"ollama": true, "lmstudio": true, "gemini": true}
	validMode := map[string]bool{"strict": true, "assisted": true, "general": true}
	if !validDensity[s.Density] {
		return s, errors.New("invalid density")
	}
	if !validProvider[s.AIProvider] {
		return s, errors.New("invalid AI provider")
	}
	if !validMode[s.AIMode] {
		return s, errors.New("invalid AI mode")
	}
	s.AIEndpoint = strings.TrimSpace(s.AIEndpoint)
	s.AIModel = strings.TrimSpace(s.AIModel)
	return m.repository.SaveSettings(ctx, s)
}
func (m *WorkspaceManager) Account(ctx context.Context) (domain.AccountSummary, error) {
	return m.repository.GetAccount(ctx)
}
func (m *WorkspaceManager) SaveAccount(ctx context.Context, a domain.AccountSummary) (domain.AccountSummary, error) {
	a.Name = strings.TrimSpace(a.Name)
	if a.Name == "" {
		return a, errors.New("account name is required")
	}
	if a.Astrite < 0 || a.RadiantTides < 0 {
		return a, errors.New("account resources cannot be negative")
	}
	return m.repository.SaveAccount(ctx, a)
}
func (m *WorkspaceManager) Goals(ctx context.Context) ([]domain.PlannerGoal, error) {
	return m.repository.ListGoals(ctx)
}
func (m *WorkspaceManager) SaveGoal(ctx context.Context, g domain.PlannerGoal) (domain.PlannerGoal, error) {
	g.Title = strings.TrimSpace(g.Title)
	if g.Title == "" {
		return g, errors.New("goal title is required")
	}
	if g.RequiredAmount < 0 || g.OwnedAmount < 0 {
		return g, errors.New("goal amounts cannot be negative")
	}
	if g.Priority < 1 || g.Priority > 3 {
		g.Priority = 2
	}
	return m.repository.SaveGoal(ctx, g)
}
func (m *WorkspaceManager) DeleteGoal(ctx context.Context, id int64) error {
	return m.repository.DeleteGoal(ctx, id)
}
func (m *WorkspaceManager) Convenes(ctx context.Context) ([]domain.ConveneRecord, error) {
	return m.repository.ListConvenes(ctx)
}
func (m *WorkspaceManager) SaveConvene(ctx context.Context, c domain.ConveneRecord) (domain.ConveneRecord, error) {
	c.ItemName = strings.TrimSpace(c.ItemName)
	c.Banner = strings.TrimSpace(c.Banner)
	if c.ItemName == "" || c.Banner == "" {
		return c, errors.New("banner and item are required")
	}
	if c.Rarity < 3 || c.Rarity > 5 {
		return c, errors.New("rarity must be between 3 and 5")
	}
	return m.repository.SaveConvene(ctx, c)
}
func (m *WorkspaceManager) DeleteConvene(ctx context.Context, id int64) error {
	return m.repository.DeleteConvene(ctx, id)
}
func (m *WorkspaceManager) Enemies(ctx context.Context) ([]domain.Enemy, error) {
	return m.repository.ListEnemies(ctx)
}
func (m *WorkspaceManager) Formulas(ctx context.Context) ([]domain.FormulaVersion, error) {
	return m.repository.ListFormulaVersions(ctx)
}
func (m *WorkspaceManager) Dashboard(ctx context.Context) (domain.DashboardSummary, error) {
	return m.repository.Dashboard(ctx)
}
