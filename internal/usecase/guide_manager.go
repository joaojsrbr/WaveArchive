package usecase

import (
	"context"
	"errors"
	"strings"

	"wavearchive/internal/domain"
	guidesource "wavearchive/internal/sources/guide"
)

type GuideManager struct {
	repository domain.GuideRepository
	client     *guidesource.Client
}

func NewGuideManager(repository domain.GuideRepository, client *guidesource.Client) *GuideManager {
	return &GuideManager{repository: repository, client: client}
}
func (m *GuideManager) List(ctx context.Context, characterID int64) ([]domain.CharacterGuide, error) {
	return m.repository.List(ctx, characterID)
}
func (m *GuideManager) ListAll(ctx context.Context) ([]domain.CharacterGuide, error) {
	return m.repository.ListAll(ctx)
}
func (m *GuideManager) Sync(ctx context.Context, characterID int64, language string) ([]domain.CharacterGuide, error) {
	if characterID <= 0 {
		return nil, errors.New("character is required")
	}
	language = strings.TrimSpace(language)
	if language == "" {
		language = "en"
	}
	guides, err := m.client.Fetch(ctx, characterID, language)
	if err != nil {
		return nil, err
	}
	if err := m.repository.Replace(ctx, characterID, guides); err != nil {
		return nil, err
	}
	return m.repository.List(ctx, characterID)
}
func (m *GuideManager) Search(ctx context.Context, query string, limit int) ([]domain.KnowledgeSource, error) {
	return m.repository.Search(ctx, query, limit)
}
