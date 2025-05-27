package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	l "github.com/seankim658/skullking/internal/logger"
)

type Config struct {
	AppEnv               string
	APIPort              string
	AppBaseURL           string
	FrontendBaseURL      string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string
	JWTSecret            string
	SessionSecretKey     string
	SessionEncryptionKey string
	ProviderAuthConfig   ProvidersConfig
	Log                  l.LogConfig
}

type ProvidersConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	// TODO : add others
}

// Load configuration from environment variables and .env files.
// Precedence:
// 1. Actual system Environment variables
// 2. .env.{APP_ENV} (e.g., .env.development)
// 3. .env
// 4. Hardcoded default values
func Load() (*Config, error) {
	// Determine application environment
	appEnv := strings.ToLower(os.Getenv("APP_ENV"))
	if appEnv == "" {
		appEnv = "development" // Default environment
		log.Info().Msg("App env not set, defaulting to 'development'")
	} else {
		log.Info().Str(l.AppEnvKey, appEnv).Msg("Application environment")
	}

	// Attempt to load environment specific .env file first
	envSpecificFilename := fmt.Sprintf(".env.%s", appEnv)
	if err := godotenv.Load(envSpecificFilename); err == nil {
		log.Info().Str(l.FileKey, envSpecificFilename).Msg("Loaded backend local environment .env file")
	} else if !os.IsNotExist(err) {
		log.Warn().Err(err).Str(l.FileKey, envSpecificFilename).Msg("Error loading backend local environment specific .env file")
	}

	if err := godotenv.Load(); err == nil {
		log.Info().Str(l.FileKey, ".env").Msg("Loaded backend local base .env configuration")
	} else if !os.IsNotExist(err) {
		log.Warn().Err(err).Str(l.FileKey, ".env").Msg("Error loading backend local base .env file")
	}

	var err error
	var dbUser, dbPassword, dbName, jwtSecret, sessionSecretKey, sessionEncryptionKey string

	dbUser, err = getRequiredEnv("POSTGRES_USER")
	if err != nil {
		return nil, err
	}
	dbPassword, err = getRequiredEnv("POSTGRES_PASSWORD")
	if err != nil {
		return nil, err
	}
	dbName, err = getRequiredEnv("POSTGRES_DB")
	if err != nil {
		return nil, err
	}
	jwtSecret, err = getRequiredEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}
	sessionSecretKey, err = getRequiredEnv("SESSION_SECRET_KEY")
	if err != nil {
		return nil, err
	}
	sessionEncryptionKey, err = getRequiredEnv("SESSION_ENCRYPTION_KEY")
	if err != nil {
		return nil, err
	}

	googleClientID := getEnv("GOOGLE_CLIENT_ID", "")
	googleClientSecret := getEnv("GOOGLE_CLIENT_SECRET", "")

	apiPort := getEnv("API_PORT", "8080")

	defaultAppBaseURL := "http://localhost:8081"
	if appEnv == "production" {
		defaultAppBaseURL = "https://skullking.uk"
	}
	appBaseURL := getEnv("APP_BASE_URL", defaultAppBaseURL)

	defaultFrontendBaseURL := "http://localhost:5173"
	if appEnv == "production" {
		defaultFrontendBaseURL = "https://skullking.uk"
	}
	frontendBaseURL := getEnv("FRONTEND_BASE_URL", defaultFrontendBaseURL)

	cfg := &Config{
		AppEnv:               appEnv,
		APIPort:              apiPort,
		AppBaseURL:           appBaseURL,
		FrontendBaseURL:      frontendBaseURL,
		DBHost:               getEnv("DB_HOST", "db"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               dbUser,
		DBPassword:           dbPassword,
		DBName:               dbName,
		DBSSLMode:            getEnv("DB_SSLMODE", "disable"),
		JWTSecret:            jwtSecret,
		SessionSecretKey:     sessionSecretKey,
		SessionEncryptionKey: sessionEncryptionKey,
		ProviderAuthConfig: ProvidersConfig{
			GoogleClientID:     googleClientID,
			GoogleClientSecret: googleClientSecret,
		},
		Log: l.LogConfig{
			AppLogPath:     getEnv("APP_LOG_PATH", "./logs/app.log"),
			AccessLogPath:  getEnv("ACCESS_LOG_PATH", "./logs/network.log"),
			ConsoleLogging: getBoolEnv("LOG_CONSOLE_LOGGING", appEnv == "development"),
			UseJSONFormat:  getBoolEnv("LOG_USE_JSON_FORMAT", appEnv != "development"),
			LogLevel:       getEnv("LOG_LEVEL", "info"),
			MaxSizeMB:      getIntEnv("LOG_MAX_SIZE_MB", 100),
			MaxBackups:     getIntEnv("LOG_MAX_BACKUPS", 3),
			MaxAgeDays:     getIntEnv("LOG_MAX_AGE_DAYS", 28),
			Compress:       getBoolEnv("LOG_COMPRESS", true),
		},
	}

	log.Info().Str("APP_ENV", cfg.AppEnv).Msg("Configuration loaded successfully")
	return cfg, nil
}

// Helper to retrieve an environment variable or return a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Debug().Str(l.KeyKey, key).Str(l.FallbackKey, fallback).Msg("Environment variable not set, using fallback")
	return fallback
}

// Helper to retrieve an environment variable as an integer or return a fallback
func getIntEnv(key string, fallback int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		} else {
			log.Warn().
				Err(err).
				Str(l.KeyKey, key).
				Str(l.ValueKey, valueStr).
				Msg("Failed to parse integer environment variable, using fallback")
		}
	} else {
		log.Debug().
			Str(l.KeyKey, key).
			Int(l.FallbackKey, fallback).
			Msg("Integer environment variable not set, using fallback")
	}
	return fallback
}

// Helper to retrieve an environment variable as a boolean or return a fallback
func getBoolEnv(key string, fallback bool) bool {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		} else {
			log.Warn().
				Err(err).
				Str(l.KeyKey, key).
				Str(l.ValueKey, valueStr).
				Msg("Failed to parse boolean environment variable, using fallback")
		}
	} else {
		log.Debug().
			Str(l.KeyKey, key).
			Bool(l.FallbackKey, fallback).
			Msg("Boolean environment variable not set, using fallback")
	}
	return fallback
}

// Helper to retrieve a required environment variable, if not found, returns an error
func getRequiredEnv(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return "", fmt.Errorf("critical: required environment variable '%s' not set or is empty", key)
	}
	return value, nil
}
