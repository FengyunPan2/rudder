package models

type ListChart struct {
	Filter     string    `json:"filter"`
	Versions   bool      `json:"versions"`
	Regexp     bool      `json:"regexp"`
	Version    string    `json:"version"`
}