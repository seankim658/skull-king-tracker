package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
	"github.com/rs/zerolog"

	a "github.com/seankim658/skullking/internal/auth"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
	modelConverters "github.com/seankim658/skullking/internal/models/convert"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

const utilComponent = "handlers-utils"
const defaultPage = 1
const defaultPageSize = 25

// Writes a JSON response with the given status code and data
func Respond(w http.ResponseWriter, r *http.Request, status int, data any, message string) {
	requestLogger := l.GetLoggerFromContext(r.Context())
	response := apiModels.APIResponse{
		Success: status >= 200 && status < 300,
		Data:    data,
		Message: message,
	}
	if !response.Success {
		if errData, ok := data.(string); ok {
			response.Error = errData
			response.Data = nil
		} else if errData, ok := data.(error); ok {
			response.Error = errData.Error()
			response.Data = nil
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		requestLogger.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

// Writes a standardized JSON error response
func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message string) {
	Respond(w, r, status, nil, message)
}

// Decodes the JSON request body
func ParseJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		ErrorResponse(w, r, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return false
	}
	return true
}

// Ensures all required fields are present
func RequireFields(w http.ResponseWriter, r *http.Request, fields map[string]string) bool {
	for field, value := range fields {
		if value == "" {
			ErrorResponse(w, r, http.StatusBadRequest, field+" is required")
			return false
		}
	}
	return true
}

// Gets a query paramter value from the request URL
func QueryParam(r *http.Request, param string) string {
	return r.URL.Query().Get(param)
}

// Gets a query paramter and converts it to an int
func QueryParamInt(r *http.Request, param string) (int, bool) {
	strValue := QueryParam(r, param)
	if strValue == "" {
		return 0, false
	}
	value, err := strconv.Atoi(strValue)
	if err != nil {
		return 0, false
	}
	return value, true
}

// Gets a query paramter and converts it to an int64
func QueryParamInt64(r *http.Request, param string) (int64, bool) {
	strValue := QueryParam(r, param)
	if strValue == "" {
		return 0, false
	}
	value, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		return 0, false
	}
	return value, true
}

// Parses page and page size query parameters with defaults
func GetPaginationParams(r *http.Request) (page, pageSize int) {
	pageStr := QueryParam(r, "page")
	pageSizeStr := QueryParam(r, "page_size")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = defaultPage
	}

	pageSize, err = strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = defaultPageSize
	}

	return page, pageSize
}

// Creates the pagination struct
func CalculatePagination(totalCount int64, page, pageSize int) apiModels.Pagination {
	totalPages := int64(0)
	if totalCount > 0 && pageSize > 0 {
		totalPages = int64(math.Ceil(float64(totalCount) / float64(pageSize)))
	}
	if page > int(totalPages) && totalPages > 0 {
		page = int(totalPages)
	}
	if page < 1 {
		page = 1
	}

	return apiModels.Pagination{
		CurrentPage: page,
		TotalPages:  totalPages,
		PageSize:    pageSize,
		TotalCount:  totalCount,
	}
}

// Start a database transaction or send the error response
func StartTx(ctx context.Context, w http.ResponseWriter, r *http.Request, logger zerolog.Logger, errorMessage string) (*sql.Tx, bool) {
	tx, txErr := db.DB.BeginTx(ctx, nil)
	if txErr != nil {
		logger.Error().Err(txErr).Msg("Database transaction could not be started")
		ErrorResponse(w, r, http.StatusInternalServerError, errorMessage)
		return nil, false
	}
	return tx, true
}

// Get the session cookie
func GetSessionStore(w http.ResponseWriter, r *http.Request, failureMessage string, httpStatus int, logger zerolog.Logger) (*sessions.Session, error) {
	session, err := gothic.Store.Get(r, a.SessionCookieName)
	if err != nil {
		logger.Warn().Err(err).Str("failure_message", failureMessage).Msg("Failed to get session store")
		ErrorResponse(w, r, httpStatus, failureMessage)
		return nil, err
	}
	return session, nil
}

// Extracts the user ID from the session
func GetAuthenticatedUserIDFromSession(w http.ResponseWriter, r *http.Request, logger zerolog.Logger) (string, bool) {
	session, err := GetSessionStore(w, r, "Not authenticated: session error", http.StatusUnauthorized, logger)
	if err != nil {
		return "", false
	}

	userIDVal, ok := session.Values[a.UserIDSessionKey]
	if !ok || userIDVal == nil {
		logger.Warn().Msg("No user_id found in session or user_id is nil")
		ErrorResponse(w, r, http.StatusUnauthorized, "Not authenticated: no user ID in session")
		return "", false
	}

	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		logger.Warn().Str(l.UserIDKey, fmt.Sprintf("%v", userIDVal)).Msg("user_id in session is not a string or is empty")
		ErrorResponse(w, r, http.StatusUnauthorized, "Not authenticated: invalid user ID in session")
		return "", false
	}

	return userID, true
}

// Fetches user details, converts to an API model, and sends an API response
func FetchUserAndRespond(w http.ResponseWriter, r *http.Request, tx *sql.Tx, userID string, logger zerolog.Logger, successStatus int, successMessage string) {
	ctx := r.Context()
	dbUser, dbErr := db.GetUserByID(ctx, tx, userID)
	if dbErr != nil {
		if errors.Is(dbErr, db.ErrUserNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "User not found")
		} else {
			logger.Error().Err(dbErr).Msg("Failed to fetch user by ID")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve user details")
		}
		return
	}

	apiUser, convErr := modelConverters.DBUserToAPIUser(dbUser)
	if convErr != nil {
		logger.Error().Err(convErr).Msg("Failed to convert DB user to API user for response")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to process user details")
		return
	}

	Respond(w, r, successStatus, apiModels.AuthenticatedUserResponse{User: *apiUser}, successMessage)
}

// Verifies a game exists and if the authenticated user is the current scorekeeper
func CheckGameAccessAndScorekeeper(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	gameID string,
	userID string,
	logger zerolog.Logger,
) (*dbModels.Game, bool) {
	game, err := db.GetGameByID(ctx, nil, gameID)
	if err != nil {
		if errors.Is(err, db.ErrGameNotFound) {
			logger.Warn().Err(err).Str(l.GameIDKey, gameID).Msg("Game not found for authentication check")
			ErrorResponse(w, r, http.StatusNotFound, "Game not found")
		} else {
			logger.Error().Err(err).Str(l.GameIDKey, gameID).Msg("Failed to fetch game for authorization check")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to verfiy game access")
		}
		return nil, false
	}

	if !game.CurrentScorekeeperUserID.Valid || game.CurrentScorekeeperUserID.String != userID {
		logger.Warn().
			Str(l.GameIDKey, gameID).
			Str(l.UserIDKey, userID).
			Str(l.ScorekeeperIDKey, game.CurrentScorekeeperUserID.String).
			Msg("User is not the scorekeeper of the game")
		ErrorResponse(w, r, http.StatusForbidden, "You are not authorized to modify this game")
		return nil, false
	}

	logger.Debug().Str(l.GameIDKey, gameID).Str(l.UserIDKey, userID).Msg("User confirmed as scorekeeper")
	return game, true
}
