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
  ErrFriendshipSelf = errors.New("cannot friend self")

	// Game
	ErrGameNotFound = errors.New("game not found")

	// Session
	ErrSessionNotFound = errors.New("game session not found")

	// Guest player
	ErrGuestPlayerNotFound = errors.New("guest player not found")

	// Game player
	ErrGamePlayerNotFound  = errors.New("game player not found")
	ErrPlayerAlreadyInGame = errors.New("player is already in this game")
)
