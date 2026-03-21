package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AlreadyCompletedError struct{ ProfileID uuid.UUID }

func (e *AlreadyCompletedError) Error() string {
	return fmt.Sprintf("profile %s has already completed today's sudoku", e.ProfileID)
}

type InvalidMoveError struct {
	Message string
}

func (e *InvalidMoveError) Error() string {
	return fmt.Sprintf("invalid move: %s", e.Message)
}

type GameNotFoundError struct{ Date time.Time }

func (e *GameNotFoundError) Error() string {
	return fmt.Sprintf("no sudoku game found for %s", e.Date.Format("2006-01-02"))
}

type SessionNotFoundError struct {
	ProfileID uuid.UUID
	GameID    uuid.UUID
}

func (e *SessionNotFoundError) Error() string {
	return fmt.Sprintf("no session found for profile %s on game %s", e.ProfileID, e.GameID)
}
