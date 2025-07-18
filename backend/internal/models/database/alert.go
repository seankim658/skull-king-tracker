package models

import "time"

// Maps to the `site_alerts` table
type SiteAlert struct {
	AlertID         string    `db:"alert_id"`
	Title           string    `db:"title"`
	Body            string    `db:"body"`
	StartTime       time.Time `db:"start_time"`
	EndTime         time.Time `db:"end_time"`
	IsActive        bool      `db:"is_active"`
	CreatedByUserID string    `db:"created_by_user_id"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type SiteAlertWithCreator struct {
	SiteAlert
	CreatorName string `db:"creator_name"`
}
