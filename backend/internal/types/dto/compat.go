package dto

import (
	"fmt"
	"sort"
	"time"

	"packwiz-web/internal/tables"
)

// Compat finding severities.
const (
	CompatSeverityError   = "error"
	CompatSeverityWarning = "warning"
	CompatSeverityInfo    = "info"
)

// CompatCounts are the finding counts of the stored scan result.
type CompatCounts struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Info     int `json:"info"`
}

// CompatFindingItem is one compatibility finding.
type CompatFindingItem struct {
	// ModId is null when the finding is about the pack and not one mod.
	ModId    *uint             `json:"modId"`
	Slug     string            `json:"slug"`
	Severity string            `json:"severity"`
	Code     string            `json:"code"`
	Message  string            `json:"message"`
	Detail   map[string]string `json:"detail"`
}

// CompatResponse is the stored jar compatibility state of a pack.
type CompatResponse struct {
	// Status is one of unknown, queued, scanning, stale, ok, warnings, errors, failed.
	Status string       `json:"status"`
	Counts CompatCounts `json:"counts"`
	// Unscanned counts mods whose jar could not be read.
	Unscanned int `json:"unscanned"`
	// ScannedAt is when the latest scan finished; null while none has.
	ScannedAt *time.Time `json:"scannedAt"`
	// RunError is a fixed user-safe reason the latest scan failed ("scan timed
	// out" or "scan failed"); empty otherwise.
	RunError string              `json:"runError"`
	Findings []CompatFindingItem `json:"findings"`
}

// CompatScanResponse is returned when a rescan is requested.
type CompatScanResponse struct {
	Status string `json:"status"`
	// Enqueued is false when a scan is already active or ran too recently.
	Enqueued bool `json:"enqueued"`
}

// CompatSummary is the small badge state shown in pack lists.
type CompatSummary struct {
	Status   string `json:"status"`
	Errors   int    `json:"errors"`
	Warnings int    `json:"warnings"`
}

// NewCompatResponse maps a stored run and findings to the response. run may be
// nil (never scanned). Findings are ordered by severity (errors first), slug,
// code, then message, so the order does not depend on the database.
func NewCompatResponse(status string, run *tables.PackCompatRun, findings []tables.PackCompatFinding) CompatResponse {
	resp := CompatResponse{Status: status, Findings: make([]CompatFindingItem, 0, len(findings))}
	if run != nil {
		resp.Counts = CompatCounts{Errors: run.ErrorCount, Warnings: run.WarningCount, Info: run.InfoCount}
		resp.Unscanned = run.UnscannedCount
		resp.ScannedAt = run.FinishedAt
		resp.RunError = run.Error
	}
	for _, f := range findings {
		resp.Findings = append(resp.Findings, CompatFindingItem{
			ModId:    f.ModID,
			Slug:     f.Slug,
			Severity: f.Severity,
			Code:     f.Code,
			Message:  f.Message,
			Detail:   detailStrings(f.Detail),
		})
	}
	sort.SliceStable(resp.Findings, func(i, j int) bool {
		a, b := resp.Findings[i], resp.Findings[j]
		if ra, rb := severityRank(a.Severity), severityRank(b.Severity); ra != rb {
			return ra > rb
		}
		if a.Slug != b.Slug {
			return a.Slug < b.Slug
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.Message < b.Message
	})
	return resp
}

func severityRank(s string) int {
	switch s {
	case CompatSeverityError:
		return 3
	case CompatSeverityWarning:
		return 2
	case CompatSeverityInfo:
		return 1
	}
	return 0
}

// detailStrings flattens a finding's JSON detail to string values. It is never
// nil, so the field serializes as an object.
func detailStrings(in map[string]any) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		if s, ok := v.(string); ok {
			out[k] = s
			continue
		}
		out[k] = fmt.Sprint(v)
	}
	return out
}
