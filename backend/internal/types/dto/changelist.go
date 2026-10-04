package dto

// Changelist periods. A changelist shows the last 10 days by day, the rest of
// the last 3 calendar months by month, and everything older by year.
const (
	ChangelistDay   = "day"
	ChangelistMonth = "month"
	ChangelistYear  = "year"
	// ChangelistPending is the not yet published changes of a draft.
	ChangelistPending = "pending"
)

// ChangelistMod is a mod added to or removed from a pack.
type ChangelistMod struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ChangelistModChange is a mod present before and after a period that changed.
// FromVersion/ToVersion differ when the mod was updated; Fields lists the
// player-relevant settings that changed (side, optional, ...).
type ChangelistModChange struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	FromVersion string   `json:"fromVersion"`
	ToVersion   string   `json:"toVersion"`
	Updated     bool     `json:"updated"`
	Fields      []string `json:"fields"`
}

// ChangelistEntry is the net change of a pack over one period: the state at
// the end of the period compared with the state at the end of the previous
// one. Start and End are UTC dates (YYYY-MM-DD) and may cover only part of a
// month or year where a finer period takes over.
type ChangelistEntry struct {
	Period string `json:"period"`
	Start  string `json:"start"`
	End    string `json:"end"`
	// Initial means there is no earlier history: the entry describes the pack
	// as first recorded instead of listing every mod as added.
	Initial  bool                  `json:"initial"`
	ModCount int                   `json:"modCount"`
	Pack     []SnapshotFieldChange `json:"pack"`
	Added    []ChangelistMod       `json:"added"`
	Removed  []ChangelistMod       `json:"removed"`
	Changed  []ChangelistModChange `json:"changed"`
}

// PackChangelistResponse lists changelist entries, newest first.
type PackChangelistResponse struct {
	Entries []ChangelistEntry `json:"entries"`
}
