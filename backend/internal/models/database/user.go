package models

import (
	"database/sql"
	"time"
)

// --- Core Entities ---

// Maps to the `users` table
type User struct {
	UserID       string         `db:"user_id"`
	Username     string         `db:"username"`
	Email        sql.NullString `db:"email"`
	DisplayName  sql.NullString `db:"display_name"`
	AvatarURL    sql.NullString `db:"avatar_url"`
	AvatarSource sql.NullString `db:"avatar_source"`
	StatsPrivacy string         `db:"stats_privacy"`
	UITheme      sql.NullString `db:"ui_theme"`
	ColorTheme   sql.NullString `db:"color_theme"`
	Role         string         `db:"role"`
	IsBanned     bool           `db:"is_banned"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	LastLoginAt  sql.NullTime   `db:"last_login_at"`
}

// --- Composite & Helper Structs ---

// Slimmed-down user struct for returning search results
type UserSearchResult struct {
	UserID      string         `db:"user_id"`
	Username    string         `db:"username"`
	DisplayName sql.NullString `db:"display_name"`
	AvatarURL   sql.NullString `db:"avatar_url"`
}

// --- Data Transfer Structs ---

// Defines the set of parameters that can be used to update a user settings
type UpdateUserParams struct {
	DisplayName  *string
	AvatarURL    *string
	AvatarSource *string
	StatsPrivacy *string
	UITheme      *string
	ColorTheme   *string
}
