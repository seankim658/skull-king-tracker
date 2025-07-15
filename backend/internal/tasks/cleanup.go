package tasks

import (
	"context"
	"database/sql"
	"time"

	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
)

const cleanupComponent = "tasks-cleanup"

// STarts a background goroutine to periodically clean up stale games and sessions
func StartCleanupTask(interval time.Duration) {
	logger := l.WithComponentAndSource(l.AppLog, cleanupComponent, "StartCleanupTask")
	logger.Info().Dur("interval", interval).Msg("Starting background cleanup task...")
	ticker := time.NewTicker(interval)

	go func() {
		for {
			<-ticker.C
			logger.Info().Msg("Running stale sessions and games cleanup...")
			if err := CleanupStaleEntities(); err != nil {
				logger.Error().Err(err).Msg("Error during stale entities cleanup task")
			}
		}
	}()
}

// Finds and process stale sessions and one-off
func CleanupStaleEntities() error {
	ctx := context.Background()
	logger := l.WithComponentAndSource(l.AppLog, cleanupComponent, "CleanupStaleEntities")

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to begin transaction for cleanup")
		return err
	}
	defer tx.Rollback()

	staleThreshold := time.Now().Add(-6 * time.Hour)

	// 1. Handle stale sessions
	staleSessions, err := db.GetStaleSessions(ctx, tx, staleThreshold)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get stale sessions")
		return err
	}

	if len(staleSessions) > 0 {
		logger.Info().Int(l.CountKey, len(staleSessions)).Msg("Found stale sessions to process")
		for _, sessionID := range staleSessions {
			if _, err := db.AbandonGamesInSession(ctx, tx, sessionID); err != nil {
				logger.Error().Err(err).Str(l.SessionIDKey, sessionID).Msg("Failed to abandon games in stale session")
				continue
			}

			completedGamesCount, err := db.CountCompletedGamesInSession(ctx, tx, sessionID)
			if err != nil {
				logger.Error().Err(err).Str(l.SessionIDKey, sessionID).Msg("Failed to count completed games in session")
				continue
			}

			if completedGamesCount > 0 {
				if err := db.UpdateSessionStatus(ctx, tx, sessionID, "completed", sql.NullTime{Time: time.Now(), Valid: true}); err != nil {
					logger.Error().Err(err).Str(l.SessionIDKey, sessionID).Msg("Failed to mark stale session as completed")
				}
			} else {
				if err := db.UpdateSessionStatus(ctx, tx, sessionID, "abandoned", sql.NullTime{}); err != nil {
					logger.Error().Err(err).Str(l.SessionIDKey, sessionID).Msg("Failed to mark stale session as abandoned")
				}
			}
		}
	} else {
		logger.Info().Msg("No stale sessions found")
	}

	// 2. Handle stale one-off games
	staleGames, err := db.GetStaleOneOffGames(ctx, tx, staleThreshold)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get stale one-off games")
		return err
	}

	if len(staleGames) > 0 {
		if _, err := db.AbandonGames(ctx, tx, staleGames); err != nil {
			logger.Error().Err(err).Msg("Failed to abandon stale one-off games")
			return err
		}
		logger.Info().Int(l.CountKey, len(staleGames)).Msg("Abandoned stale one-off games")
	} else {
		logger.Info().Msg("No stale one-off games found")
	}

	if err := tx.Commit(); err != nil {
		logger.Error().Err(err).Msg("Failed to commit cleanup transaction")
		return err
	}

	logger.Info().Msg("Cleanup task completed successfully")
	return nil
}
