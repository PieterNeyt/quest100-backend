package api

import "Quest100Backend/internal/profile/domain"

type SyncProfileResponse struct {
	Profile                 *domain.Profile `json:"profile"`
	MicrosoftProfilePicture string          `json:"microsoftProfilePicture"`
}
