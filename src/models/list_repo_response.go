package models

import (
	"time"
)

// Entry represents a collection of parameters for chart repository
type entry struct {
	Name     string `json:"name"`
	Cache    string `json:"cache"`
	URL      string `json:"url"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
	CAFile   string `json:"caFile"`
}

type ListRepo struct {
	APIVersion   string    `json:"apiVersion"`
	Generated    time.Time `json:"generated"`
	Repositories []*entry  `json:"repositories"`
}
