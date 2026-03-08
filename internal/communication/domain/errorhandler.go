package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type NotInChatError struct {
	UserID uuid.UUID
	ChatID uuid.UUID
}

func (e *NotInChatError) Error() string {
	return fmt.Sprintf("user %s is not in chat %s", e.UserID, e.ChatID)
}
