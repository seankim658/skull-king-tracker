package handlers

import (
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
