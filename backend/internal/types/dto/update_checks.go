package dto

import "time"

// Update check statuses reported by UpdateChecksResponse and UpdateCheckJobResponse.
const (
	UpdateCheckIdle    = "idle"
	UpdateCheckQueued  = "queued"
	UpdateCheckRunning = "running"
	UpdateCheckDone    = "done"
	UpdateCheckFailed  = "failed"
)

// UpdateCheckJobResponse is returned when a check is requested.
type UpdateCheckJobResponse struct {
	JobId  int64  `json:"jobId"`
	Status string `json:"status"`
}

// UpdateCheckItem is the latest check outcome of one mod.
type UpdateCheckItem struct {
	ModId           uint   `json:"modId"`
	UpdateAvailable bool   `json:"updateAvailable"`
	UpdateString    string `json:"updateString,omitempty"`
	// Error is set when checking this mod failed; UpdateAvailable is then false.
	Error string `json:"error,omitempty"`
}

// UpdateChecksResponse is the stored state of a pack's update check.
type UpdateChecksResponse struct {
	// Status is one of idle, queued, running, done, failed.
	Status    string     `json:"status"`
	CheckedAt *time.Time `json:"checkedAt"`
	// Error is the reason a whole check run failed.
	Error   string            `json:"error,omitempty"`
	Results []UpdateCheckItem `json:"results"`
	// AvailableCount counts mods with an update that are not pinned.
	AvailableCount int `json:"availableCount"`
}
