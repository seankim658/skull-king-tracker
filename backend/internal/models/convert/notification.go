package models

import (
	"errors"
	"time"

	apiModels "github.com/seankim658/skullking/internal/models/api"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

func DBNotificationsToAPIResponse(dbNotifications []dbModels.NotificationWithActor) []apiModels.Notification {
	apiNotifications := make([]apiModels.Notification, 0, len(dbNotifications))

	for _, dbNotif := range dbNotifications {
		var actorDisplayName, actorAvatarURL, friendshipID, actorUpdatedAt *string
		if dbNotif.ActorDisplayName.Valid {
			actorDisplayName = &dbNotif.ActorDisplayName.String
		}
		if dbNotif.ActorAvatarURL.Valid {
			actorAvatarURL = &dbNotif.ActorAvatarURL.String
		}
		if dbNotif.FriendshipID.Valid {
			friendshipID = &dbNotif.FriendshipID.String
		}
		if dbNotif.ActorUpdatedAt.Valid {
			formattedTime := dbNotif.ActorUpdatedAt.Time.Format(time.RFC3339)
			actorUpdatedAt = &formattedTime
		}

		apiNotif := apiModels.Notification{
			NotificationID: dbNotif.NotificationID,
			Type:           dbNotif.Type,
			Message:        dbNotif.Message,
			IsRead:         dbNotif.IsRead,
			FriendshipID:   friendshipID,
			CreatedAt:      dbNotif.CreatedAt,
			Actor: apiModels.NotificationActor{
				UserID:      dbNotif.ActorUserID,
				Username:    dbNotif.ActorUsername,
				DisplayName: actorDisplayName,
				AvatarURL:   actorAvatarURL,
				UpdatedAt:   actorUpdatedAt,
			},
		}
		apiNotifications = append(apiNotifications, apiNotif)
	}
	return apiNotifications
}

func DBNotificationWithActorToAPI(dbNotif *dbModels.NotificationWithActor) (*apiModels.Notification, error) {
	if dbNotif == nil {
		return nil, errors.New("cannot convert nil db notification to api notification")
	}
	var actorDisplayName, actorAvatarURL, friendshipID, actorUpdatedAt *string
	if dbNotif.ActorDisplayName.Valid {
		actorDisplayName = &dbNotif.ActorDisplayName.String
	}
	if dbNotif.ActorAvatarURL.Valid {
		actorAvatarURL = &dbNotif.ActorAvatarURL.String
	}
	if dbNotif.FriendshipID.Valid {
		friendshipID = &dbNotif.FriendshipID.String
	}
	if dbNotif.ActorUpdatedAt.Valid {
		formattedTime := dbNotif.ActorUpdatedAt.Time.Format(time.RFC3339)
		actorUpdatedAt = &formattedTime
	}
	apiNotif := &apiModels.Notification{
		NotificationID: dbNotif.NotificationID,
		Type:           dbNotif.Type,
		Message:        dbNotif.Message,
		IsRead:         dbNotif.IsRead,
		FriendshipID:   friendshipID,
		CreatedAt:      dbNotif.CreatedAt,
		Actor: apiModels.NotificationActor{
			UserID:      dbNotif.ActorUserID,
			Username:    dbNotif.ActorUsername,
			DisplayName: actorDisplayName,
			AvatarURL:   actorAvatarURL,
			UpdatedAt:   actorUpdatedAt,
		},
	}
	return apiNotif, nil
}
