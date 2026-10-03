package dto

import (
	"github.com/go-playground/validator/v10"
)

// MaxAccessDays is the widest range the access metrics endpoints accept.
const MaxAccessDays = 365

// PackAccessSummaryQuery selects the trailing range, in days, for a summary.
type PackAccessSummaryQuery struct {
	Days int `form:"days" validate:"gte=1,lte=365"`
}

func (q *PackAccessSummaryQuery) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(q)
}

// PackAccessRecentQuery pages through individual access records.
type PackAccessRecentQuery struct {
	Days     int `form:"days" validate:"gte=1,lte=365"`
	Page     int `form:"page" validate:"gte=1"`
	PageSize int `form:"pageSize" validate:"gte=1,lte=100"`
	// Outcome and PackId are only honored by the system-wide endpoint.
	Outcome string `form:"outcome" validate:"omitempty,oneof=all success failure"`
	PackId  uint   `form:"packId"`
}

func (q *PackAccessRecentQuery) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(q)
}

// AccessDay is one UTC day of access counts.
type AccessDay struct {
	Date    string `json:"date"` // YYYY-MM-DD
	Success int64  `json:"success"`
	Failure int64  `json:"failure"`
}

type AccessUserCount struct {
	UserId   uint   `json:"userId"`
	Username string `json:"username"`
	Count    int64  `json:"count"`
}

type AccessIpCount struct {
	IpAddress string `json:"ipAddress"`
	Count     int64  `json:"count"`
}

type AccessPackCount struct {
	PackId  uint   `json:"packId"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Success int64  `json:"success"`
	Failure int64  `json:"failure"`
}

// AccessRecord is one row of the recent-access list.
type AccessRecord struct {
	ID         uint   `json:"id"`
	CreatedAt  string `json:"createdAt"`
	PackId     *uint  `json:"packId"`
	PackSlug   string `json:"packSlug"`
	PackName   string `json:"packName"`
	UserId     *uint  `json:"userId"`
	Username   string `json:"username"`
	IpAddress  string `json:"ipAddress"`
	UserAgent  string `json:"userAgent"`
	StatusCode int    `json:"statusCode"`
	Success    bool   `json:"success"`
}

// PackAccessSummary is the per-pack metrics view; successful requests only.
type PackAccessSummary struct {
	Days        int               `json:"days"`
	Series      []AccessDay       `json:"series"`
	Total       int64             `json:"total"`
	UniqueIps   int64             `json:"uniqueIps"`
	UniqueUsers int64             `json:"uniqueUsers"`
	TopUsers    []AccessUserCount `json:"topUsers"`
	TopIps      []AccessIpCount   `json:"topIps"`
}

// SystemAccessSummary is the system-wide metrics view; includes failures.
type SystemAccessSummary struct {
	Days        int               `json:"days"`
	Series      []AccessDay       `json:"series"`
	Success     int64             `json:"success"`
	Failure     int64             `json:"failure"`
	UniqueIps   int64             `json:"uniqueIps"`
	Packs       []AccessPackCount `json:"packs"`
	TopFailedIp []AccessIpCount   `json:"topFailedIps"`
}
