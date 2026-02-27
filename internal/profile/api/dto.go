package api

import (
	"Quest100Backend/internal/profile/domain"

	"github.com/google/uuid"
)

type AwardTransaction struct {
	Type     domain.KudoType `json:"type"`
	Receiver uuid.UUID       `json:"receiver"`
	Message  string          `json:"message"`
}
type SyncProfileResponse struct {
	Profile                 *domain.Profile `json:"profile"`
	MicrosoftProfilePicture string          `json:"microsoftProfilePicture"`
}
