package models

import (
	apiModels "github.com/seankim658/skullking/internal/models/api"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

func DBFriendshipStatusToAPIStatus(dbStatus dbModels.DBFriendshipStatus) apiModels.FriendshipStatus {
	switch dbStatus {
	case dbModels.DBFriendshipStatusFriends:
		return apiModels.FriendshipStatusAPIFriends
	case dbModels.DBFriendshipStatusNotFriends:
		return apiModels.FriendshipStatusAPINotFriends
	case dbModels.DBFriendshipStatusPendingFirstSentToSecond:
		return apiModels.FriendshipStatusAPIPendingSentToProfile
	case dbModels.DBFriendshipStatusPendingSecondSentToFirst:
		return apiModels.FriendshipStatusAPIPendingSentToViewer
	case dbModels.DBFriendshipStatusBlockedSecondByFirst:
		return apiModels.FriendshipStatusAPIBlockedByProfileUser
	case dbModels.DBFriendshipStatusBlockedFirstBySecond:
		return apiModels.FriendshipStatusAPIBlockedByViewer
	default:
		return apiModels.FriendshipStatusAPIUnknown
	}
}
