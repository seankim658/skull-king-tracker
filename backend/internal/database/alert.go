package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	l "github.com/seankim658/skullking/internal/logger"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

const alertComponent = "database-alert"

// Creates a new site alert
func CreateSiteAlert(ctx context.Context, tx *sql.Tx, alert dbModels.SiteAlert) (string, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		alertComponent,
		"CreateSiteAlert",
	)

	newAlertID := uuid.NewString()
	query := `
  INSERT INTO site_alerts (
    alert_id, title, body, start_time, end_time, is_active, created_by_user_id
  )
  VALUES ($1, $2, $3, $4, $5, $6, $7)
  RETURNING alert_id;
  `
	logger.Debug().Str(l.QueryKey, query).Interface(l.AlertKey, alert).Msg("Attempting to create alert")

	var returnedID string
	err := querier.QueryRowContext(
		ctx, query, newAlertID, alert.Title,
		alert.Body, alert.StartTime, alert.EndTime,
		alert.IsActive, alert.CreatedByUserID,
	).Scan(&returnedID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create site alert")
		return "", fmt.Errorf("error creating site alert: %w", err)
	}
	return returnedID, nil
}

// Retrieves all currently active and valid alerts
func GetActiveSiteAlerts(ctx context.Context, tx *sql.Tx) ([]dbModels.SiteAlert, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		alertComponent,
		"GetActiveSiteAlerts",
	)

	query := `
  SELECT alert_id, title, body, start_time, end_time, is_active, created_by_user_id, created_at, updated_at
  FROM site_alerts
  WHERE is_active = TRUE AND NOW() BETWEEN start_time AND end_time
  ORDER BY created_at DESC;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get all active alerts")

	rows, err := querier.QueryContext(ctx, query)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query active site alerts")
		return nil, fmt.Errorf("error querying active alerts: %w", err)
	}
	defer rows.Close()

	var alerts []dbModels.SiteAlert
	for rows.Next() {
		var a dbModels.SiteAlert
		if err := rows.Scan(
			&a.AlertID, &a.Title, &a.Body, &a.StartTime,
			&a.EndTime, &a.IsActive, &a.CreatedByUserID,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan active alert row")
			return nil, fmt.Errorf("error scanning active alert: %w", err)
		}
		alerts = append(alerts, a)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating over active site alerts")
		return nil, fmt.Errorf("error iterating active site alerts: %w", err)
	}
	return alerts, nil
}

// Retrieves a paginated list of all alerts for the admin panel
func GetPaginatedSiteAlerts(ctx context.Context, tx *sql.Tx, page, pageSize int) ([]dbModels.SiteAlertWithCreator, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		alertComponent,
		"GetPaginatedSiteAlerts",
	).With().Int(l.PageKey, page).Int(l.PageSizeKey, pageSize).Logger()

	offset := (page - 1) * pageSize
	query := `
  SELECT
    sa.alert_id, sa.title, sa.body, 
    sa.start_time, sa.end_time, sa.is_active, 
    sa.created_by_user_id, sa.created_at, sa.updated_at,
  COALESCE(u.display_name, u.username) AS creator_name
  FROM site_alerts sa
  JOIN users u ON sa.created_by_user_id = u.user_id
  ORDER BY sa.start_time DESC, sa.created_at DESC
  LIMIT $1 OFFSET $2;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get paginated site alerts")

	rows, err := querier.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query paginated site alerts")
		return nil, fmt.Errorf("error querying site alerts: %w", err)
	}
	defer rows.Close()

	var alerts []dbModels.SiteAlertWithCreator
	for rows.Next() {
		var a dbModels.SiteAlertWithCreator
		if err := rows.Scan(
			&a.AlertID, &a.Title, &a.Body, &a.StartTime,
			&a.EndTime, &a.IsActive, &a.CreatedByUserID,
			&a.CreatedAt, &a.UpdatedAt, &a.CreatorName,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan paginated alert row")
			return nil, fmt.Errorf("error scanning site alert: %w", err)
		}
		alerts = append(alerts, a)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating over paginated alert rows")
		return nil, fmt.Errorf("error iterating paginated alert rows: %w", err)
	}
	return alerts, nil
}

// Counts all site alerts
func CountSiteAlerts(ctx context.Context, tx *sql.Tx) (int64, error) {
	querier := GetQuerier(tx)
	var totalCount int64
	err := querier.QueryRowContext(ctx, "SELECT COUNT(*) FROM site_alerts;").Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("error counting site alerts: %w", err)
	}
	return totalCount, nil
}

// Updates an existing site alert
func UpdateSiteAlert(ctx context.Context, tx *sql.Tx, alertID string, alert dbModels.SiteAlert) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		alertComponent,
		"UpdateSiteAlert",
	).With().Interface(l.AlertKey, alert).Logger()

	query := `
  UPDATE site_alerts
  SET title = $1, body = $2, start_time = $3, end_time = $4, is_active = $5, updated_at = NOW()
  WHERE alert_id = $6;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to update alert")

	result, err := querier.ExecContext(
		ctx, query, alert.Title, alert.Body,
		alert.StartTime, alert.EndTime, alert.IsActive, alertID,
	)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update site alert")
		return fmt.Errorf("error updating site alert: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrAlertNotFound
	}

	return nil
}

// Deletes a site alert by ID
func DeleteSiteAlert(ctx context.Context, tx *sql.Tx, alertID string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		alertComponent,
		"DeleteSiteAlert",
	).With().Str(l.AlertIDKey, alertID).Logger()

	query := "DELTE FROM site_alerts WHERE alert_id = $1;"
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to delete site alert")

	result, err := querier.ExecContext(ctx, query, alertID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to delete site alert")
		return fmt.Errorf("error deleting site alert: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrReportNotFound
	}

	return nil
}
