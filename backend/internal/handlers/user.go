package handlers

import (
	"errors"
	"math"
	"net/http"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
	modelConverters "github.com/seankim658/skullking/internal/models/convert"
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
	finalResponse := apiModels.UserProfileResponse{
		Profile: apiProfile,
	}

	// 4. Fetch Stats if Permitted
	canViewStats := false
	switch profileDBUser.StatsPrivacy {
	case "public":
		canViewStats = true
	case "friends_only":
		if apiFriendshipStatus == apiModels.FriendshipStatusAPIFriends || apiFriendshipStatus == apiModels.FriendshipStatusAPISelf {
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
				winPercentage = math.Round((float64(dbUserStats.TotalWins)/float64(dbUserStats.TotalGamesPlayed))*10000) / 100
			}
			finalResponse.Stats = &apiModels.UserStats{
				TotalGamesPlayed: dbUserStats.TotalGamesPlayed,
				TotalWins:        dbUserStats.TotalWins,
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
