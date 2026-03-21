package api

import (
	"time"

	"github.com/google/uuid"
)

type SubmitGuessRequest struct {
	Guess string `json:"guess" binding:"required"`
}

type GameStatusResponse struct {
	GameID       uuid.UUID  `json:"gameId"`
	Date         time.Time  `json:"date"`
	HasSession   bool       `json:"hasSession"`
	Solved       bool       `json:"solved"`
	AttemptsUsed int        `json:"attemptsUsed"`
	AttemptsLeft int        `json:"attemptsLeft"`
	MaxAttempts  int        `json:"maxAttempts"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
}
