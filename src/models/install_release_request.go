package models

// InstallReleaseRequest is the request body needed for installing a new release
type InstallReleaseRequest struct {
	Name         string        `json:"name"`
	Namespace    string        `json:"namespace"`
	Chart        string        `json:"chartPath"`
	DryRun       bool          `json:"dryRun"`
	DisableHooks bool          `json:"disableHooks"`
	Replace      bool          `json:"replace"`
	NameTemplate string        `json:"nameTemplate"`
	Verify       bool          `json:"verify"`
	Keyring      string        `json:"keyring"`
	Version      string        `json:"version"`
	Timeout      int64         `json:"timeout"`
	Wait         bool          `json:"wait"`
}
