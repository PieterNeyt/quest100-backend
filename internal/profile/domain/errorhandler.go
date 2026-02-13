package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type NegativeKudosError struct {
	Arg     int
	Message string
}

func (e *NegativeKudosError) Error() string {
	return fmt.Sprintf("%d - %s", e.Arg, e.Message)
}

type ProfileIdError struct {
	Arg     uuid.UUID
	Message string
}

func (e *ProfileIdError) Error() string {
	return fmt.Sprintf("%d - %s", e.Arg, e.Message)
}
