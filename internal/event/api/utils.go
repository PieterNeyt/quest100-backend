package api

import "github.com/google/uuid"

// kwni ofda dees de beste plek is omda te zetten ma zag het nie echt ergens anders passen...
func parseOptionalUUID(s *string) (*uuid.UUID, error) {
	if s == nil {
		return nil, nil
	}
	parsed, err := uuid.Parse(*s)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
