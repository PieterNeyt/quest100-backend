package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AlreadyCompletedError struct{ ProfileID uuid.UUID }

func (e *AlreadyCompletedError) Error() string {
	return fmt.Sprintf("profile %s has already completed today's minesweeper", e.ProfileID)
}

type GameOverError struct{ ProfileID uuid.UUID }

func (e *GameOverError) Error() string {
	return fmt.Sprintf("profile %s has hit a mine — game is over", e.ProfileID)
}

type GameNotFoundError struct{ Date time.Time }

func (e *GameNotFoundError) Error() string {
	return fmt.Sprintf("no minesweeper game found for %s", e.Date.Format("2006-01-02"))
}

type InvalidMoveError struct {
	Message string
}

func (e *InvalidMoveError) Error() string {
	return fmt.Sprintf("invalid move: %s", e.Message)
}

type SessionNotFoundError struct {
	ProfileID uuid.UUID
	Date      time.Time
}

func (e *SessionNotFoundError) Error() string {
	return fmt.Sprintf("no session found for profile %s on %s", e.ProfileID, e.Date.Format("2006-01-02"))
}
