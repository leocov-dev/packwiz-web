package tables

import "time"

// UserIdentity links a User to an account at an OidcProvider. Identities are
// matched on ProviderID + Subject, never on email.
type UserIdentity struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	UserID      uint       `json:"userId"`
	ProviderID  uint       `json:"providerId"`
	Subject     string     `json:"subject"`
	Email       string     `json:"email"`
	CreatedAt   time.Time  `json:"createdAt"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
}

func (UserIdentity) TableName() string {
	return "user_identities"
}
