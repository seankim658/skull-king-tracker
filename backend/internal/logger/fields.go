package logger

const (
	ComponentKey  = "component"
	FallbackKey   = "fallback"
	FileKey       = "file"
	KeyKey        = "key"
	SourceKey     = "source"
	ValueKey      = "value"
	AppEnvKey     = "APP_ENV"
	PanicKey      = "panic"
	StackTraceKey = "stack_trace"
	PathKey       = "path"
	SizeBytesKey  = "size_bytes"

	// General network fields
	InternalServerError  = "internal-server-error"
	TooManyRequestsError = "too-many-requests"
	RequestIDKey         = "request_id"
	StatusCodeKey        = "status_code"
	ContentTypeKey       = "content_type"
	SchemeKey            = "scheme"
	ImageURLKey          = "image_url"
	HostnameKey          = "hostname"
	IPKey                = "ip"

	// --- Database Logging Fields ---
	QueryKey                = "query"
	ArgsKey                 = "args"
	UpdatesKey              = "updates"
	CountKey                = "count"
	PostgresErrorCodeKey    = "pg_code"
	PostgresConstraintKey   = "pg_constraint"
	PosgresErrorDetailKey   = "pg_detail"
	PostgresErrorMessageKey = "pg_error_messsage"
	StatusKey               = "status"

	// User
	UserIDKey     = "user_id"
	UsernameKey   = "username"
	EmailKey      = "email"
	UIThemeKey    = "ui_theme"
	ColorThemeKey = "color_theme"
	AvatarURLKey  = "avatar_url"

	// Auth
	ProviderKey           = "provider"
	ProviderUserIDKey     = "provider_user_id"
	ProviderIdentityIDKey = "provider_identity_id"
	CallbackUrlKey        = "callback_url"
	GothUsernameKey       = "goth_username"
	GothNickNameKey       = "goth_nickname"

	// Game
	GameIDKey        = "game_id"
	ScorekeeperIDKey = "scorekeeper_id"

	// Session
	SessionIDKey   = "session_id"
	SessionNameKey = "session_name"

	// Guest player
	GuestPlayerIDKey   = "guest_player_id"
	GuestPlayerNameKey = "guest_player_display_name"

	// Game player
	GamePlayerIDKey = "game_player_id"
	SeatingOrderKey = "seating_order"
	GameStatusKey   = "game_status"
)
