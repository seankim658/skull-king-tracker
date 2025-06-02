package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	l "github.com/seankim658/skullking/internal/logger"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

const friendshipComponent = "database-friendship"

// Returns a users friends
func CountFriends(ctx context.Context, tx *sql.Tx, userID string) (int, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"CountFriends",
	).With().Str(l.UserIDKey, userID).Logger()

	query := `
  SELECT COUNT(*) FROM user_friendships
  WHERE (requester_id = $1 OR addressee_id = $1) AND status = 'accepted';
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to count friends")

	var count int
	err := querier.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to count friends")
		return 0, fmt.Errorf("error counting friends for user %s: %w", userID, err)
	}

	logger.Info().Int(l.CountKey, count).Msg("Friends counted successfully")
	return count, nil
}

// Determines the friendship status between two users
func GetFriendshipStatus(
	ctx context.Context,
	tx *sql.Tx,
	firstUserID, secondUserID string,
) (dbModels.DBFriendshipStatus, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"GetFriendshipStatus",
	).With().Str("first_user_id", firstUserID).Str("second_user_id", secondUserID).Logger()

	if firstUserID == "" || secondUserID == "" {
		logger.Error().Msg("firstUserID or secondUserID cannot be empty")
		return "", errors.New("firstUserID or secondUserID cannot be empty")
	}
	if firstUserID == secondUserID {
		return dbModels.DBFriendshipStatusSelf, nil
	}

	query := `
  SELECT status, requester_id, addressee_id
  FROM user_friendships
  WHERE (requester_id = $1 AND addressee_id = $2) OR (requester_id = $2 AND addressee_id = $1);
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get raw friendship status")

	var dbStatus, dbRequesterID, dbAddresseeID string
	err := querier.QueryRowContext(
		ctx, query,
		firstUserID,
		secondUserID,
	).Scan(
		&dbStatus,
		&dbRequesterID,
		&dbAddresseeID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dbModels.DBFriendshipStatusNotFriends, nil
		}
    logger.Error().Err(err).Msg("Failed to query friendship status")
		return "", fmt.Errorf(
			"error querying friendship status between %s and %s: %w",
			firstUserID,
			secondUserID,
			err,
		)
	}

	switch dbStatus {
	case "accepted":
		return dbModels.DBFriendshipStatusFriends, nil
	case "pending":
		if dbRequesterID == firstUserID && dbAddresseeID == secondUserID {
			return dbModels.DBFriendshipStatusPendingFirstSentToSecond, nil
		} else if dbRequesterID == secondUserID && dbAddresseeID == firstUserID {
			return dbModels.DBFriendshipStatusPendingSecondSentToFirst, nil
		} else {
			logger.Warn().Msg("Pending status with mismatched requester/addressee")
			return dbModels.DBFriendshipStatusUnknown, fmt.Errorf(
				"inconsistent pending state for user %s and %s", firstUserID, secondUserID,
			)
		}
	case "blocked":
		if dbRequesterID == firstUserID && dbAddresseeID == secondUserID {
			return dbModels.DBFriendshipStatusBlockedSecondByFirst, nil
		} else if dbRequesterID == secondUserID && dbAddresseeID == firstUserID {
			return dbModels.DBFriendshipStatusBlockedFirstBySecond, nil
		} else {
			logger.Warn().Msg("Block status with mismatched requester/addressee")
			return dbModels.DBFriendshipStatusUnknown, fmt.Errorf(
				"inconsistent blocked state for users %s and %s", firstUserID, secondUserID,
			)
		}
	case "declined":
		return dbModels.DBFriendshipStatusNotFriends, nil
	default:
		logger.Warn().Str("unknown_db_status", dbStatus).Msg("Unknown friendship status value")
		return dbModels.DBFriendshipStatusUnknown, fmt.Errorf(
			"unknown status '%s' for users %s and %s",
			dbStatus,
			firstUserID,
			secondUserID,
		)
	}
}
