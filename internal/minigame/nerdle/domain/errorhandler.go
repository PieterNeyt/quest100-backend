package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AlreadySolvedError struct{ ProfileID uuid.UUID }

func (e *AlreadySolvedError) Error() string {
	return fmt.Sprintf("profile %s has already solved today's nerdle", e.ProfileID)
}

type MaxAttemptsReachedError struct{ ProfileID uuid.UUID }

func (e *MaxAttemptsReachedError) Error() string {
	return fmt.Sprintf("profile %s has reached the maximum number of attempts", e.ProfileID)
}

type GameNotFoundError struct{ Date time.Time }

func (e *GameNotFoundError) Error() string {
	return fmt.Sprintf("no nerdle game found for %s", e.Date.Format("2006-01-02"))
}

type InvalidGuessError struct {
	Guess   string
	Message string
}

func (e *InvalidGuessError) Error() string {
	return fmt.Sprintf("invalid guess %q: %s", e.Guess, e.Message)
}
