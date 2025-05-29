package database

import (
	"database/sql"
	"errors"
	"fmt"

	dbModels "github.com/seankim658/skullking/internal/models/database"
)

// Defines the interface for scanning a single row, satisfied by *sql.Row
type RowScanner interface {
	Scan(dest ...any) error
}

// Scan a user row
func scanUser(row RowScanner) (*dbModels.User, error) {
	user := &dbModels.User{}
	err := row.Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.StatsPrivacy,
		&user.UITheme,
		&user.ColorTheme,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error scanning user data: %w", err)
	}
	return user, nil
}

// Scan a game session row
func scanGameSession(row RowScanner) (*dbModels.GameSession, error) {
	s := &dbModels.GameSession{}
	err := row.Scan(
		&s.SessionID,
		&s.SessionName,
		&s.CreatedByUserID,
		&s.Status,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("error scanning game session data: %w", err)
	}
	return s, nil
}

// Scan a game row
func scanGame(row RowScanner) (*dbModels.Game, error) {
	g := &dbModels.Game{}
	err := row.Scan(
		&g.GameID,
		&g.SessionID,
		&g.CreatedByUserID,
		&g.CurrentScorekeeperUserID,
		&g.Status,
		&g.StartingDealerGamePlayerID,
		&g.PlayerSeatingOrderRandomized,
		&g.CreatedAt,
		&g.UpdatedAt,
		&g.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGameNotFound
		}
		return nil, fmt.Errorf("error scanning game data: %w", err)
	}
	return g, nil
}

// Scan a guest player row
func scanGuestPlayer(row RowScanner) (*dbModels.GuestPlayer, error) {
	gp := &dbModels.GuestPlayer{}
	err := row.Scan(
		&gp.GuestPlayerID,
		&gp.DisplayName,
		&gp.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGuestPlayerNotFound
		}
		return nil, fmt.Errorf("error scanning guest player data: %w", err)
	}
	return gp, nil
}

// Scan a game player row
func scanGamePlayer(row RowScanner) (*dbModels.GamePlayer, error) {
	p := &dbModels.GamePlayer{}
	err := row.Scan(
		&p.GamePlayerID,
		&p.GameID,
		&p.UserID,
		&p.GuestPlayerID,
		&p.SeatingOrder,
		&p.FinalScore,
		&p.FinishingPosition,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGamePlayerNotFound
		}
		return nil, fmt.Errorf("error scanning game player data: %w", err)
	}
	return p, nil
}
