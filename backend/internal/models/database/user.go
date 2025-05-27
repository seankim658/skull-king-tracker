package models

import (
	"database/sql"
	"time"
)

// Databse user model
type User struct {
	UserID       string         `db:"user_id"`
	Username     string         `db:"username"`
	Email        sql.NullString `db:"email"`
	DisplayName  sql.NullString `db:"display_name"`
	AvatarURL    sql.NullString `db:"avatar_url"`
	StatsPrivacy string         `db:"stats_privacy"`
	UITheme      sql.NullString `db:"ui_theme"`
	ColorTheme   sql.NullString `db:"color_theme"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	LastLoginAt  sql.NullTime   `db:"last_login_at"`
}
