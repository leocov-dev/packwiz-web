package tables

import "time"

// OidcProvider is an admin-managed OpenID Connect identity provider.
type OidcProvider struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	Slug            string    `gorm:"unique" json:"slug"`
	DisplayName     string    `json:"displayName"`
	IssuerURL       string    `gorm:"column:issuer_url" json:"issuerUrl"`
	ClientID        string    `gorm:"column:client_id" json:"clientId"`
	ClientSecretEnc string    `json:"-"`
	Scopes          string    `json:"scopes"`
	Enabled         bool      `json:"enabled"`
	AutoCreateUsers bool      `json:"autoCreateUsers"`
	LinkByEmail     bool      `json:"linkByEmail"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (OidcProvider) TableName() string {
	return "oidc_providers"
}
