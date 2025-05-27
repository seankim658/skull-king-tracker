package logger

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	AppLog    zerolog.Logger
	AccessLog zerolog.Logger
)

type contextKey string

const loggerKey = contextKey("logger") // Key used to store the logger in the context

type LogConfig struct {
	// Path for the application log file
	AppLogPath string
	// Path for the network log file
	AccessLogPath string
	// Whether to also log to console (for development)
	ConsoleLogging bool
	// Whether to format logs as JSON
	UseJSONFormat bool
	// Logging level
	LogLevel string
	// Max size in megabytes before log rotation
	MaxSizeMB int
	// Max number of old log files to retain
	MaxBackups int
	// Max number of days to retain old log files
	MaxAgeDays int
	// Whether to compress rotated log files
	Compress bool
}

// Initalizes the AppLog and AccessLog based on provided configuration, this function
// should be called once at application startup.
func InitLoggers(cfg LogConfig) {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		// Fallback to info log level if parsing fails and log an internal warning using a temporary
		// console logger for this initial setup message
		initErrLogger := zerolog.New(os.Stderr).With().Timestamp().Logger()
		initErrLogger.Warn().Err(err).Str("provided_level", cfg.LogLevel).Msg("Invalid log level, defaulting to 'info'")
		level = zerolog.InfoLevel
	}

	AppLog = createLoggerInternal(
		cfg.AppLogPath,
		level,
		cfg.ConsoleLogging,
		cfg.UseJSONFormat,
		cfg.MaxSizeMB,
		cfg.MaxBackups,
		cfg.MaxAgeDays,
		cfg.Compress,
		false,
	)
	AppLog.Info().
		Str("app_log_path", cfg.AppLogPath).
		Str("log_level", level.String()).
		Bool("console_logging", cfg.ConsoleLogging).
		Bool("use_json_format", cfg.UseJSONFormat).
		Int("max_size_mb", cfg.MaxSizeMB).
		Int("max_backups", cfg.MaxBackups).
		Int("max_age_days", cfg.MaxAgeDays).
		Bool("compress", cfg.Compress).
		Msg("Application logger initialized")

	AccessLog = createLoggerInternal(
		cfg.AccessLogPath,
		level,
		cfg.ConsoleLogging,
		cfg.UseJSONFormat,
		cfg.MaxSizeMB,
		cfg.MaxBackups,
		cfg.MaxAgeDays,
		cfg.Compress,
		true,
	)
	AccessLog.Info().Msg("Access logger initialized")
}

// Helper to create the main loggers
func createLoggerInternal(
	logPath string,
	level zerolog.Level,
	consoleLogging bool,
	useJSONFormatForFile bool,
	maxSizeMB int,
	maxBackups int,
	maxAgeDays int,
	compress bool,
	isAccessLog bool,
) zerolog.Logger {
	var writers []io.Writer

	if logPath != "" {
		logDir := filepath.Dir(logPath)
		if _, statErr := os.Stat(logDir); os.IsNotExist(statErr) {
			if mkdirErr := os.MkdirAll(logDir, 0750); mkdirErr != nil {
				bootstrapLogger := zerolog.New(os.Stderr).With().Timestamp().Logger()
				bootstrapLogger.Error().Err(mkdirErr).Str("path", logDir).Msg("Failed to create log directory")
			}
		}

		// Only proceed with file writer if directory exists
		if _, statErr := os.Stat(logDir); !os.IsNotExist(statErr) {
			fileWriter := &lumberjack.Logger{
				Filename:   logPath,
				MaxSize:    maxSizeMB,
				MaxBackups: maxBackups,
				MaxAge:     maxAgeDays,
				Compress:   compress,
			}
			if useJSONFormatForFile {
				writers = append(writers, fileWriter)
			} else {
				writers = append(writers, zerolog.New(fileWriter).With().Timestamp().Logger())
			}
		}
	}

	// Console writer
	if consoleLogging {
		output := os.Stderr
		if isAccessLog {
			output = os.Stdout
		}
		consoleWriter := zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
		writers = append(writers, consoleWriter)
	}

	if len(writers) == 0 {
		// Fallback to stderr console logging if no other writers are configured
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	multiLevelWriter := io.MultiWriter(writers...)

	// Base logger
	logger := zerolog.New(multiLevelWriter).Level(level).With().Timestamp().Logger()

	if consoleLogging && !isAccessLog {
		logger = logger.With().Caller().Logger()
	}

	return logger
}

func WithComponent(baseLogger zerolog.Logger, component string) zerolog.Logger {
	return baseLogger.With().Str(ComponentKey, component).Logger()
}

func WithSource(baseLogger zerolog.Logger, source string) zerolog.Logger {
	return baseLogger.With().Str(SourceKey, source).Logger()
}

func WithComponentAndSource(baseLogger zerolog.Logger, component, source string) zerolog.Logger {
	tmpLogger := WithComponent(baseLogger, component)
	return WithSource(tmpLogger, source)
}

// Creates a new context logger with the provided logger
func NewContextWithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// Retrieves the logger from the context
func GetLoggerFromContext(ctx context.Context) zerolog.Logger {
	if logger, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
		return logger
	}
	// Fallback to global AppLog
	AppLog.Warn().Msg("No logger found in context, using global AppLog as fallback")
	return AppLog
}
