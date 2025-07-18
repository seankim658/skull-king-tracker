package handlers

import (
	"net/http"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
)

const alertComponent = "handlers-alert"

type AlertHandler struct {
	Cfg *cf.Config
}

func NewAlertHandler(cfg *cf.Config) *AlertHandler {
	return &AlertHandler{Cfg: cfg}
}

// Retrieves currently active site-wide alerts
// Path: /alerts/active
// Method: GET
func (ah *AlertHandler) HandleGetActiveSiteAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		alertComponent,
		"HandleGetActiveSiteAlerts",
	)

	dbAlerts, err := db.GetActiveSiteAlerts(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve active alerts")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve site alerts")
		return
	}

	apiAlerts := make([]apiModels.SiteAlertResponse, len(dbAlerts))
	for i, a := range dbAlerts {
		apiAlerts[i] = apiModels.SiteAlertResponse{
			AlertID:   a.AlertID,
			Title:     a.Title,
			Body:      a.Body,
			StartTime: a.StartTime,
			EndTime:   a.EndTime,
			IsActive:  a.IsActive,
		}
	}

	Respond(w, r, http.StatusOK, apiAlerts, "Active site alerts retrieved successfully")
}
