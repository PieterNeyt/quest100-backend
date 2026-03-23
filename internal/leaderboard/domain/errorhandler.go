package domain

import (
	"fmt"
)

type NegativeKudosError struct {
	Arg     int
	Message string
}

func (e *NegativeKudosError) Error() string {
	return fmt.Sprintf("%d - %s", e.Arg, e.Message)
}
