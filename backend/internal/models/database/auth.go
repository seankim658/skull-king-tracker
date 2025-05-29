package models

import (
	"database/sql"
	"time"
)

// Maps to the `user_provider_identities` table
type UserProviderIdentity struct {
	ProviderIdentityID  string         `db:"provider_identity_id"`
	UserID              string         `db:"user_id"`
	ProviderName        string         `db:"provider_name"`
	ProviderUserID      string         `db:"provider_user_id"`
	ProviderEmail       sql.NullString `db:"provider_email"`
	ProviderDisplayName sql.NullString `db:"provider_display_name"`
	ProviderAvatarURL   sql.NullString `db:"provider_avatar_url"`
	CreatedAt           time.Time      `db:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at"`
}
