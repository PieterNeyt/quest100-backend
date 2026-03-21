package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type GameNotFoundError struct {
	Date time.Time
}

func (e *GameNotFoundError) Error() string {
	return fmt.Sprintf("no nerdle game found for date %s", e.Date.Format("2006-01-02"))
}

type AlreadySolvedError struct {
	ProfileID uuid.UUID
}

func (e *AlreadySolvedError) Error() string {
	return fmt.Sprintf("profile %s already solved today's nerdle", e.ProfileID)
}

type MaxAttemptsReachedError struct {
	ProfileID uuid.UUID
}

func (e *MaxAttemptsReachedError) Error() string {
	return fmt.Sprintf("profile %s has no attempts remaining", e.ProfileID)
}

type InvalidGuessError struct {
	Guess   string
	Message string
}

func (e *InvalidGuessError) Error() string {
	return fmt.Sprintf("invalid guess %q: %s", e.Guess, e.Message)
}
