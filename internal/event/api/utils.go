package api

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

func (r *UpdateEventRequest) UnmarshalJSON(data []byte) error {
	type Alias UpdateEventRequest
	aux := &struct {
		NewOrganizerID *string `json:"newOrganizerID"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if aux.NewOrganizerID != nil {
		parsed, err := uuid.Parse(*aux.NewOrganizerID)
		if err != nil {
			return fmt.Errorf("invalid newOrganizerID: %w", err)
		}
		r.NewOrganizerID = &parsed
	}

	return nil
}
