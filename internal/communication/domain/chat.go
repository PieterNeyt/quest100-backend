package domain

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

type ChatRepository interface {
	GetChatById(chatId uuid.UUID) (*Chat, error)
	SaveChat(chat *Chat) error
	DeleteChat(chatId uuid.UUID) error
}

type Chat struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Members  []Member  `gorm:"foreignKey:ChatId;reference:ID;constraint:OnDelete:CASCADE;" json:"members"`
	Messages []Message `gorm:"foreignKey:ChatId;reference:ID;constraint:OnDelete:CASCADE;" json:"messages"`
}

type Member struct {
	ProfileId uuid.UUID `gorm:"type:uuid;primary_key;" json:"profileId"`
	ChatId    uuid.UUID `gorm:"type:uuid;primary_key;" json:"chatId"`
}

type Message struct {
	ID      uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	ChatId  uuid.UUID `gorm:"type:uuid;;" json:"chatId"`
	Sender  uuid.UUID `gorm:"type:uuid;" json:"sender"`
	Message string    `gorm:"type:text;" json:"message"`
}

func CreateChat(eventId uuid.UUID, profileId uuid.UUID) *Chat {
	return &Chat{
		ID:       eventId,
		Messages: []Message{},
		Members:  []Member{{ProfileId: profileId, ChatId: eventId}},
	}
}

func (c *Chat) JoinChat(profileId uuid.UUID) error {
	inChat := slices.Contains(c.Members, Member{ProfileId: profileId, ChatId: c.ID})
	if inChat {
		return fmt.Errorf("already in chat %s", c.ID)
	}
	c.Members = append(c.Members, Member{ProfileId: profileId, ChatId: c.ID})
	return nil
}

func (c *Chat) LeaveChat(profileId uuid.UUID) error {
	inChat := slices.Contains(c.Members, Member{ProfileId: profileId})
	if !inChat {
		return fmt.Errorf("not in chat %s", c.ID)
	}
	c.Members = slices.DeleteFunc(c.Members, func(i Member) bool {
		return i.ProfileId == profileId
	})
	return nil
}
