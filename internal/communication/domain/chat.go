package domain

import (
	"Quest100Backend/internal/profile/domain"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

type ChatRepository interface {
	GetChatById(chatId uuid.UUID) (*Chat, error)
	SaveChat(chat *Chat) error
	SaveMessage(message *Message) (*Message, error)
	DeleteChat(chatId uuid.UUID) error
	GetAllMessagesOfChatRoom(chatId uuid.UUID) ([]*Message, error)
}

type Chat struct {
	ID       uuid.UUID    `gorm:"type:uuid;primary_key;" json:"id"`
	Members  []ChatMember `gorm:"foreignKey:ChatId;reference:ID;constraint:OnDelete:CASCADE;" json:"members"`
	Messages []Message    `gorm:"foreignKey:ChatId;reference:ID;constraint:OnDelete:CASCADE;" json:"messages"`
}

type ChatMember struct {
	ProfileId uuid.UUID `gorm:"type:uuid;primary_key;" json:"profileId"`
	ChatId    uuid.UUID `gorm:"type:uuid;primary_key;" json:"chatId"`
}

type Message struct {
	ID       uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	ChatId   uuid.UUID      `gorm:"type:uuid;" json:"chatId"`
	SenderId uuid.UUID      `gorm:"type:uuid;" json:"senderId"`
	Sender   domain.Profile `gorm:"foreignKey:SenderId;references:ID;" json:"sender"`
	Message  string         `gorm:"type:text;" json:"message"`
}

func CreateChat(eventId uuid.UUID) *Chat {
	return &Chat{
		ID:       eventId,
		Messages: []Message{},
		Members:  []ChatMember{},
	}
}

func (c *Chat) JoinChat(profileId uuid.UUID) error {
	inChat := slices.Contains(c.Members, ChatMember{ProfileId: profileId, ChatId: c.ID})
	if inChat {
		return fmt.Errorf("already in chat %s", c.ID)
	}
	c.Members = append(c.Members, ChatMember{ProfileId: profileId, ChatId: c.ID})
	return nil
}

func (c *Chat) LeaveChat(profileId uuid.UUID) error {
	oldLen := len(c.Members)

	c.Members = slices.DeleteFunc(c.Members, func(member ChatMember) bool {
		return member.ProfileId == profileId
	})
	if oldLen == len(c.Members) {
		return fmt.Errorf("not in chat %s", c.ID)
	}
	return nil
}

func (c *Chat) AddMessage(senderId uuid.UUID, message string) error {
	c.Messages = append(c.Messages, Message{
		ID:       uuid.New(),
		ChatId:   c.ID,
		SenderId: senderId,
		Message:  message,
	})
	return nil
}

func CreateMessage(senderId uuid.UUID, chatId uuid.UUID, message string) *Message {
	return &Message{
		ID:       uuid.New(),
		ChatId:   chatId,
		SenderId: senderId,
		Message:  message,
	}
}
