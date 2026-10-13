package tables

import "time"

// PackWebhook is the outgoing webhook of a pack (at most one per pack).
type PackWebhook struct {
	PackID  uint   `gorm:"primaryKey" json:"packId"`
	Enabled bool   `json:"enabled"`
	URL     string `gorm:"column:url" json:"url"`
	Method  string `json:"method"`
	// HeadersEnc is the secretbox-encrypted JSON object of request headers.
	HeadersEnc string `json:"-"`
	// LastStatusCode is the HTTP status of the last attempt (0: no response).
	LastStatusCode int        `json:"lastStatusCode"`
	LastError      string     `json:"lastError"`
	LastAttemptAt  *time.Time `json:"lastAttemptAt"`
	LastSuccessAt  *time.Time `json:"lastSuccessAt"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (PackWebhook) TableName() string {
	return "pack_webhooks"
}
