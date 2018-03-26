package models

type DeleteRelease struct {
	Name        string    `json:"name"`
	DryRun      bool      `json:"dry-run"`
	NoHooks     bool      `json:"no-hooks"`
	Purge       bool      `json:"purge"`
	Timeout     int       `json:"timeout"`
}
