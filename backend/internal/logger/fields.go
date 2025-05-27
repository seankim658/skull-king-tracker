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

	// General network fields
	InternalServerError  = "internal-server-error"
	TooManyRequestsError = "too-many-requests"
	RequestIDKey         = "request_id"

	// --- Database Logging Fields ---
	QueryKey                = "query"
	ArgsKey                 = "args"
	UpdatesKey              = "updates"
	CountKey                = "count"
	PostgresErrorCodeKey    = "pg_code"
	PostgresConstraintKey   = "pg_constraint"
	PosgresErrorDetailKey   = "pg_detail"
	PostgresErrorMessageKey = "pg_error_messsage"

	// User
	UserIDKey     = "user_id"
	UsernameKey   = "username"
	EmailKey      = "email"
	UIThemeKey    = "ui_theme"
	ColorThemeKey = "color_theme"

	// Auth
	ProviderKey           = "provider"
	ProviderUserIDKey     = "provider_user_id"
	ProviderIdentityIDKey = "provider_identity_id"
	CallbackUrlKey        = "callback_url"
	GothUsernameKey       = "goth_username"
	GothNickNameKey       = "goth_nickname"
)
