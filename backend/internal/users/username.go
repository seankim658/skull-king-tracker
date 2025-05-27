package users

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"

	l "github.com/seankim658/skullking/internal/logger"
)

const usernameComponent = "users-username"

// Cleans a string to be used as a username, converts to lowercase, replaces
// common separators with underscores, filters to allowed characters, removes
// leading/trailing/multiple underscores, and enforces length constraints.
func SanitizeUsername(username string) string {
	// Convert to lowercase
	sanitized := strings.ToLower(username)
	// Replace common separators with underscores
	sanitized = strings.NewReplacer(" ", "_", ".", "_", "-", "_").Replace(sanitized)
	// Filter to allowed characters
	sanitized = regexp.MustCompile("[^a-z0-9_]+").ReplaceAllString(sanitized, "")
	// Remove leading/trailing underscores and collapse multiple underscores
	sanitized = strings.Trim(sanitized, "_")
	sanitized = regexp.MustCompile("_+").ReplaceAllString(sanitized, "_")
	// Handle length constraints
	if len(sanitized) > 30 {
		sanitized = strings.Trim(sanitized[:30], "_")
	}
	// If empty or too short, create fallback
	if len(sanitized) < 3 {
		return "user_" + strings.ToLower(uuid.NewString()[:8])
	}
	return sanitized
}

// Generates a random username (does not guarantee uniqueness)
func GenerateUniqueUsername(ctx context.Context, baseUsername, email string) (string, error) {
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		usernameComponent,
		"GenerateUniqueUsername",
	).With().Str("base_username", baseUsername).Str(l.EmailKey, email).Logger()

	// Determine base candidate
	candidate := baseUsername
	if strings.TrimSpace(candidate) == "" && strings.TrimSpace(email) != "" {
		parts := strings.Split(email, "@")
		if len(parts) > 0 {
			candidate = parts[0]
		}
	}

	// Generate username
	var username string
	if strings.TrimSpace(candidate) == "" {
		username = SanitizeUsername("user_" + strings.ToLower(uuid.NewString()[:8]))
	} else {
		sanitized := SanitizeUsername(candidate)
		if sanitized == "user" || len(sanitized) < 4 {
			username = SanitizeUsername(sanitized + "_" + strings.ToLower(uuid.NewString()[:4]))
		} else {
			username = sanitized
		}
	}

	logger.Debug().Str("base_candidate", candidate).Str("final_username", username).Msg("Generated username candidate")
	return username, nil
}
