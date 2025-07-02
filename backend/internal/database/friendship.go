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
		return ErrFriendshipNotFound
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
			return nil, ErrFriendshipNotFound
		}
		logger.Error().Err(err).Msg("Failed to get friendship by ID")
		return nil, fmt.Errorf("error getting friendship by ID %s: %w", friendshipID, err)
	}

	return &f, nil
}

// Deletes an accepted friendship by user IDs
func DeleteFriendship(ctx context.Context, tx *sql.Tx, userIDOne, userIDTwo string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"DeleteFriendship",
	).With().
		Str(l.UserIDOneKey, userIDOne).
		Str(l.UserIDTwoKey, userIDTwo).
		Logger()

	query := `
  DELETE FROM user_friendships
  WHERE status = 'accepted' AND (
    (requester_id = $1 AND addressee_id = $2) OR
    (requester_id = $2 AND addressee_id = $1)
  );
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to delete friendship")

	result, err := querier.ExecContext(ctx, query, userIDOne, userIDTwo)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute friendship deletion")
		return fmt.Errorf("error deleting friendship between %s and %s: %w", userIDOne, userIDTwo, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Msg("Error checking rows affected for friendship deletion")
		return fmt.Errorf("error checking rows affected for friendship deletion: %w", err)
	}
	if rowsAffected == 0 {
		return ErrFriendshipNotFound
	}

	logger.Info().Msg("Friendship deleted successfully")
	return nil
}

// Deletes a friendship by ID
func DeleteFriendshipByID(ctx context.Context, tx *sql.Tx, friendshipID string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"DeleteFriendshipByID",
	).With().Str(l.FriendshipIDKey, friendshipID).Logger()

	query := `
  DELETE FROM user_friendships
  WHERE friendship_id = $1;
  `
	result, err := querier.ExecContext(ctx, query, friendshipID)
	if err != nil {
		logger.Error().Err(err).Msg("Error deleting friendship")
		return fmt.Errorf("error deleting friendship by ID %s: %w", friendshipID, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrFriendshipNotFound
	}
	return nil
}

// Retrieves a pending friendship initiated by the requester
func GetPendingFriendshipByUsers(ctx context.Context, tx *sql.Tx, requesterID, addresseeID string) (*dbModels.Friendship, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"GetPendingFriendshipByUsers",
	).With().Str(l.RequesterIDKey, requesterID).Str(l.AddresseeIDKey, addresseeID).Logger()

	query := `
  SELECT friendship_id, requester_id, addressee_id, status
  FROM user_friendships
  WHERE requester_id = $1 AND addressee_id = $2 AND status = 'pending';
  `

	var f dbModels.Friendship
	err := querier.QueryRowContext(
		ctx,
		query,
		requesterID,
		addresseeID,
	).Scan(
		&f.FriendshipID,
		&f.RequesterID,
		&f.AddresseeID,
		&f.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFriendshipNotFound
		}
		logger.Error().Err(err).Msg("Failed to get pending friendship by users")
		return nil, fmt.Errorf("error getting pending friendship: %w", err)
	}
	return &f, nil
}

// Block another user
func BlockUser(ctx context.Context, tx *sql.Tx, blockerID, blockedID string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"BlockUser",
	).With().Str(l.BlockerIDKey, blockerID).Str(l.BlockedIDKey, blockedID).Logger()

	deleteQuery := `
  DELETE FROM user_friendships
  WHERE
    (requester_id = $1 AND addressee_id = $2) OR
    (requester_id = $2 AND addressee_id = $1);
  `
	if _, err := querier.ExecContext(ctx, deleteQuery, blockerID, blockedID); err != nil {
		logger.Error().Err(err).Msg("Failed to clear existing friendship before blocking")
		return fmt.Errorf("error clearing previous friendship state: %w", err)
	}

	insertQuery := `
  INSERT INTO user_friendships (requester_id, addressee_id, status)
  VALUES ($1, $2, 'blocked');
  `
	if _, err := querier.ExecContext(ctx, insertQuery, blockerID, blockedID); err != nil {
		logger.Error().Err(err).Msg("Failed to insert new blocked friendship record")
		return fmt.Errorf("error inserting block record: %w", err)
	}
	logger.Info().Msg("User blocked successfully")
	return nil
}

// Removes a block status for a user
func UnblockUser(ctx context.Context, tx *sql.Tx, unblockerID, unblockedID string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"UnblockUser",
	).With().
		Str(l.UnblockerIDKey, unblockerID).
		Str(l.UnblockedIDKey, unblockedID).
		Logger()

	query := `
  DELETE FROM user_friendships
  WHERE requester_id = $1 AND addressee_id = $2 AND status = 'blocked';
  `
	result, err := querier.ExecContext(ctx, query, unblockerID, unblockedID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute unblock query")
		return fmt.Errorf("error unblocking user: %s: %w", unblockedID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected for unblock: %w", err)
	}

	if rowsAffected == 0 {
		return ErrFriendshipNotFound
	}

	logger.Info().Msg("User unblocked successfully")
	return nil
}

// Retrieves a list of all accepted friends for a given user
func GetFriendsByUserID(ctx context.Context, tx *sql.Tx, userID string) ([]dbModels.UserSearchResult, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"GetFriendsByUserID",
	).With().Str(l.UserIDKey, userID).Logger()

	query := `
  SELECT
    u.user_id,
    u.username,
    u.display_name,
    u.avatar_url
  FROM user_friendships f
  JOIN users u ON
    CASE
      WHEN f.requester_id = $1 THEN u.user_id = f.addressee_id
      ELSE u.user_id = f.requester_id
    END
  WHERE
    (f.requester_id = $1 OR f.addressee_id = $1)
    AND f.status = 'accepted'
  ORDER BY u.username ASC;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get friends by user ID")

	rows, err := querier.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query user's friends")
		return nil, fmt.Errorf("error querying friends for user %s: %w", userID, err)
	}
	defer rows.Close()

	var friends []dbModels.UserSearchResult
	for rows.Next() {
		var friend dbModels.UserSearchResult
		if err := rows.Scan(
			&friend.UserID,
			&friend.Username,
			&friend.DisplayName,
			&friend.AvatarURL,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan friend row")
      return nil, fmt.Errorf("error scanning friend data for user %s: %w", userID, err)
		}
    friends = append(friends, friend)
	}

  if err = rows.Err(); err != nil {
    logger.Error().Err(err).Msg("Error iterating over friend rows")
    return nil, fmt.Errorf("error iterating friend rows for user %s: %w", userID, err)
  }

  logger.Info().Int(l.CountKey, len(friends)).Msg("Friends list retrieved successfully")
  return friends, nil
}
