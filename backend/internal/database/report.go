package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	l "github.com/seankim658/skullking/internal/logger"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

const reportComponent = "database-report"

// Creates a new user report in the database
func CreateReport(ctx context.Context, tx *sql.Tx, reporterID, reportedID, reason string) (string, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		reportComponent,
		"CreateReport",
	).With().Str(l.ReporterIDKey, reporterID).Str(l.ReportedIDKey, reportedID).Str(l.ReportReasonKey, reason).Logger()

	newReportID := uuid.NewString()
	query := `
  INSERT INTO user_reports (report_id, reporter_user_id, reported_user_id, reason)
  VALUES ($1, $2, $3, $4)
  RETURNING report_id;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to create user report")

	var returnedID string
	err := querier.QueryRowContext(ctx, query, newReportID, reporterID, reportedID, reason).Scan(&reportedID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create user report")
		return "", fmt.Errorf("error creating report: %w", err)
	}

	logger.Info().Str(l.ReportIDKey, returnedID).Msg("User report created successfully")
	return returnedID, nil
}

type ReportFilterOptions struct {
	Status     string
	ReporterID string
	ReportedID string
	SortBy     string
	SortOrder  string
	Page       int
	PageSize   int
}

// Counts user reports based on the provided filters
func CountReports(ctx context.Context, tx *sql.Tx, options ReportFilterOptions) (int64, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		reportComponent,
		"CountReports",
	)

	var args []any
	argCounter := 1
	var whereClauses []string

	if options.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("ur.status = $%d", argCounter))
		args = append(args, options.Status)
		argCounter++
	}
	if options.ReporterID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("ur.reporter_user_id = $%d", argCounter))
		args = append(args, options.ReporterID)
		argCounter++
	}
	if options.ReportedID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("ur.reported_user_id = $%d", argCounter))
		args = append(args, options.ReportedID)
		argCounter++
	}

	query := "SELECT COUNT(*) FROM user_reports ur"
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var totalCount int64
	err := querier.QueryRowContext(ctx, query, args...).Scan(&totalCount)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to count user reports")
		return 0, fmt.Errorf("error counting reports: %w", err)
	}
	return totalCount, nil
}

// Retrieves a paginated and filtered list of user reports
func GetPaginatedReports(ctx context.Context, tx *sql.Tx, options ReportFilterOptions) ([]dbModels.UserReportWithNames, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		reportComponent,
		"GetPaginatedReports",
	).With().Interface(l.ArgsKey, options).Logger()

	var args []any
	argCounter := 1
	var whereClauses []string

	if options.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("ur.status = $%d", argCounter))
		args = append(args, options.Status)
		argCounter++
	}
	if options.ReporterID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("ur.reporter_user_id = $%d", argCounter))
		args = append(args, options.ReporterID)
		argCounter++
	}
	if options.ReportedID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("ur.reported_user_id = $%d", argCounter))
		args = append(args, options.ReportedID)
		argCounter++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	sortColumnMap := map[string]string{
		"created_at": "ur.created_at",
		"updated_at": "ur.updated_at",
		"status":     "ur.status",
	}
	orderByClause := "ORDER BY ur.created_at DESC"
	if validColumn, ok := sortColumnMap[options.SortBy]; ok {
		orderDirection := "ASC"
		if strings.ToUpper(options.SortOrder) == "DESC" {
			orderDirection = "DESC"
		}
		orderByClause = fmt.Sprintf("ORDER BY %s %s, ur.created_at DESC", validColumn, orderDirection)
	}

	offset := (options.Page - 1) * options.PageSize
	args = append(args, options.PageSize, offset)

	query := fmt.Sprintf(`
    SELECT
      ur.report_id, ur.reporter_user_id, ur.reported_user_id, ur.reason, ur.status, ur.created_at, ur.updated_at,
      COALESCE(reporter.display_name, reporter.username) AS reporter_name,
      COALESCE(reported.display_name, reported.username) AS reported_name
    FROM user_reports ur
    JOIN users AS reporter ON ur.reporter_user_id = reporter.user_id
    JOIN users AS reported ON ur.reported_user_id = reported.user_id
    %s
    %s
    LIMIT $%d OFFSET $%d;
  `, whereClause, orderByClause, argCounter, argCounter+1)
	logger.Debug().Str(l.QueryKey, query).Interface(l.ArgsKey, args).Msg("Executing paginated reports query")

	rows, err := querier.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query paginated reports")
		return nil, fmt.Errorf("error querying reports: %w", err)
	}
	defer rows.Close()

	var reports []dbModels.UserReportWithNames
	for rows.Next() {
		var r dbModels.UserReportWithNames
		err := rows.Scan(
			&r.ReportID, &r.ReporterUserID, &r.ReportedUserID, &r.Reason, &r.Status, &r.CreatedAt, &r.UpdatedAt,
			&r.ReporterName, &r.ReportedName,
		)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to scan paginated report row")
			return nil, fmt.Errorf("error scanning report: %w", err)
		}
		reports = append(reports, r)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating over paginated report rows")
		return nil, fmt.Errorf("error iterating report rows: %w", err)
	}

	return reports, nil
}

// Updates the status of a specific report
func UpdateReportStatus(ctx context.Context, tx *sql.Tx, reportID, newStatus string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		reportComponent,
		"UpdateReportStatus",
	).With().Str(l.ReportIDKey, reportID).Str(l.StatusKey, newStatus).Logger()

	query := "UPDATE user_reports SET status = $1, updated_at = NOW() WHERE report_id = $2;"
	result, err := querier.ExecContext(ctx, query, newStatus, reportID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update report status")
		return fmt.Errorf("error updating report status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrReportNotFound
	}

	logger.Info().Str(l.ReportIDKey, reportID).Str(l.StatusKey, newStatus).Msg("Report status updated")
	return nil
}
