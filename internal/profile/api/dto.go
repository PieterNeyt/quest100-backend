package api

import (
	"Quest100Backend/internal/profile/domain"

	"github.com/google/uuid"
)

type AwardsTransaction struct {
	Type     domain.KudoType `json:"type"`
	Receiver uuid.UUID       `json:"receiver"`
	Message  string          `json:"message"`
}
