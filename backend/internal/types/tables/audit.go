package tables

import "time"

type Audit struct {
	ID           uint `gorm:"primarykey"`
	CreatedAt    time.Time
	UserID       uint
	Action       string
	ActionParams string `gorm:"type:json"`
}
