package domain

import "context"

type AppSettings struct {
	Density          string `json:"density"`
	SidebarCollapsed bool   `json:"sidebarCollapsed"`
	AIProvider       string `json:"aiProvider"`
	AIEndpoint       string `json:"aiEndpoint"`
	AIModel          string `json:"aiModel"`
	AIMode           string `json:"aiMode"`
	ReduceMotion     bool   `json:"reduceMotion"`
}

type AccountSummary struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Notes           string `json:"notes"`
	Astrite         int    `json:"astrite"`
	RadiantTides    int    `json:"radiantTides"`
	OwnedCharacters int    `json:"ownedCharacters"`
	OwnedWeapons    int    `json:"ownedWeapons"`
	OwnedEchoes     int    `json:"ownedEchoes"`
}

type PlannerGoal struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	GoalType       string `json:"goalType"`
	TargetName     string `json:"targetName"`
	RequiredAmount int    `json:"requiredAmount"`
	OwnedAmount    int    `json:"ownedAmount"`
	ShellCredits   int    `json:"shellCredits"`
	Priority       int    `json:"priority"`
	DueDate        string `json:"dueDate"`
	Completed      bool   `json:"completed"`
	Notes          string `json:"notes"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type ConveneRecord struct {
	ID         int64  `json:"id"`
	Banner     string `json:"banner"`
	BannerType string `json:"bannerType"`
	ItemName   string `json:"itemName"`
	Rarity     int    `json:"rarity"`
	PullNumber int    `json:"pullNumber"`
	Guaranteed bool   `json:"guaranteed"`
	ObtainedAt string `json:"obtainedAt"`
	Notes      string `json:"notes"`
}

type Enemy struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Level            int     `json:"level"`
	Resistance       float64 `json:"resistance"`
	DamageReduction  float64 `json:"damageReduction"`
	ElementReduction float64 `json:"elementReduction"`
	Notes            string  `json:"notes"`
}

type FormulaVersion struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	GameVersion     string  `json:"gameVersion"`
	DefenseConstant float64 `json:"defenseConstant"`
	LevelFactor     float64 `json:"levelFactor"`
	Confidence      string  `json:"confidence"`
	References      string  `json:"references"`
	RoundingPolicy  string  `json:"roundingPolicy"`
	Active          bool    `json:"active"`
}

type DashboardSummary struct {
	Characters   int     `json:"characters"`
	Weapons      int     `json:"weapons"`
	Echoes       int     `json:"echoes"`
	Builds       int     `json:"builds"`
	Teams        int     `json:"teams"`
	RecentBuilds []Build `json:"recentBuilds"`
}

type BuildVersion struct {
	ID        int64  `json:"id"`
	BuildID   int64  `json:"buildId"`
	Snapshot  string `json:"snapshot"`
	CreatedAt string `json:"createdAt"`
}

type WorkspaceRepository interface {
	GetSettings(ctx context.Context) (AppSettings, error)
	SaveSettings(ctx context.Context, settings AppSettings) (AppSettings, error)
	GetAccount(ctx context.Context) (AccountSummary, error)
	SaveAccount(ctx context.Context, account AccountSummary) (AccountSummary, error)
	ListGoals(ctx context.Context) ([]PlannerGoal, error)
	SaveGoal(ctx context.Context, goal PlannerGoal) (PlannerGoal, error)
	DeleteGoal(ctx context.Context, id int64) error
	ListConvenes(ctx context.Context) ([]ConveneRecord, error)
	SaveConvene(ctx context.Context, record ConveneRecord) (ConveneRecord, error)
	DeleteConvene(ctx context.Context, id int64) error
	ListEnemies(ctx context.Context) ([]Enemy, error)
	ListFormulaVersions(ctx context.Context) ([]FormulaVersion, error)
	Dashboard(ctx context.Context) (DashboardSummary, error)
}
