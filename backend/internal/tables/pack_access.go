package tables

import "time"

// PackAccess is one request for a pack's pack.toml.
type PackAccess struct {
	ID         uint      `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	PackId     *uint     `json:"packId"`
	PackSlug   string    `json:"packSlug"`
	UserId     *uint     `json:"userId"`
	IpAddress  string    `json:"ipAddress"`
	UserAgent  string    `json:"userAgent"`
	StatusCode int       `json:"statusCode"`
	Success    bool      `json:"success"`
}
