package dto

import (
	"errors"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
)

// BulkAddRequest is a pasted list of mods, one name, slug or mod page URL per
// line. List order is priority: earlier lines win conflicts.
type BulkAddRequest struct {
	Text string `json:"text" validate:"required,max=50000"`
}

func (r *BulkAddRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}

// Bulk add item review actions.
const (
	BulkItemActionPick   = "pick"
	BulkItemActionReject = "reject"
)

// ResolveBulkItemRequest reviews one line: pick a mod for it, or reject it.
// A Modrinth pick needs Slug or ProjectId; a CurseForge pick needs ProjectId.
type ResolveBulkItemRequest struct {
	Action    string `json:"action" validate:"required,oneof=pick reject"`
	Source    string `json:"source" validate:"omitempty,oneof=modrinth curseforge"`
	Slug      string `json:"slug"`
	ProjectId string `json:"projectId"`
}

func (r *ResolveBulkItemRequest) Validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(r); err != nil {
		return err
	}
	if r.Action != BulkItemActionPick {
		return nil
	}
	switch r.Source {
	case "modrinth":
		if r.Slug == "" && r.ProjectId == "" {
			return errors.New("a modrinth pick needs a slug or projectId")
		}
	case "curseforge":
		if _, err := strconv.ParseUint(r.ProjectId, 10, 32); err != nil {
			return errors.New("a curseforge pick needs a numeric projectId")
		}
	default:
		return errors.New("a pick needs a source")
	}
	return nil
}

// BulkAddCandidate is a search hit offered for review of a line.
type BulkAddCandidate struct {
	Source    string `json:"source"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	ProjectId string `json:"projectId"`
	IconUrl   string `json:"iconUrl,omitempty"`
}

// BulkAddItem is the state of one line of a bulk add.
type BulkAddItem struct {
	Id                    uint               `json:"id"`
	Position              int                `json:"position"`
	Line                  string             `json:"line"`
	Status                string             `json:"status"`
	Source                string             `json:"source"`
	Slug                  string             `json:"slug"`
	Name                  string             `json:"name"`
	ProjectId             string             `json:"projectId"`
	Reason                string             `json:"reason"`
	ConflictsWithPosition int                `json:"conflictsWithPosition"`
	Candidates            []BulkAddCandidate `json:"candidates"`
}

// BulkAddRunResponse is a bulk add run with its lines in list order.
type BulkAddRunResponse struct {
	Id         uint          `json:"id"`
	Status     string        `json:"status"`
	Error      string        `json:"error"`
	CreatedAt  time.Time     `json:"createdAt"`
	StartedAt  *time.Time    `json:"startedAt"`
	FinishedAt *time.Time    `json:"finishedAt"`
	Items      []BulkAddItem `json:"items"`
}
