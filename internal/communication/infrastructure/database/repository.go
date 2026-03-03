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

	result := r.db.Preload("Members").Preload("Messages").First(&chat, "id = ?", chatId)
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
		// 1. Save the Chat itself first
		if err := tx.Omit("Members").Save(chat).Error; err != nil {
			return err
		}

		// 2. Now save the associations
		if len(chat.Members) > 0 {
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(chat).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *chatRepository) DeleteChat(chatId uuid.UUID) error {
	if err := r.db.Delete(&domain.Chat{}, "id = ?", chatId).Error; err != nil {
		return fmt.Errorf("failed to delete chat %v: %w", chatId, err)
	}
	return nil
}
