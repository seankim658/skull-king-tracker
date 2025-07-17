package handlers

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	i "github.com/seankim658/skullking/internal/images"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
	dbModels "github.com/seankim658/skullking/internal/models/database"
	"github.com/seankim658/skullking/internal/sse"
)

type AdminHandler struct {
	Cfg    *cf.Config
	SSEHub *sse.Hub
}

const adminComponent = "handlers-admin"

func NewAdminHandler(cfg *cf.Config, sseHub *sse.Hub) *AdminHandler {
	return &AdminHandler{Cfg: cfg, SSEHub: sseHub}
}

// Retrieves a paginated and filtered list of user reports
// Path: /admin/reports
// Method: GET
func (ah *AdminHandler) HandleGetReports(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		adminComponent,
		"HandleGetReports",
	)

	page, pageSize := GetPaginationParams(r)
	filters := db.ReportFilterOptions{
		Status:     QueryParam(r, "status"),
		ReporterID: QueryParam(r, "reporter_id"),
		ReportedID: QueryParam(r, "reported_id"),
		SortBy:     QueryParam(r, "sort_by"),
		SortOrder:  QueryParam(r, "sort_order"),
		Page:       page,
		PageSize:   pageSize,
	}

	totalCount, err := db.CountReports(ctx, nil, filters)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to count reports")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to count reports")
		return
	}

	dbReports, err := db.GetPaginatedReports(ctx, nil, filters)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve reports")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve reports")
		return
	}

	apiReports := make([]apiModels.UserReportResponse, len(dbReports))
	for i, rpt := range dbReports {
		apiReports[i] = apiModels.UserReportResponse{
			ReportID:       rpt.ReportID,
			ReporterUserID: rpt.ReporterUserID,
			ReporterName:   rpt.ReporterName,
			ReportedUserID: rpt.ReportedUserID,
			ReportedName:   rpt.ReportedName,
			Reason:         rpt.Reason,
			Status:         rpt.Status,
			CreatedAt:      rpt.CreatedAt,
		}
	}

	pagination := CalculatePagination(totalCount, page, pageSize)
	response := apiModels.PaginatedReportsResponse{
		Reports:    apiReports,
		Pagination: pagination,
	}

	Respond(w, r, http.StatusOK, response, "Reports retrieved successfully")
}

// Updates the status of a specific user report.
// Path: /admin/reports/{report_id}
// Method: PUT
func (ah *AdminHandler) HandleUpdateReportStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(l.GetLoggerFromContext(ctx), adminComponent, "HandleUpdateReportStatus")

	reportID, ok := PathVar(w, r, "report_id")
	if !ok {
		return
	}

	var req apiModels.UpdateReportRequest
	if !ParseJSON(w, r, &req) {
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to update report status")
	if !txOk {
		return
	}
	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	opErr = db.UpdateReportStatus(ctx, tx, reportID, req.Status)
	if opErr != nil {
		if errors.Is(opErr, db.ErrReportNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Report not found")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to update report status")
		}
		responseSent = true
		return
	}

	if !responseSent {
		Respond(w, r, http.StatusOK, nil, "Report status updated successfully")
		responseSent = true
	}
}

// Handles banning a user account
// Path: /admin/users/{user_id}/ban
// Method: POST
func (ah *AdminHandler) HandleBanUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		adminComponent,
		"HandleBanUser",
	)

	userToBanID, ok := PathVar(w, r, "user_id")
	if !ok {
		return
	}

	var req apiModels.BanUserRequest
	if !ParseJSON(w, r, &req) {
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to ban user")
	if !txOk {
		return
	}
	defer tx.Rollback()

	userToBan, err := db.GetUserByID(ctx, tx, userToBanID)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "User to ban not found")
		} else {
			logger.Error().Err(err).Msg("Failed to fetch user for ban operation")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to ban user")
		}
		return
	}

	if err := db.SetUserBanStatus(ctx, tx, userToBanID, true, req.Reason); err != nil {
		logger.Error().Err(err).Str(l.UserIDKey, userToBanID).Msg("Failed to ban user")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to ban user")
		return
	}

	var localAvatarFilename string
	if userToBan.AvatarURL.Valid && strings.HasPrefix(userToBan.AvatarURL.String, i.AvatarWebPrefixPath) {
		localAvatarFilename = filepath.Base(userToBan.AvatarURL.String)
		emptyString := ""
		params := dbModels.UpdateUserParams{
			AvatarURL:    &emptyString,
			AvatarSource: &emptyString,
		}
		if err := db.UpdateUserProfile(ctx, tx, userToBanID, params); err != nil {
			logger.Error().Err(err).Msg("Failed to clear avatar fields from DB during ban operation")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to update user record during ban")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error().Err(err).Msg("Failed to commit transaction for user ban")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to finalize user ban")
		return
	}

	if localAvatarFilename != "" {
		if err := i.DeleteLocalAvatar(ah.Cfg.AvatarStoragePath, localAvatarFilename); err != nil {
			logger.Warn().Err(err).Msg("User banned and DB updated, but failed to delete avatar file from disk")
		} else {
			logger.Info().Str(l.FileKey, localAvatarFilename).Msg("Successfully deleted banned user's avatar file")
		}
	}

	Respond(w, r, http.StatusOK, nil, "User successfully banned")
}

// Sets a user's is_banned flag to false
// Path: /admin/users/{user_id}/unban
// Method: POST
func (ah *AdminHandler) HandleUnbanUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(l.GetLoggerFromContext(ctx), adminComponent, "HandleUnbanUser")

	userToUnbanID, ok := PathVar(w, r, "user_id")
	if !ok {
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to unban user")
	if !txOk {
		return
	}
	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	opErr = db.SetUserBanStatus(ctx, tx, userToUnbanID, false, "")
	if opErr != nil {
		if errors.Is(opErr, db.ErrUserNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "User to unban not found")
		} else {
			logger.Error().Err(opErr).Str(l.UserIDKey, userToUnbanID).Msg("Failed to unban user")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to unban user")
		}
		responseSent = true
		return
	}

	if !responseSent {
		Respond(w, r, http.StatusOK, nil, "User successfully unbanned.")
		responseSent = true
	}
}

// HandleSendAdminNotification sends a notification to a specific user or all users.
// Path: /admin/notifications
// Method: POST
func (ah *AdminHandler) HandleSendAdminNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(l.GetLoggerFromContext(ctx), adminComponent, "HandleSendAdminNotification")

	adminUserID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}

	var req apiModels.SendNotificationRequest
	if !ParseJSON(w, r, &req) {
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to send notification")
	if !txOk {
		return
	}
	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	sendAndBroadcast := func(recipientID string, notificationType string) {
		notificationID, err := db.CreateNotification(ctx, tx, recipientID, adminUserID, notificationType, req.Message, nil)
		if err != nil {
			logger.Error().Err(err).Str(l.RecipientIDKey, recipientID).Msg("Failed to create notification for user")
			return
		}

		fullNotification, err := db.GetNotificationWithActorByID(ctx, tx, notificationID)
		if err != nil {
			logger.Error().Err(err).Str(l.NotificationIDKey, notificationID).Msg("Failed to fetch created notification for SSE broadcast")
		}

		go broadcastNotificationEvent(ah.SSEHub, fullNotification, logger)
	}

	if req.IsBroadcast {
		// Broadcast to all users
		allUserIDs, err := db.GetAllUserIDs(ctx, tx)
		if err != nil {
			opErr = err
			ErrorResponse(w, r, http.StatusInternalServerError, "Could not retrieve users for broadcast")
			responseSent = true
			return
		}

		for _, recipientID := range allUserIDs {
			// You might want to skip sending a notification to the admin who initiated it
			if recipientID == adminUserID {
				continue
			}
			sendAndBroadcast(recipientID, "admin_broadcast")
		}
	} else if len(req.UserIDs) > 0 {
		// Send to a list of specific users
		for _, recipientID := range req.UserIDs {
			_, err := db.CreateNotification(ctx, tx, recipientID, adminUserID, "admin_message", req.Message, nil)
			if err != nil {
				logger.Error().Err(err).Str(l.RecipientIDKey, recipientID).Msg("Failed to create targeted notification for users")
			}
		}
	} else {
		ErrorResponse(w, r, http.StatusBadRequest, "Request must be a broadcast or include a user_id")
		responseSent = true
		return
	}

	if !responseSent {
		Respond(w, r, http.StatusAccepted, nil, "Notification sent successfully")
		responseSent = true
	}
}
