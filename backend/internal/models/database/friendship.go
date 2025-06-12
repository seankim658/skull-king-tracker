package models

import "time"

type Friendship struct {
	FriendshipID string    `db:"friendship_id"`
	RequesterID  string    `db:"requester_id"`
	AddresseeID  string    `db:"addressee_id"`
	Status       string    `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// The possible states a friendship can be within the database
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
