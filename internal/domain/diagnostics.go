package domain

type Diagnostics struct {
	DataDirectory string `json:"dataDirectory"`
	DatabasePath  string `json:"databasePath"`
	DatabaseBytes int64  `json:"databaseBytes"`
	Migrations    int    `json:"migrations"`
	GameVersion   string `json:"gameVersion"`
	CatalogCount  int    `json:"catalogCount"`
	GoVersion     string `json:"goVersion"`
}
