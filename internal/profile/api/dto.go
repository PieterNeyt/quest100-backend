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

type AssetDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	LayerOrder int    `json:"layer_order"`
	Price      int    `json:"price"`
	IsOwned    bool   `json:"isOwned"`
	Link       string `json:"link"`
	Equipped   bool   `json:"equipped"`
}

type CategoryDTO struct {
	Name  string     `json:"name"`
	Items []AssetDTO `json:"items"`
}
