package handlers

import (
	"context"
	"errors"
	"math"
	"net/http"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
	modelConverters "github.com/seankim658/skullking/internal/models/convert"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

const userComponent = "handers-user"

type UserProfileHandler struct {
	Cfg *cf.Config
}

func NewUserProfileHandler(cfg *cf.Config) *UserProfileHandler {
	return &UserProfileHandler{Cfg: cfg}
}

// Retrieves and serves a user's profile
// Path: /users/{userID}/profile
// Method: GET
func (uph *UserProfileHandler) HandleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"HandleGetUserProfile",
	)

	profileUserIDFromPath, ok := PathVar(w, r, "user_id")
	if !ok {
		return
	}
	logger = logger.With().Str("profile_user_id_to_view", profileUserIDFromPath).Logger()

	viewerUserID, isAuthenticated := GetAuthenticatedUserIDFromSession(w, r, logger)
	if isAuthenticated {
		logger = logger.With().Str("viewer_user_id", viewerUserID).Logger()
	} else {
		logger.Info().Msg("Viewer is not authenticated for this profile request")
	}

	// 1. Fetch Profile User's Basic Data
	profileDBUser, dbErr := db.GetUserByID(ctx, nil, profileUserIDFromPath)
	if dbErr != nil {
		if errors.Is(dbErr, db.ErrUserNotFound) {
			logger.Warn().Msg("Profile user not found in database")
			ErrorResponse(w, r, http.StatusNotFound, "User profile not found")
		} else {
			logger.Error().Err(dbErr).Msg("Database error fetching profile user by ID")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve user profile details")
		}
		return
	}

	apiProfileUserPart, convErr := modelConverters.DBUserToAPIUser(profileDBUser)
	if convErr != nil {
		logger.Error().Err(convErr).Msg("Failed to convert profile DB user to API user model")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to process user profile data")
		return
	}

	// 2. Fetch Friend Count
	friendCount, fcErr := db.CountFriends(ctx, nil, profileUserIDFromPath)
	if fcErr != nil {
		logger.Error().Err(fcErr).Msg("Database error counting friends for profile user, proceeding with count 0")
		friendCount = 0
	}

	// 3. Determine Friendship Status
	var apiFriendshipStatus apiModels.FriendshipStatus
	if !isAuthenticated {
		apiFriendshipStatus = apiModels.FriendshipStatusAPIViewerNotAuth
	} else if viewerUserID == profileUserIDFromPath {
		apiFriendshipStatus = apiModels.FriendshipStatusAPISelf
	} else {
		dbStatusBetweenUsers, fsErr := db.GetFriendshipStatus(ctx, nil, viewerUserID, profileUserIDFromPath)
		if fsErr != nil {
			logger.Error().Err(fsErr).Msg("Database error getting friendship status, defaulting to unknown")
			apiFriendshipStatus = apiModels.FriendshipStatusAPIUnknown
		} else {
			apiFriendshipStatus = modelConverters.DBFriendshipStatusToAPIStatus(dbStatusBetweenUsers)
		}
	}
	logger.Debug().Str(l.FriendshipStatusKey, string(apiFriendshipStatus)).Msg("Determined API friendship status")

	apiProfile := apiModels.UserProfile{
		UserID:           apiProfileUserPart.UserID,
		Username:         apiProfileUserPart.Username,
		DisplayName:      apiProfileUserPart.DisplayName,
		AvatarURL:        apiProfileUserPart.AvatarURL,
		StatsPrivacy:     profileDBUser.StatsPrivacy,
		CreatedAt:        apiProfileUserPart.CreatedAt,
		FriendCount:      friendCount,
		FriendshipStatus: apiFriendshipStatus,
	}
	if isAuthenticated && viewerUserID != profileUserIDFromPath {
		mutualCount, mfErr := db.CountMutualFriends(ctx, nil, viewerUserID, profileUserIDFromPath)
		if mfErr != nil {
			logger.Error().Err(mfErr).Msg("Database error counting mutual friends, proceeding without it")
		} else {
			apiProfile.MutualFriendCount = &mutualCount
		}
	}
	finalResponse := apiModels.UserProfileResponse{
		Profile: apiProfile,
	}

	// 4. Fetch Stats if Permitted
	canViewStats := false
	switch profileDBUser.StatsPrivacy {
	case "public":
		canViewStats = true
	case "friends_only":
		if apiFriendshipStatus == apiModels.FriendshipStatusAPIFriends ||
			apiFriendshipStatus == apiModels.FriendshipStatusAPISelf {
			canViewStats = true
		}
	case "private":
		if apiFriendshipStatus == apiModels.FriendshipStatusAPISelf {
			canViewStats = true
		}
	}

	if canViewStats {
		logger.Debug().Msg("Viewer has permission to see stats for this profile")
		dbUserStats, statsErr := db.GetUserBasicStats(ctx, nil, profileUserIDFromPath)
		if statsErr != nil {
			logger.Error().Err(statsErr).Msg("Database error fetching basic stats, stats will be omitted")
		} else if dbUserStats != nil {
			var winPercentage float64
			if dbUserStats.TotalGamesPlayed > 0 {
				winPercentage = math.Round(
					(float64(dbUserStats.TotalWins)/float64(dbUserStats.TotalGamesPlayed))*10000) / 100
			}
			finalResponse.Stats = &apiModels.UserStats{
				TotalGamesPlayed: dbUserStats.TotalGamesPlayed,
				TotalWins:        dbUserStats.TotalWins,
				Top3Finishes:     dbUserStats.Top3Finishes,
				WinPercentage:    winPercentage,
			}
			logger.Debug().Interface("stats_data_for_api", finalResponse.Stats).Msg("Stats data prepared")
		}
	} else {
		logger.Debug().
			Str(l.StatsPrivacyKey, profileDBUser.StatsPrivacy).
			Str("friendship_with_viewer", string(apiFriendshipStatus)).
			Msg("Viewer does not have permission to see stats for this profile")
	}

	Respond(w, r, http.StatusOK, finalResponse, "User profile retrieved successfully")
}

// Handles requests to search for users
// Path: /users/search
// Method: GET
func (uph *UserProfileHandler) HandleSearchUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"HandleSearchUsers",
	)

	searchQuery := QueryParam(r, "q")
	if searchQuery == "" {
		ErrorResponse(w, r, http.StatusBadRequest, "Search query 'q' is required")
		return
	}

	limit, ok := QueryParamInt(r, "limit")
	if !ok {
		limit = 10
	}
	logger = logger.With().Str(l.SearchQueryKey, searchQuery).Int(l.LimitKey, limit).Logger()

	dbUsers, err := db.SearchUsers(ctx, nil, searchQuery, limit)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to search users in database")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to search for users")
		return
	}

	apiUsers := make(apiModels.UserSearchResponse, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		apiUser, convErr := modelConverters.DBUserSearchResultToAPISearchItem(&dbUser)
		if convErr != nil {
			logger.Error().Err(convErr).Msg("Failed to convert DB user search result to API model")
			continue
		}
		apiUsers = append(apiUsers, *apiUser)
	}

	Respond(w, r, http.StatusOK, apiUsers, "User search completed successfully")
}

// Retrieves a user's friends list with friendship status relative to the viewer
// Path: /users/{user_id}/friends
// Method: GET
func (uph *UserProfileHandler) HandleGetFriendsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"HandleGetFriendsList",
	)

	profileUserID, ok := PathVar(w, r, "user_id")
	if !ok {
		return
	}
	page, pageSize := GetPaginationParams(r)
	viewerUserID, isAuthenticated := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !isAuthenticated {
		viewerUserID = ""
	}

	logger = logger.With().Str(l.ProfileUserIDKey, profileUserID).Str(l.ViewerUserIDKey, viewerUserID).Logger()

	totalCount, err := db.CountFriends(ctx, nil, profileUserID)
	if err != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to count friends")
		return
	}

	dbFriends, err := db.GetFriendshipWithViewerStatus(ctx, nil, profileUserID, viewerUserID, page, pageSize)
	if err != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve friends list")
		return
	}

	apiFriends := make([]apiModels.FriendListItem, 0, len(dbFriends))
	for _, f := range dbFriends {
		var displayName, avatarURL *string
		if f.DisplayName.Valid {
			displayName = &f.DisplayName.String
		}
		if f.AvatarURL.Valid {
			avatarURL = &f.AvatarURL.String
		}

		var status apiModels.FriendshipStatus
		if f.UserID == viewerUserID {
			status = apiModels.FriendshipStatusAPISelf
		} else {
			var logicalStatus dbModels.DBFriendshipStatus
			switch f.FriendshipStatus {
			case "accepted":
				logicalStatus = dbModels.DBFriendshipStatusFriends
			case "not_friends":
				logicalStatus = dbModels.DBFriendshipStatusNotFriends
			case "blocked":
				if f.RequesterID.Valid && f.RequesterID.String == viewerUserID {
					logicalStatus = dbModels.DBFriendshipStatusBlockedFirstBySecond
				} else {
					logicalStatus = dbModels.DBFriendshipStatusBlockedSecondByFirst
				}
			case "pending":
				if f.RequesterID.Valid && f.RequesterID.String == viewerUserID {
					logicalStatus = dbModels.DBFriendshipStatusPendingFirstSentToSecond
				} else {
					logicalStatus = dbModels.DBFriendshipStatusPendingSecondSentToFirst
				}
			default:
				logicalStatus = dbModels.DBFriendshipStatusUnknown
			}
			status = modelConverters.DBFriendshipStatusToAPIStatus(logicalStatus)
		}

		apiFriends = append(apiFriends, apiModels.FriendListItem{
			UserID:           f.UserID,
			Username:         f.Username,
			DisplayName:      displayName,
			AvatarURL:        avatarURL,
			FriendshipStatus: status,
		})
	}

	pagination := CalculatePagination(int64(totalCount), page, pageSize)
	response := apiModels.PaginatedFriendsListResponse{
		Friends:    apiFriends,
		Pagination: pagination,
	}

	Respond(w, r, http.StatusOK, response, "Friends list retrieved successfully")
}

var awardTypeToTitle = map[string]string{
	"oracle":          "The Oracle",
	"gambler":         "The Gambler",
	"treasure-hunter": "The Treasure Hunter",
	"scallywag":       "The Scallywag",
	"buccaneer":       "The Buccaneer",
	"maverick":        "The Maverick",
	"conservative":    "The Conservative",
}

// Retrieves the awards statistics for a user
// Path: /users/{user_id}/stats/awards
// Method: GET
func (uph *UserProfileHandler) HandleGetUserAwardStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"HandleGetUserAwardStats",
	)

	profileUserID, ok := PathVar(w, r, "user_id")
	if !ok {
		return
	}
	logger = logger.With().Str(l.ProfileUserIDKey, profileUserID).Logger()

	viewerUserID, _ := GetAuthenticatedUserIDFromSession(w, r, logger)
	authorized, err := isAuthorizedToViewStats(ctx, profileUserID, viewerUserID)
	if err != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to verify stats privacy")
		return
	}
	if !authorized {
		ErrorResponse(w, r, http.StatusForbidden, "You are not authorized to view these statistics")
		return
	}

	dbAwards, err := db.GetUserAwardsSummary(ctx, nil, profileUserID)
	if err != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve award statistics")
		return
	}

	apiResponse := make(apiModels.UserAwardsStatsResponse, len(dbAwards))
	for i, dbAward := range dbAwards {
		apiResponse[i] = apiModels.UserAwardStat{
			AwardType:  dbAward.AwardType,
			AwardTitle: awardTypeToTitle[dbAward.AwardType],
			Count:      dbAward.AwardCount,
			Percentile: dbAward.PercentilRank * 100,
		}
	}

	Respond(w, r, http.StatusOK, apiResponse, "User awards statistics retrieved successfully")
}

// HandleReportUser allows an authenticated user to report another user.
// Path: /users/{user_id}/report
// Method: POST
func (uph *UserProfileHandler) HandleReportUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(l.GetLoggerFromContext(ctx), userComponent, "HandleReportUser")

	reporterID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}

	reportedUserID, ok := PathVar(w, r, "user_id")
	if !ok {
		return
	}

	if reporterID == reportedUserID {
		ErrorResponse(w, r, http.StatusBadRequest, "You cannot report yourself")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if !ParseJSON(w, r, &req) {
		return
	}
	if !RequireFields(w, r, map[string]string{"reason": req.Reason}) {
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to submit report")
	if !txOk {
		return
	}
	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	_, opErr = db.CreateReport(ctx, tx, reporterID, reportedUserID, req.Reason)
	if opErr != nil {
		logger.Error().Err(opErr).Msg("Failed to create user report")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to submit report")
		responseSent = true
		return
	}

	if !responseSent {
		Respond(w, r, http.StatusCreated, nil, "Report submitted successfully")
		responseSent = true
	}
}

// Helper to check if a viewer can see a target user's stats
func isAuthorizedToViewStats(ctx context.Context, targetUserID, viewerUserID string) (bool, error) {
	if targetUserID == viewerUserID {
		return true, nil
	}

	targetUser, err := db.GetUserByID(ctx, nil, targetUserID)
	if err != nil {
		return false, err
	}

	switch targetUser.StatsPrivacy {
	case "public":
		return true, nil
	case "private":
		return false, nil
	case "friends_only":
		if viewerUserID == "" {
			return false, nil
		}
		status, err := db.GetFriendshipStatus(ctx, nil, viewerUserID, targetUserID)
		if err != nil {
			return false, err
		}
		return status == dbModels.DBFriendshipStatusFriends, nil
	default:
		return false, nil
	}
}
