package application

import (
	chatDom "Quest100Backend/internal/communication/domain"
	eventDom "Quest100Backend/internal/event/domain"
	profileApp "Quest100Backend/internal/profile/application"
	"fmt"

	"github.com/google/uuid"
)

type ChatService interface {
	IsUserInChat(userId uuid.UUID, chatId uuid.UUID) error
	CreateChatRoom(roomId uuid.UUID) error
	JoinChatRoom(userId uuid.UUID, eventId uuid.UUID) error
	LeaveChatRoom(userId uuid.UUID, eventId uuid.UUID) error
	DeleteChatRoom(userId uuid.UUID, eventId uuid.UUID) error
	CreateMessage(chatId uuid.UUID, userId uuid.UUID, message string) error
}

type chatService struct {
	profileServ profileApp.ProfileService
	chatRepo    chatDom.ChatRepository
	eventRepo   eventDom.EventRepository
}

func NewChatService(profileServ profileApp.ProfileService, chatRepo chatDom.ChatRepository, eventRep eventDom.EventRepository) ChatService {
	return &chatService{
		profileServ: profileServ,
		chatRepo:    chatRepo,
		eventRepo:   eventRep,
	}
}

func (s *chatService) IsUserInChat(userId uuid.UUID, chatId uuid.UUID) error {
	profile, err := s.profileServ.GetProfileById(userId)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	chat, err := s.chatRepo.GetChatById(chatId)
	if err != nil {
		return fmt.Errorf("chat not found: %w", err)
	}
	for _, member := range chat.Members {
		if member.ProfileId == profile.ID {
			return nil
		}
	}
	return &chatDom.NotInChatError{ChatID: chatId, UserID: userId}
}

func (s *chatService) CreateChatRoom(eventId uuid.UUID) error {
	event, err := s.eventRepo.GetEventByID(eventId)
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	chat := chatDom.CreateChat(event.ID)
	if err := s.chatRepo.SaveChat(chat); err != nil {
		return fmt.Errorf("chat save failed: %w", err)
	}
	return nil
}

func (s *chatService) JoinChatRoom(userId uuid.UUID, eventId uuid.UUID) error {
	profile, err := s.profileServ.GetProfileById(userId)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	event, err := s.eventRepo.GetEventByID(eventId)
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	chat, err := s.chatRepo.GetChatById(event.ID)
	if err != nil {
		return fmt.Errorf("chat not found: %w", err)
	}

	if err := chat.JoinChat(profile.ID); err != nil {
		return fmt.Errorf("join chat failed: %w", err)
	}
	if err := s.chatRepo.SaveChat(chat); err != nil {
		return fmt.Errorf("chat save failed: %w", err)
	}
	return nil
}

func (s *chatService) LeaveChatRoom(userId uuid.UUID, eventId uuid.UUID) error {
	_, err := s.profileServ.GetProfileById(userId)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	event, err := s.eventRepo.GetEventByID(eventId)
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	chat, err := s.chatRepo.GetChatById(event.ID)
	if err != nil {
		return fmt.Errorf("chat not found: %w", err)
	}

	if err := chat.LeaveChat(userId); err != nil {
		return fmt.Errorf("leave chat failed: %w", err)
	}

	if err := s.chatRepo.SaveChat(chat); err != nil {
		return fmt.Errorf("chat save failed: %w", err)
	}
	return nil
}

func (s *chatService) DeleteChatRoom(userId uuid.UUID, eventId uuid.UUID) error {
	_, err := s.profileServ.GetProfileById(userId)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	event, err := s.eventRepo.GetEventByID(eventId)
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	chat, err := s.chatRepo.GetChatById(event.ID)
	if err != nil {
		return fmt.Errorf("chat not found: %w", err)
	}
	if err := s.chatRepo.DeleteChat(chat.ID); err != nil {
		return fmt.Errorf("chat delete failed: %w", err)
	}
	return nil
}

func (s *chatService) CreateMessage(chatId uuid.UUID, userId uuid.UUID, message string) error {
	chat, err := s.chatRepo.GetChatById(chatId)
	if err != nil {
		return fmt.Errorf("chat not found: %w", err)
	}

	profile, err := s.profileServ.GetProfileById(userId)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}
	if err := chat.AddMessage(profile.ID, message); err != nil {
		return fmt.Errorf("add message failed: %w", err)
	}

	if err := s.chatRepo.SaveChat(chat); err != nil {
		return fmt.Errorf("chat save failed: %w", err)
	}
	return nil
}
