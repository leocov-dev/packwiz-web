package dto

// UpdateModResponse is the result of updating a single mod.
type UpdateModResponse struct {
	// Updated is false when the mod was already up to date.
	Updated bool `json:"updated"`
}

// UpdateAllItem describes one mod in an UpdateAllResponse section.
type UpdateAllItem struct {
	ModId uint   `json:"modId"`
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	// FileName is the new file name (updated mods only).
	FileName string `json:"fileName,omitempty"`
	// Reason is why a mod was skipped, e.g. "pinned".
	Reason string `json:"reason,omitempty"`
	// Error is why a mod failed to update.
	Error string `json:"error,omitempty"`
}

// UpdateAllResponse is the partial-success summary of updating a whole pack.
type UpdateAllResponse struct {
	Updated []UpdateAllItem `json:"updated"`
	Skipped []UpdateAllItem `json:"skipped"`
	Failed  []UpdateAllItem `json:"failed"`
	// UpToDate counts mods that were checked and are current.
	UpToDate int `json:"upToDate"`
	// NotChecked counts mods whose status is unknown: manual sources with no
	// updater, or pinned mods whose update check failed.
	NotChecked int `json:"notChecked"`
}
