package models

import (
	"database/sql"
	"time"
)

// --- Core Entities ---

// Maps to the `user_friendships` table
type Friendship struct {
	FriendshipID string    `db:"friendship_id"`
	RequesterID  string    `db:"requester_id"`
	AddresseeID  string    `db:"addressee_id"`
	Status       string    `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// --- Enums & Composite Structs ---

// Represents the logical state of a friendship
type DBFriendshipStatus string

const (
	DBFriendshipStatusSelf                     DBFriendshipStatus = "self"
	DBFriendshipStatusViewerNotAuthenticated   DBFriendshipStatus = "viewer_not_authenticated"
	DBFriendshipStatusNotFriends               DBFriendshipStatus = "not_friends"
	DBFriendshipStatusFriends                  DBFriendshipStatus = "friends"
	DBFriendshipStatusPendingFirstSentToSecond DBFriendshipStatus = "pending_first_sent_to_second"
	DBFriendshipStatusPendingSecondSentToFirst DBFriendshipStatus = "pending_second_sent_to_first"
	DBFriendshipStatusBlockedSecondByFirst     DBFriendshipStatus = "blocked_second_by_first"
	DBFriendshipStatusBlockedFirstBySecond     DBFriendshipStatus = "blocked_first_by_second"
	DBFriendshipStatusUnknown                  DBFriendshipStatus = "unknown"
)

// Composite struct for queries that fetch a user's friends
type FriendshipWithViewerStatus struct {
	UserID           string         `db:"user_id"`
	Username         string         `db:"username"`
	DisplayName      sql.NullString `db:"display_name"`
	AvatarURL        sql.NullString `db:"avatar_url"`
	UpdatedAt        sql.NullTime   `db:"updated_at"`
	FriendshipStatus string         `db:"friendship_status_with_viewer"`
	RequesterID      sql.NullString `db:"requester_id"`
}
