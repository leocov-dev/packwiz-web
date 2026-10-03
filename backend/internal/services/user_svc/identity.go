package user_svc

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
	"packwiz-web/internal/utils"
)

// ExternalUserInput describes the account an OIDC login should create.
type ExternalUserInput struct {
	ProviderID        uint
	Subject           string
	Email             string
	PreferredUsername string
	FullName          string
}

const maxUsernameLen = 32

// DeriveUsernameBase builds a safe, lowercase username candidate from the
// identity provider's preferred_username or, failing that, the local part of
// the email. It never returns a form of "admin".
func DeriveUsernameBase(preferred, email string) string {
	raw := strings.TrimSpace(preferred)
	if raw == "" {
		raw, _, _ = strings.Cut(strings.TrimSpace(email), "@")
	}

	var b strings.Builder
	for _, r := range strings.ToLower(raw) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		}
		if b.Len() >= maxUsernameLen {
			break
		}
	}

	name := b.String()
	if name == "" || name == "admin" {
		return "user"
	}
	return name
}

// CanUnlinkIdentity reports whether removing one identity still leaves the
// user with a way to sign in: a local password or another identity.
func CanUnlinkIdentity(hasPassword bool, otherIdentities int64) bool {
	return hasPassword || otherIdentities > 0
}

// FindIdentity returns the identity for a provider subject, or nil.
func (s *UserService) FindIdentity(providerID uint, subject string) (*tables.UserIdentity, error) {
	// Find+Limit rather than First: "not found" is the normal first-login case
	// and must not be logged as a failed query
	var found []tables.UserIdentity
	if err := s.db.Where("provider_id = ? AND subject = ?", providerID, subject).
		Limit(1).Find(&found).Error; err != nil {
		return nil, fmt.Errorf("find identity: %w", err)
	}
	if len(found) == 0 {
		return nil, nil
	}
	return &found[0], nil
}

// FindByEmail returns the non-deleted user with this email (case-insensitive),
// or nil.
func (s *UserService) FindByEmail(email string) (*tables.User, error) {
	var found []tables.User
	if err := s.db.Where("LOWER(email) = LOWER(?)", email).
		Limit(1).Find(&found).Error; err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	if len(found) == 0 {
		return nil, nil
	}
	return &found[0], nil
}

// HasIdentityAtProvider reports whether the user already has any identity at
// the provider.
func (s *UserService) HasIdentityAtProvider(userID, providerID uint) (bool, error) {
	var count int64
	if err := s.db.Model(&tables.UserIdentity{}).
		Where("user_id = ? AND provider_id = ?", userID, providerID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check identity: %w", err)
	}
	return count > 0, nil
}

// LinkIdentity attaches an external identity to an existing user.
func (s *UserService) LinkIdentity(userID, providerID uint, subject, email string) error {
	now := time.Now()
	identity := tables.UserIdentity{
		UserID:      userID,
		ProviderID:  providerID,
		Subject:     subject,
		Email:       email,
		LastLoginAt: &now,
	}
	if err := s.db.Create(&identity).Error; err != nil {
		return fmt.Errorf("link identity: %w", err)
	}
	return nil
}

// TouchIdentity records a successful login through an identity. Failure is
// logged by the caller's context only; it must not block sign-in.
func (s *UserService) TouchIdentity(identity tables.UserIdentity, email string) error {
	updates := map[string]interface{}{"last_login_at": time.Now()}
	if email != "" {
		updates["email"] = email
	}
	return s.db.Model(&tables.UserIdentity{}).Where("id = ?", identity.ID).Updates(updates).Error
}

// CreateExternalUser creates a non-admin, password-less user together with its
// first identity in one transaction. The username is derived from the
// identity claims and de-duplicated with a numeric suffix.
func (s *UserService) CreateExternalUser(in ExternalUserInput) (tables.User, error) {
	var user tables.User

	err := s.db.Transaction(func(tx *gorm.DB) error {
		username, err := uniqueUsername(tx, DeriveUsernameBase(in.PreferredUsername, in.Email))
		if err != nil {
			return err
		}

		fullName := strings.TrimSpace(in.FullName)
		if fullName == "" {
			fullName = username
		}

		user = tables.User{
			Username:  username,
			FullName:  fullName,
			Email:     in.Email,
			IsActive:  true,
			LinkToken: utils.GenerateLinkToken(16),
		}
		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("create external user: %w", err)
		}
		if err := AssignDefaultRole(tx, user.ID); err != nil {
			return fmt.Errorf("assign default role: %w", err)
		}

		now := time.Now()
		identity := tables.UserIdentity{
			UserID:      user.ID,
			ProviderID:  in.ProviderID,
			Subject:     in.Subject,
			Email:       in.Email,
			LastLoginAt: &now,
		}
		if err := tx.Create(&identity).Error; err != nil {
			return fmt.Errorf("create external identity: %w", err)
		}
		return nil
	})
	if err != nil {
		return tables.User{}, err
	}
	return user, nil
}

func uniqueUsername(tx *gorm.DB, base string) (string, error) {
	for i := 0; i < 100; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s%d", base, i+1)
		}
		var count int64
		// Unscoped: the unique constraint also covers soft-deleted rows.
		if err := tx.Unscoped().Model(&tables.User{}).
			Where("LOWER(username) = ?", candidate).
			Count(&count).Error; err != nil {
			return "", fmt.Errorf("check username: %w", err)
		}
		if count == 0 {
			return candidate, nil
		}
	}
	return "", errors.New("could not find a free username")
}

// ListIdentities returns a user's linked accounts.
func (s *UserService) ListIdentities(userID uint) ([]dto.UserIdentityResponse, response.ServerError) {
	var rows []dto.UserIdentityResponse
	if err := s.db.Table("user_identities AS i").
		Select("i.id AS id, i.provider_id AS provider_id, p.slug AS provider_slug, "+
			"p.display_name AS provider_name, i.email AS email, "+
			"i.created_at AS created_at, i.last_login_at AS last_login_at").
		Joins("JOIN oidc_providers p ON p.id = i.provider_id").
		Where("i.user_id = ?", userID).
		Order("i.created_at ASC").
		Scan(&rows).Error; err != nil {
		return nil, response.New(http.StatusInternalServerError, "failed to list linked accounts")
	}
	if rows == nil {
		rows = []dto.UserIdentityResponse{}
	}
	return rows, nil
}

// UnlinkIdentity removes one of the user's identities. It refuses when that
// would leave the user with no way to sign in. Only identities owned by
// userID can be removed.
func (s *UserService) UnlinkIdentity(userID, identityID uint) response.ServerError {
	user, err := s.FindById(userID)
	if err != nil {
		return response.New(http.StatusNotFound, fmt.Sprintf("user %d not found", userID))
	}

	var identity tables.UserIdentity
	if err := s.db.Where("id = ? AND user_id = ?", identityID, userID).First(&identity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.New(http.StatusNotFound, "linked account not found")
		}
		return response.New(http.StatusInternalServerError, "failed to load linked account")
	}

	var others int64
	if err := s.db.Model(&tables.UserIdentity{}).
		Where("user_id = ? AND id <> ?", userID, identityID).
		Count(&others).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to check linked accounts")
	}

	if !CanUnlinkIdentity(user.Password != "", others) {
		return response.New(
			http.StatusConflict,
			"this is the only way to sign in to the account; set a password or link another account first",
		)
	}

	if err := s.db.Delete(&tables.UserIdentity{}, identity.ID).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to unlink account")
	}
	return nil
}

// SetPassword sets a first local password on an account that has none. It is
// refused for accounts that already have one (use ChangePassword) and for
// admin.
func (s *UserService) SetPassword(user tables.User, form dto.SetPasswordForm) response.ServerError {
	if user.Username == "admin" {
		return response.New(
			http.StatusBadRequest,
			"admin password can only be set via PWW_ADMIN_PASSWORD env var",
		)
	}
	if user.Password != "" {
		return response.New(http.StatusBadRequest, "a password is already set; use change password instead")
	}

	newPassword := strings.TrimSpace(form.NewPassword)
	if !s.CheckPasswordComplexity(newPassword) {
		return response.New(
			http.StatusBadRequest,
			"Password must contain at least one letter and one number",
		)
	}

	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return response.New(http.StatusInternalServerError, "Failed to hash password")
	}

	// the password = '' guard makes this a no-op if a concurrent request won
	res := s.db.Model(&tables.User{}).
		Where("id = ? AND COALESCE(password, '') = ''", user.ID).
		Update("password", hashed)
	if res.Error != nil {
		return response.New(http.StatusInternalServerError, "Failed to update db password")
	}
	if res.RowsAffected == 0 {
		return response.New(http.StatusConflict, "a password is already set")
	}
	return nil
}
