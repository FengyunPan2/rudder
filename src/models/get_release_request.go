package models

type GetReleaseRequest struct {
	Name         string        `json:"name"`
	Revision     int32         `json:"revision"`
	Max          int32         `json:"max"`
}
