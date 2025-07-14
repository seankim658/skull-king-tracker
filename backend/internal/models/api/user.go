package models

import (
	"time"
)

// --- Request Payloads ---

// Payload for updating user's preferred themes
type UpdateUserThemeRequest struct {
	UITheme    string `json:"ui_theme" validate:"required"`
	ColorTheme string `json:"color_theme" validate:"required"`
}

// Payload for updating a users profile display details
type UpdateUserProfileRequest struct {
	DisplayName  *string `json:"display_name"`
	AvatarURL    *string `json:"avatar_url"`
	StatsPrivacy *string `json:"stats_privacy"`
}

// --- Response Payloads ---

// Response for the /auth/me endpoint
type AuthenticatedUserResponse struct {
	User User `json:"user"`
}

// Main response for a user's public profile page
type UserProfileResponse struct {
	Profile UserProfile        `json:"profile"`
	Stats   *UserDetailedStats `json:"stats,omitempty"`
}

// Array of users returned from a search query
type UserSearchResponse []UserSearchItem

// Success message on logout
type LogoutResponse struct {
	Message string `json:"message"`
}

// Paginated response for a user's friends list
type PaginatedFriendsListResponse struct {
	Friends    []FriendListItem `json:"friends"`
	Pagination Pagination       `json:"pagination"`
}

// --- Component Structs ---

// Main API model representing a user's details for an authenticated session
type User struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	Email        *string   `json:"email,omitempty"`
	DisplayName  *string   `json:"display_name,omitempty"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	StatsPrivacy string    `json:"stats_privacy"`
	Role         string    `json:"role"`
	AvatarSource *string   `json:"avatar_source,omitempty"`
	UITheme      *string   `json:"ui_theme,omitempty"`
	ColorTheme   *string   `json:"color_theme,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastLoginAt  *string   `json:"last_login_at,omitempty"`
}

// The public-facing view of a user
type UserProfile struct {
	UserID            string           `json:"user_id"`
	Username          string           `json:"username"`
	DisplayName       *string          `json:"display_name,omitempty"`
	AvatarURL         *string          `json:"avatar_url,omitempty"`
	StatsPrivacy      string           `json:"stats_privacy"`
	CreatedAt         time.Time        `json:"created_at"`
	FriendCount       int              `json:"friend_count"`
	MutualFriendCount *int             `json:"mutual_friend_count,omitempty"`
	FriendshipStatus  FriendshipStatus `json:"friendship_status_with_viewer"`
}

// Represents a single OAuth provider linked to a user's account
type LinkedAccount struct {
	ProviderName        string  `json:"provider_name"`
	ProviderDisplayName *string `json:"provider_display_name,omitempty"`
	ProviderAvatarURL   *string `json:"provider_avatar_url,omitempty"`
	ProviderEmail       *string `json:"provider_email,omitempty"`
}

// A single result in a user search response
type UserSearchItem struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// Represents a single user in another user's friends list
type FriendListItem struct {
	UserID           string           `json:"user_id"`
	Username         string           `json:"username"`
	DisplayName      *string          `json:"display_name,omitempty"`
	AvatarURL        *string          `json:"avatar_url,omitempty"`
	FriendshipStatus FriendshipStatus `json:"friendship_status_with_viewer"`
}
