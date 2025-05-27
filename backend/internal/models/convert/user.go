package models

import (
	"errors"
	"time"

	apiModels "github.com/seankim658/skullking/internal/models/api"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

func DBUserToAPIUser(dbUser *dbModels.User) (*apiModels.User, error) {
	if dbUser == nil {
		return nil, errors.New("cannot convert nil db user to api user")
	}
	var email, displayName, avatarURL, uiTheme, colorTheme, lastLoginAt *string
	if dbUser.Email.Valid {
		email = &dbUser.Email.String
	}
	if dbUser.DisplayName.Valid {
		displayName = &dbUser.DisplayName.String
	}
	if dbUser.AvatarURL.Valid {
		avatarURL = &dbUser.AvatarURL.String
	}
	if dbUser.UITheme.Valid {
		uiTheme = &dbUser.UITheme.String
	}
	if dbUser.ColorTheme.Valid {
		colorTheme = &dbUser.ColorTheme.String
	}
	if dbUser.LastLoginAt.Valid {
		formattedTime := dbUser.LastLoginAt.Time.Format(time.RFC3339)
		lastLoginAt = &formattedTime
	}

	return &apiModels.User{
		UserID:       dbUser.UserID,
		Username:     dbUser.Username,
		Email:        email,
		DisplayName:  displayName,
		AvatarURL:    avatarURL,
		StatsPrivacy: dbUser.StatsPrivacy,
		UITheme:      uiTheme,
		ColorTheme:   colorTheme,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		LastLoginAt:  lastLoginAt,
	}, nil
}

func DBProviderIdentityToLinkedAccount(userProviderIdentity *dbModels.UserProviderIdentity) (*apiModels.LinkedAccount, error) {
	if userProviderIdentity == nil {
		return nil, errors.New("cannot convert nil db user provider to api linked account")
	}
	var displayName, avatarURL, email *string
	if userProviderIdentity.ProviderDisplayName.Valid {
		displayName = &userProviderIdentity.ProviderDisplayName.String
	}
	if userProviderIdentity.ProviderAvatarURL.Valid {
		avatarURL = &userProviderIdentity.ProviderAvatarURL.String
	}
	if userProviderIdentity.ProviderEmail.Valid {
		email = &userProviderIdentity.ProviderEmail.String
	}

	return &apiModels.LinkedAccount{
		ProviderName:        userProviderIdentity.ProviderName,
		ProviderDisplayName: displayName,
		ProviderAvatarURL:   avatarURL,
		ProviderEmail:       email,
	}, nil
}
