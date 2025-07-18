package models

import "time"

// --- Request Payloads ---

// Payload for creating or updating a site alert
type SiteAlertRequest struct {
	Title     string    `json:"title" validate:"required"`
	Body      string    `json:"body" validate:"required"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
	IsActive  bool      `json:"is_active"`
}

// --- Response Payloads ---

// Represents a single site alert in an API response
type SiteAlertResponse struct {
	AlertID     string    `json:"alert_id"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsActive    bool      `json:"is_active"`
	CreatorName string    `json:"creator_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Paginated response for a list of site alerts
type PaginatedAlertsResponse struct {
	Alerts     []SiteAlertResponse `json:"alerts"`
	Pagination Pagination          `json:"pagination"`
}
