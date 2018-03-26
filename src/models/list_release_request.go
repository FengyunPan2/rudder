package models

type ListRelease struct {
	Filter     string    `json:"filter"`
	Short      bool      `json:"short"`
	ByDate     bool      `json:"byDate"`
	SortDesc   bool      `json:"sortDesc"`
	Limit      int       `json:"limit"`
	Offset     string    `json:"offset"`
	All        bool      `json:"all"`
	Deleted    bool      `json:"deleted"`
	Deleting   bool      `json:"deleting"`
	Deployed   bool      `json:"deployed"`
	Failed     bool      `json:"failed"`
	Superseded bool      `json:"superseded"`
	Namespace  string    `json:"namespace"`
}