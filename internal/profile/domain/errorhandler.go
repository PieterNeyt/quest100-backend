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

type NoPlayerStatsError struct {
	Arg     int
	Message string
}

func (e *NoPlayerStatsError) Error() string {
	return fmt.Sprintf("%d - %s", e.Arg, e.Message)
}

type ProfileIdError struct {
	Arg     uuid.UUID
	Message string
}

func (e *ProfileIdError) Error() string {
	return fmt.Sprintf("%d - %s", e.Arg, e.Message)
}

type ProfileError struct {
	Arg     int
	Message string
}

func (e *ProfileError) Error() string {
	return fmt.Sprintf("%d - %s", e.Arg, e.Message)
}

type DuplicateAttendanceError struct {
	ProfileID uuid.UUID
	ClassID   int
	Message   string
}

func (e *DuplicateAttendanceError) Error() string {
	return fmt.Sprintf("duplicate attendance for profile %v in class %v: %s", e.ProfileID, e.ClassID, e.Message)
}
