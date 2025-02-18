package tables

import "time"

type Audit struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UserID       uint      `json:"userId"`
	Action       string    `json:"action"`
	ActionParams string    `gorm:"type:json" json:"actionParams"`
}
