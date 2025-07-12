package handlers

import (
	"math"
	"net/http"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
)

const statsComponent = "handlers-stats"

type StatsHandler struct {
	Cfg *cf.Config
}

func NewStatsHandler(cfg *cf.Config) *StatsHandler {
	return &StatsHandler{Cfg: cfg}
}

// Returns the site-wide summary statistics
// Path: /stats/summary
// Method: GET
func (sh *StatsHandler) HandleGetSiteSummaryStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		statsComponent,
		"HandleGetSiteSummaryStats",
	)

	dbStats, err := db.GetSiteWideSummaryStats(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve site-wide summary stats")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve site summary statistics")
		return
	}

	apiResponse := apiModels.SiteSummaryStatsResponse{
		TotalPlayers:      dbStats.TotalPlayers,
		SessionsThisMonth: dbStats.SessionsThisMonth,
		GamesThisMonth:    dbStats.GamesThisMonth,
		NewUsersThisMonth: dbStats.NewUsersThisMonth,
	}

	Respond(w, r, http.StatusOK, apiResponse, "Site summary statistics retrieved successfully")
}

// Retrieves the monthly global leaderboard
// Path: /stats/leaderboard
// Method: GET
func (sh *StatsHandler) HandleGetGlobalLeaderboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		statsComponent,
		"HandleGetGlobalLeaderboard",
	)

	dbLeaderboard, err := db.GetGlobalLeaderboard(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve global leaderboard")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve global leaderboard")
		return
	}

	apiLeaderboard := make(apiModels.GlobalLeaderboardresponse, len(dbLeaderboard))
	for i, row := range dbLeaderboard {
		apiLeaderboard[i] = apiModels.GlobalLeaderboardItem{
			Rank:             row.Rank,
			UserID:           row.UserID,
			PlayerName:       row.DisplayName,
			GamesPlayed:      row.GamesPlayed,
			Wins:             row.Wins,
			TotalPoints:      row.TotalPoints,
			AveragePoints:    math.Round(row.AveragePoints*100) / 100,
			AverageFinishPos: math.Round(row.AverageFinishPos*100) / 100,
		}
	}

  Respond(w, r, http.StatusOK, apiLeaderboard, "Global leaderboard retrieved successfully")
}
