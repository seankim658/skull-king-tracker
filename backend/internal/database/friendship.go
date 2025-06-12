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

// Creates a new pending friendship or re-activates a declined one
func CreateFriendship(ctx context.Context, tx *sql.Tx, requesterID, addresseeID string) (string, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"CreateFriendship",
	).With().
		Str(l.RequesterIDKey, requesterID).
		Str(l.AddresseeIDKey, addresseeID).
		Logger()

	var existingID, existingStatus string
	checkQuery := `
  SELECT friendship_id, status 
  FROM user_friendships 
  WHERE 
    (requester_id = $1 AND addressee_id = $2) OR
    (requester_id = $2 AND addressee_id = $1);
  `
	err := querier.QueryRowContext(
		ctx, checkQuery,
		requesterID, addresseeID,
	).Scan(&existingID, &existingStatus)

	if err == nil {
		switch existingStatus {
		case "pending", "accepted":
			return "", ErrFriendshipAlreadyExists
		case "blocked":
			return "", ErrFriendshipBlocked
		case "declined":
			updateQuery := `
      UPDATE user_friendships 
      SET requester_id = $1, addressee_id = $2, status = 'pending', updated_at = NOW()
      WHERE friendship_id = $3
      RETURNING friendship_id;
      `
			_, updateErr := querier.ExecContext(ctx, updateQuery, requesterID, addresseeID, existingID)
			if updateErr != nil {
        logger.Error().Err(updateErr).Msg("Error re-initiating friendship request")
				return "", fmt.Errorf("error re-initiating friendship: %w", updateErr)
			}
			logger.Info().Msg("Re-initiated a friend request over a previously declined one.")
			return existingID, nil
		}
	}

	if !errors.Is(err, sql.ErrNoRows) {
    logger.Error().Err(err).Msg("Error checking for existing friendship")
		return "", fmt.Errorf("error checking for existing friendship: %w", err)
	}

	insertQuery := `
  INSERT INTO user_friendships (
    requester_id, addressee_id, status
  ) 
  VALUES ($1, $2, 'pending') 
  RETURNING friendship_id;
  `
	var newFriendshipID string
	if err := querier.QueryRowContext(
		ctx,
		insertQuery,
		requesterID,
		addresseeID,
	).Scan(&newFriendshipID); err != nil {
		constraintMappings := map[string]error{"uq_requester_addressee": ErrFriendshipAlreadyExists}
		handled, appErr := HandlePgError(err, logger, constraintMappings)
		if handled {
			return "", appErr
		}
		return "", fmt.Errorf("error creating friendship: %w", err)
	}
	return newFriendshipID, nil
}

// Updates the status of a friendship request
func UpdateFriendshipStatus(ctx context.Context, tx *sql.Tx, friendshipID, status string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"UpdateFriendshipStatus",
	).With().
		Str(l.FriendshipIDKey, friendshipID).
		Str(l.FriendshipStatusKey, status).
		Logger()

	query := `
  UPDATE user_friendships
  SET status = $1, updated_at = NOW()
  WHERE friendship_id = $2;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to update friendship status")

	result, err := querier.ExecContext(ctx, query, status, friendshipID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute friendship status update")
		return fmt.Errorf("error updating friendship status for %s: %w", friendshipID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
    logger.Error().Err(err).Msg("Error checking rows affected for friendship status update")
		return fmt.Errorf(
			"error checking rows affected for friendship %s status update: %w", friendshipID, err,
		)
	}
	if rowsAffected == 0 {
		return ErrFriendhipNotFound
	}

	logger.Info().Msg("Friendship status updated successfully")
	return nil
}

// Retrieves a single friendship record by its ID
func GetFriendshipByID(ctx context.Context, tx *sql.Tx, friendshipID string) (*dbModels.Friendship, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"GetFriendshipByID",
	).With().Str(l.FriendshipIDKey, friendshipID).Logger()

	query := `
  SELECT friendship_id, requester_id, addressee_id, status
  FROM user_friendships WHERE friendship_id = $1;
  `

	var f dbModels.Friendship
	err := querier.QueryRowContext(
		ctx,
		query,
		friendshipID,
	).Scan(&f.FriendshipID, &f.RequesterID, &f.AddresseeID, &f.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFriendhipNotFound
		}
		logger.Error().Err(err).Msg("Failed to get friendship by ID")
		return nil, fmt.Errorf("error getting friendship by ID %s: %w", friendshipID, err)
	}

	return &f, nil
}
