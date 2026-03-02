package application

import (
	"Quest100Backend/internal/gotcha/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type GotchaService interface {
	CreateGame(campus string, startDate time.Time, killDeadlineHours int) (*domain.Game, error)
	StartGame(campus string) error

	OptIn(campus string, profileID uuid.UUID) error
	SubmitKill(gameID, hunterID, victimID uuid.UUID, photoURL string) (*domain.Kill, error)

	ReviewKill(killID, reviewerID uuid.UUID, approve bool) error

	GetKillFeed(campus string, limit, offset int) ([]*domain.Kill, error)
	LikeKill(killID, profileID uuid.UUID) error
	UnlikeKill(killID, profileID uuid.UUID) error

	GetMyStatus(campus string, profileID uuid.UUID) (*domain.Participant, error)
	GetLeaderboard(campus string) ([]*domain.Participant, error)

	ProcessTimeouts() error
}

type gotchaService struct {
	gameRepo        domain.GameRepository
	participantRepo domain.ParticipantRepository
	killRepo        domain.KillRepository
	propRepo        domain.PropRepository
}

func NewGotchaService(
	gameRepo domain.GameRepository,
	participantRepo domain.ParticipantRepository,
	killRepo domain.KillRepository,
	propRepo domain.PropRepository,
) GotchaService {
	return &gotchaService{gameRepo, participantRepo, killRepo, propRepo}
}

func (s *gotchaService) CreateGame(campus string, startDate time.Time, killDeadlineHours int) (*domain.Game, error) {
	game := &domain.Game{
		ID:                uuid.New(),
		Campus:            campus,
		Status:            domain.StatusOptIn,
		StartDate:         startDate,
		KillDeadlineHours: killDeadlineHours,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := s.gameRepo.SaveGame(game); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}
	return game, nil
}

func (s *gotchaService) StartGame(campus string) error {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return err
	}
	if game.Status != domain.StatusOptIn {
		return fmt.Errorf("game for campus %s is not in opt-in phase", campus)
	}

	participants, err := s.participantRepo.GetParticipantsByGame(game.ID)
	if err != nil {
		return err
	}
	for _, p := range participants {
		game.Participants = append(game.Participants, *p)
	}

	if err := game.AssignTargets(); err != nil {
		return err
	}

	for i := range game.Participants {
		prop, err := s.propRepo.GetRandomProp()
		if err == nil {
			game.Participants[i].AssignedPropID = &prop.ID
		}
	}

	return s.gameRepo.SaveGame(game)
}

func (s *gotchaService) OptIn(campus string, profileID uuid.UUID) error {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return fmt.Errorf("no game found for campus %s: %w", campus, err)
	}
	if game.Status != domain.StatusOptIn {
		return fmt.Errorf("opt-in period has ended")
	}

	existing, _ := s.participantRepo.GetParticipant(game.ID, profileID)
	if existing != nil {
		return &domain.AlreadyOptedInError{ProfileID: profileID}
	}

	p := &domain.Participant{
		ID:        uuid.New(),
		GameID:    game.ID,
		ProfileID: profileID,
		IsAlive:   true,
		OptedInAt: time.Now(),
	}
	return s.participantRepo.SaveParticipant(p)
}

func (s *gotchaService) SubmitKill(gameID, hunterID, victimID uuid.UUID, photoURL string) (*domain.Kill, error) {
	hunter, err := s.participantRepo.GetParticipant(gameID, hunterID)
	if err != nil || hunter == nil {
		return nil, &domain.NotParticipantError{ProfileID: hunterID}
	}
	if hunter.TargetID == nil || *hunter.TargetID != victimID {
		return nil, fmt.Errorf("victim is not your current target")
	}

	prop := hunter.AssignedPropID

	kill := &domain.Kill{
		ID:        uuid.New(),
		GameID:    gameID,
		HunterID:  hunterID,
		VictimID:  victimID,
		PhotoURL:  photoURL,
		PropID:    prop,
		Status:    domain.KillPending,
		CreatedAt: time.Now(),
	}
	if err := s.killRepo.SaveKill(kill); err != nil {
		return nil, fmt.Errorf("failed to submit kill: %w", err)
	}
	return kill, nil
}

func (s *gotchaService) ReviewKill(killID, reviewerID uuid.UUID, approve bool) error {
	kill, err := s.killRepo.GetKillByID(killID)
	if err != nil {
		return fmt.Errorf("kill not found: %w", err)
	}
	if kill.Status != domain.KillPending {
		return fmt.Errorf("kill already reviewed")
	}

	now := time.Now()
	kill.ReviewedBy = &reviewerID
	kill.ReviewedAt = &now

	if !approve {
		kill.Status = domain.KillDenied
		return s.killRepo.SaveKill(kill)
	}

	kill.Status = domain.KillApproved

	game, err := s.gameRepo.GetGameByID(kill.GameID)
	if err != nil {
		return err
	}
	participants, err := s.participantRepo.GetParticipantsByGame(game.ID)
	if err != nil {
		return err
	}
	for _, p := range participants {
		game.Participants = append(game.Participants, *p)
	}

	if err := game.ProcessKill(kill.HunterID, kill.VictimID); err != nil {
		return err
	}

	// sla alles op
	if err := s.killRepo.SaveKill(kill); err != nil {
		return err
	}
	return s.gameRepo.SaveGame(game)
}

func (s *gotchaService) ProcessTimeouts() error {
	expired, err := s.participantRepo.GetExpiredParticipants(time.Now())
	if err != nil {
		return err
	}

	for _, victim := range expired {
		game, err := s.gameRepo.GetGameByID(victim.GameID)
		if err != nil {
			continue
		}
		participants, _ := s.participantRepo.GetParticipantsByGame(game.ID)
		for _, p := range participants {
			game.Participants = append(game.Participants, *p)
		}
		_ = game.ProcessTimeout(victim.ProfileID)
		_ = s.gameRepo.SaveGame(game)
	}
	return nil
}

func (s *gotchaService) GetKillFeed(campus string, limit, offset int) ([]*domain.Kill, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, err
	}
	return s.killRepo.GetKillFeed(game.ID, limit, offset)
}

func (s *gotchaService) LikeKill(killID, profileID uuid.UUID) error {
	return s.killRepo.SaveKillLike(&domain.KillLike{
		KillID:    killID,
		ProfileID: profileID,
		LikedAt:   time.Now(),
	})
}

func (s *gotchaService) UnlikeKill(killID, profileID uuid.UUID) error {
	return s.killRepo.DeleteKillLike(killID, profileID)
}

func (s *gotchaService) GetMyStatus(campus string, profileID uuid.UUID) (*domain.Participant, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, err
	}
	return s.participantRepo.GetParticipant(game.ID, profileID)
}

func (s *gotchaService) GetLeaderboard(campus string) ([]*domain.Participant, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, err
	}
	return s.participantRepo.GetParticipantsByGame(game.ID)
}
