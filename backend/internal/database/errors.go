package database

import "errors"

// PostgreSQL errors
const (
	uniqueConstraintErrorCode = "23505"
)

// Database layer errors
var (
	// User
	ErrUserNotFound                 = errors.New("user not found")
	ErrUserProviderIdentityNotFound = errors.New("user provider identity not found")
	ErrUsernameTaken                = errors.New("username is already taken")
	ErrEmailTaken                   = errors.New("email is already registered")
	ErrProviderIdentityConflict     = errors.New("provider identity conflict (e.g., already linked or user has different link with provider)")
	ErrInvalidStatsPrivacy          = errors.New("invalid value for stats_privacy field")
	ErrDeleteLastProviderIdentity   = errors.New("cannot delete the last linked authentication method")

	// Friendship
	ErrFriendshipSelf          = errors.New("cannot friend self")
	ErrFriendshipAlreadyExists = errors.New("friendship already exists")
	ErrFriendshipNotFound      = errors.New("friendship not found")
	ErrFriendshipBlocked       = errors.New("friendship blocked")

	// Game
	ErrGameNotFound          = errors.New("game not found")
	ErrGameNotInPendingState = errors.New("game is not in pending state")
	ErrGameMissingDealer     = errors.New("game setup is incomplete, a starting dealer has not been set")
	ErrGameNotEnoughPlayers  = errors.New("cannot start a game with fewer than 2 player")

	// Round
	ErrNoRoundsFound             = errors.New("no rounds found for this game")
	ErrCannotEditBids            = errors.New("bids can only be edited when the round is in the 'playing' state")
	ErrCannotEditHistoricalRound = errors.New("tricks can only be edited for the most recently completed round")

	// Session
	ErrSessionNotFound = errors.New("game session not found")

	// Guest player
	ErrGuestPlayerNotFound = errors.New("guest player not found")

	// Game player
	ErrGamePlayerNotFound  = errors.New("game player not found")
	ErrPlayerAlreadyInGame = errors.New("player is already in this game")

	// Notification
	ErrNotificationNotFound = errors.New("notification not found")

	// Report
	ErrReportNotFound = errors.New("report not found")

	// Alert
	ErrAlertNotFound = errors.New("alert not found")
)
