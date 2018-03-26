package models

type UpdateRelease struct {
	Rollback        bool             `json:"rollback"`
	Revision        int32            `json:"revision"`
	Release         string           `json:"release"`
	Chart           string           `json:"chart"`
	DryRun          bool             `json:"dryRun"`
	Recreate        bool             `json:"recreate"`
	DisableHooks    bool             `json:"disableHooks"`
	Verify          bool             `json:"verify"`
	Keyring         string           `json:"keyring"`
	Install         bool             `json:"install"`
	Namespace       string           `json:"namespace"`
	Version         string           `json:"version"`
	Timeout         int64            `json:"timeout"`
	ResetValues     bool             `json:"resetValues"`
	ReuseValues     bool             `json:"reuseValues"`
	Wait            bool             `json:"wait"`
}
