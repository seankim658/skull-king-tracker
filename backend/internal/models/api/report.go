package models

import "time"

// Represents a single user report in an API response
type UserReportResponse struct {
	ReportID       string    `json:"report_id"`
	ReporterUserID string    `json:"reporter_user_id"`
	ReporterName   string    `json:"reporter_name"`
	ReportedUserID string    `json:"reported_user_id"`
	ReportedName   string    `json:"reported_name"`
	Reason         string    `json:"reason"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// Paginated response for a list of user reports
type PaginatedReportsResponse struct {
	Reports    []UserReportResponse `json:"reports"`
	Pagination Pagination           `json:"pagination"`
}
