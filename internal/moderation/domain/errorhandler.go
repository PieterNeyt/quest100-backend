package domain

import "fmt"

type InvalidReportError struct {
	Message string
}

func (e *InvalidReportError) Error() string {
	return fmt.Sprintf("invalid report: %s", e.Message)
}
