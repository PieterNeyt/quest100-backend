package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type NotParticipantError struct{ ProfileID uuid.UUID }

func (e *NotParticipantError) Error() string {
	return fmt.Sprintf("profile %s is not a participant in this game", e.ProfileID)
}

type AlreadyOptedInError struct{ ProfileID uuid.UUID }

func (e *AlreadyOptedInError) Error() string {
	return fmt.Sprintf("profile %s already opted in", e.ProfileID)
}

type GameNotActiveError struct{ Campus string }

func (e *GameNotActiveError) Error() string {
	return fmt.Sprintf("no active game for campus %s", e.Campus)
}

type UnauthorizedError struct{ Message string }

func (e *UnauthorizedError) Error() string { return e.Message }
