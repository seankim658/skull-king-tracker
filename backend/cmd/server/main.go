package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"

	"github.com/seankim658/skullking/internal/auth"
	"github.com/seankim658/skullking/internal/config"
	"github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	"github.com/seankim658/skullking/internal/router"
)

func main() {
	bootstrapLogger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	cfg, err := config.Load()
	if err != nil {
		bootstrapLogger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	l.InitLoggers(cfg.Log)
	log := l.AppLog

	if err := database.Connect(cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to the database")
	}
	defer database.Close()

	if err := auth.InitAuth(cfg); err != nil {
		log.Error().Err(err).Msg("Failed to initialize authentication providers. Some OAuth logins may not be available or app may not function correctly.")
	}

	log.Info().Msg("Application components initialized")

	r := router.New(cfg)

	log.Info().Str("port", cfg.APIPort).Str("app_base_url", cfg.AppBaseURL).Msgf("Starting server on port %s", cfg.APIPort)

	server := &http.Server{
		Addr:    ":" + cfg.APIPort,
		Handler: r,
		// TODO
		// ReadTimeout
		// WriteTimeout
		//IdleTimeout
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msgf("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shut down server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

  log.Info().Msg("Shutting down server...")
}
