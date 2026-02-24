package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChatRepository interface {
	GetMessages(eventID uuid.UUID) ([]ChatMessage, error)
	SaveMessage(msg *ChatMessage) error
}
type ChatMessage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	EventID   uuid.UUID `gorm:"type:uuid;not null;index" json:"eventId"`
	ProfileID uuid.UUID `gorm:"type:uuid;not null" json:"profileId"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	SentAt    time.Time `json:"sentAt"`
}

func NewChatMessage(eventID uuid.UUID, profileID uuid.UUID, content string) *ChatMessage {
	return &ChatMessage{
		ID:        uuid.New(),
		EventID:   eventID,
		ProfileID: profileID,
		Content:   content,
		SentAt:    time.Now(),
	}
}
