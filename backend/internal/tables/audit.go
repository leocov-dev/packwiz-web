package tables

import (
	"time"
)

type Audit struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UserId       uint      `json:"userId"`
	Action       string    `json:"action"`
	ActionParams string    `gorm:"type:json" json:"actionParams"`
	IpAddress    string    `json:"ipAddress"`
}
