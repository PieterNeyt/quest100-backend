package database

import (
	"Quest100Backend/internal/communication/domain"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) domain.ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) GetChatById(chatId uuid.UUID) (*domain.Chat, error) {
	var chat domain.Chat

	result := r.db.Preload("Members").First(&chat, "id = ?", chatId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("chat %v not found", chatId)
		}
		return nil, fmt.Errorf("failed to fetch chat %v: %w", chatId, result.Error)
	}
	return &chat, nil
}

func (r *chatRepository) SaveChat(chat *domain.Chat) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Members").Save(chat).Error; err != nil {
			return err
		}

		if len(chat.Members) > 0 {
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(chat).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *chatRepository) SaveMessage(message *domain.Message) (*domain.Message, error) {
	if err := r.db.Create(message).Error; err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}
	if err := r.db.Preload("Sender", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "first_name", "last_name")
	}).First(&message, "id = ?", message.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}
	return message, nil
}

func (r *chatRepository) DeleteChat(chatId uuid.UUID) error {
	if err := r.db.Delete(&domain.Chat{}, "id = ?", chatId).Error; err != nil {
		return fmt.Errorf("failed to delete chat %v: %w", chatId, err)
	}
	return nil
}

func (r *chatRepository) GetAllMessagesOfChatRoom(chatId uuid.UUID) ([]*domain.Message, error) {
	var messages []*domain.Message

	result := r.db.Preload("Sender", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "first_name", "last_name")
	}).Find(&messages, "chat_id = ?", chatId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("chat %v not found", chatId)
		}
		return nil, fmt.Errorf("failed to fetch all messages of chat %v: %w", chatId, result.Error)
	}
	return messages, nil
}
