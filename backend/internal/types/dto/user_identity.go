package dto

import "time"

// UserIdentityResponse is a linked external account as shown on the profile
// and admin user pages.
type UserIdentityResponse struct {
	ID           uint       `json:"id"`
	ProviderID   uint       `json:"providerId"`
	ProviderSlug string     `json:"providerSlug"`
	ProviderName string     `json:"providerName"`
	Email        string     `json:"email"`
	CreatedAt    time.Time  `json:"createdAt"`
	LastLoginAt  *time.Time `json:"lastLoginAt"`
}

// LinkIdentityResponse carries the provider authorization URL the browser must
// navigate to in order to link an account.
type LinkIdentityResponse struct {
	RedirectUrl string `json:"redirectUrl"`
}
