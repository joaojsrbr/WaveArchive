package domain

import "context"

type ConveneProfile struct {
	ID             int64  `json:"id"`
	PlayerID       string `json:"playerId"`
	ServerID       string `json:"serverId"`
	Region         string `json:"region"`
	LanguageCode   string `json:"languageCode"`
	LastImportedAt string `json:"lastImportedAt"`
	HistoryPartial bool   `json:"historyPartial"`
}

type ConvenePull struct {
	ID           int64  `json:"id"`
	ProfileID    int64  `json:"profileId"`
	PoolType     int    `json:"poolType"`
	PoolName     string `json:"poolName"`
	ResourceID   string `json:"resourceId"`
	ResourceType string `json:"resourceType"`
	ItemName     string `json:"itemName"`
	Rarity       int    `json:"rarity"`
	Quantity     int    `json:"quantity"`
	ObtainedAt   string `json:"obtainedAt"`
	SourceIndex  int    `json:"sourceIndex"`
	Fingerprint  string `json:"-"`
	IconPath     string `json:"iconPath"`
}

type ConvenePoolSummary struct {
	PoolType       int           `json:"poolType"`
	Name           string        `json:"name"`
	ShortName      string        `json:"shortName"`
	Kind           string        `json:"kind"`
	Total          int           `json:"total"`
	Count5         int           `json:"count5"`
	Count4         int           `json:"count4"`
	Count3         int           `json:"count3"`
	CurrentPity    int           `json:"currentPity"`
	HardPity       int           `json:"hardPity"`
	CurrentPity4   int           `json:"currentPity4"`
	AveragePity5   float64       `json:"averagePity5"`
	GuaranteeState string        `json:"guaranteeState"`
	HistoryPartial bool          `json:"historyPartial"`
	RecentFiveStar []ConvenePull `json:"recentFiveStar"`
}

type ConvenePoolDefinition struct {
	PoolType  int    `json:"poolType"`
	LocaleKey string `json:"localeKey"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Kind      string `json:"kind"`
	HardPity  int    `json:"hardPity"`
	SortOrder int    `json:"sortOrder"`
}

type ConveneOverview struct {
	Profile        *ConveneProfile      `json:"profile,omitempty"`
	Pools          []ConvenePoolSummary `json:"pools"`
	Pulls          []ConvenePull        `json:"pulls"`
	Total          int                  `json:"total"`
	Count5         int                  `json:"count5"`
	Count4         int                  `json:"count4"`
	Count3         int                  `json:"count3"`
	LastImportedAt string               `json:"lastImportedAt"`
}

type ConveneImportResult struct {
	Imported       int             `json:"imported"`
	Duplicates     int             `json:"duplicates"`
	PoolsUpdated   int             `json:"poolsUpdated"`
	Profile        ConveneProfile  `json:"profile"`
	Overview       ConveneOverview `json:"overview"`
	Source         string          `json:"source"`
	HistoryPartial bool            `json:"historyPartial"`
}

type ConveneImportPayload struct {
	PlayerID     string
	ServerID     string
	Region       string
	LanguageCode string
	Pulls        []ConvenePull
	Pools        []ConvenePoolDefinition
}

type ConveneHistoryRepository interface {
	SaveImportedPulls(ctx context.Context, payload ConveneImportPayload) (ConveneProfile, int, int, error)
	GetConveneProfile(ctx context.Context) (*ConveneProfile, error)
	ListConvenePulls(ctx context.Context) ([]ConvenePull, error)
	ListConvenePoolDefinitions(ctx context.Context, profileID int64) ([]ConvenePoolDefinition, error)
	DeleteConveneHistory(ctx context.Context) error
}
