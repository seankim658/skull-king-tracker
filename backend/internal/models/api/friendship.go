package models

type FriendshipStatus string

const (
	FriendshipStatusAPIViewerNotAuth        FriendshipStatus = "viewer_not_authenticated"
	FriendshipStatusAPISelf                 FriendshipStatus = "self"
	FriendshipStatusAPIFriends              FriendshipStatus = "friends"
	FriendshipStatusAPINotFriends           FriendshipStatus = "not_friends"
	FriendshipStatusAPIPendingSentToViewer  FriendshipStatus = "pending_sent_to_viewer"
	FriendshipStatusAPIPendingSentToProfile FriendshipStatus = "pending_sent_to_profile"
	FriendshipStatusAPIBlockedByViewer      FriendshipStatus = "blocked_by_viewer"
	FriendshipStatusAPIBlockedByProfileUser FriendshipStatus = "blocked_by_profile_user"
	FriendshipStatusAPIUnknown              FriendshipStatus = "unknown"
)

// --- Request Payloads ---

// Request payload for accepting or declining a friend request
type RespondToFriendRequest struct {
	Response string `json:"response" validate:"required,oneof=accept decline"`
}

// Request payload for sending a new friend request
type SendFriendRequest struct {
	AddresseeID string `json:"addressee_id" validate:"required,uuid"`
}
