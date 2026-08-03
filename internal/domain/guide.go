package domain

import "context"

type CharacterGuide struct {
	ID          string    `json:"id"`
	CharacterID int64     `json:"characterId"`
	Name        string    `json:"name"`
	Source      string    `json:"source"`
	LikeCount   int       `json:"likeCount"`
	Language    string    `json:"language"`
	Teams       [][]int64 `json:"teams"`
	DataJSON    string    `json:"dataJson"`
	SyncedAt    string    `json:"syncedAt"`
}

type BuildExportIcons struct {
	ElementIconPath    string `json:"elementIconPath"`
	WeaponTypeIconPath string `json:"weaponTypeIconPath"`
}
type KnowledgeSource struct {
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
	Title      string `json:"title"`
	Snippet    string `json:"snippet"`
}
type AIProviderStatus struct {
	Provider string   `json:"provider"`
	Online   bool     `json:"online"`
	Models   []string `json:"models"`
	Message  string   `json:"message"`
}
type GuideRepository interface {
	List(ctx context.Context, characterID int64) ([]CharacterGuide, error)
	ListAll(ctx context.Context) ([]CharacterGuide, error)
	Replace(ctx context.Context, characterID int64, guides []CharacterGuide) error
	Search(ctx context.Context, query string, limit int) ([]KnowledgeSource, error)
}
