package models

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
