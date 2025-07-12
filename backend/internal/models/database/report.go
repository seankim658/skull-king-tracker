package models

import "time"

// Maps to the `user_reports` table
type UserReport struct {
	ReportID       string    `db:"report_id"`
	ReporterUserID string    `db:"reporter_user_id"`
	ReportedUserID string    `db:"reported_user_id"`
	Reason         string    `db:"reason"`
	Status         string    `db:"status"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// Composite struct for queries that include user display names
type UserReportWithNames struct {
	UserReport
	ReporterName string `db:"reporter_name"`
	ReportedName string `db:"reported_name"`
}
