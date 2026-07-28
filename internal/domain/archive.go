package domain

type ArchiveBundle struct {
	SchemaVersion int               `json:"schemaVersion"`
	ExportedAt    string            `json:"exportedAt"`
	Builds        []Build           `json:"builds"`
	Teams         []Team            `json:"teams"`
	Theorycraft   []TeamTheorycraft `json:"theorycraft"`
	Settings      AppSettings       `json:"settings"`
	Goals         []PlannerGoal     `json:"goals"`
	Convenes      []ConveneRecord   `json:"convenes"`
}
type ArchiveReport struct {
	Builds    int      `json:"builds"`
	Teams     int      `json:"teams"`
	Buffs     int      `json:"buffs"`
	Rotations int      `json:"rotations"`
	Goals     int      `json:"goals"`
	Convenes  int      `json:"convenes"`
	Warnings  []string `json:"warnings"`
}
