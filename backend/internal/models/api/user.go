package models

import (
	"time"
)

// API user model
type User struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	Email        *string   `json:"email,omitempty"`
	DisplayName  *string   `json:"display_name,omitempty"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	StatsPrivacy string    `json:"stats_privacy"`
	AvatarSource *string   `json:"avatar_source,omitempty"`
	UITheme      *string   `json:"ui_theme,omitempty"`
	ColorTheme   *string   `json:"color_theme,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastLoginAt  *string   `json:"last_login_at,omitempty"`
}

type LinkedAccount struct {
	ProviderName        string  `json:"provider_name"`
	ProviderDisplayName *string `json:"provider_display_name,omitempty"`
	ProviderAvatarURL   *string `json:"provider_avatar_url,omitempty"`
	ProviderEmail       *string `json:"provider_email,omitempty"`
}

type AuthenticatedUserResponse struct {
	User User `json:"user"`
}

type UpdateUserThemeRequest struct {
	UITheme    string `json:"ui_theme" validate:"required"`
	ColorTheme string `json:"color_theme" validate:"required"`
}

type UpdateUserProfileRequest struct {
	DisplayName  *string `json:"display_name"`
	AvatarURL    *string `json:"avatar_url"`
	StatsPrivacy *string `json:"stats_privacy"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

type UserProfile struct {
	UserID           string           `json:"user_id"`
	Username         string           `json:"username"`
	DisplayName      *string          `json:"display_name,omitempty"`
	AvatarURL        *string          `json:"avatar_url,omitempty"`
	StatsPrivacy     string           `json:"stats_privacy"`
	CreatedAt        time.Time        `json:"created_at"`
	FriendCount      int              `json:"friend_count"`
	FriendshipStatus FriendshipStatus `json:"friendship_status_with_viewer"`
}

type UserProfileResponse struct {
	Profile UserProfile `json:"profile"`
	Stats   *UserStats  `json:"stats,omitempty"`
}
