package application

import (
	"Quest100Backend/internal/minigame/nerdle/domain"
	profileApp "Quest100Backend/internal/profile/application"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type NerdleService interface {
	GetTodayGame() (*domain.NerdleGame, error)
	GetOrCreateSession(profileID uuid.UUID) (*domain.NerdleSession, error)
	SubmitGuess(profileID uuid.UUID, guess string) (*GuessResponse, error)
}

type GuessResponse struct {
	Attempt      *domain.NerdleAttempt `json:"attempt"`
	Session      *domain.NerdleSession `json:"session"`
	Solved       bool                  `json:"solved"`
	GameOver     bool                  `json:"gameOver"`
	AttemptsLeft int                   `json:"attemptsLeft"`
	KudosEarned  int                   `json:"kudosEarned"`
}

type nerdleService struct {
	repo           domain.NerdleRepository
	profileService profileApp.ProfileService
}

func NewNerdleService(repo domain.NerdleRepository, profileService profileApp.ProfileService) NerdleService {
	return &nerdleService{
		repo:           repo,
		profileService: profileService,
	}
}

func (s *nerdleService) GetTodayGame() (*domain.NerdleGame, error) {
	today := domain.TodayUTC()
	game, err := s.repo.GetGameByDate(today)
	if err == nil {
		return game, nil
	}

	var notFound *domain.GameNotFoundError
	if !errors.As(err, &notFound) {
		return nil, fmt.Errorf("failed to fetch today's game: %w", err)
	}

	formula := domain.GenerateDailyFormula(today)
	game = &domain.NerdleGame{
		ID:      uuid.New(),
		Date:    today,
		Formula: formula,
	}
	if err := s.repo.CreateGame(game); err != nil {
		return nil, fmt.Errorf("failed to create today's game: %w", err)
	}
	return game, nil
}

func (s *nerdleService) GetOrCreateSession(profileID uuid.UUID) (*domain.NerdleSession, error) {
	game, err := s.GetTodayGame()
	if err != nil {
		return nil, err
	}

	session, err := s.repo.GetSessionByProfileAndGame(profileID, game.ID)
	if err == nil {
		attempts, aErr := s.repo.GetAttemptsBySession(session.ID)
		if aErr != nil {
			return nil, fmt.Errorf("failed to load attempts: %w", aErr)
		}
		session.Attempts = attempts
		return session, nil
	}

	now := time.Now().UTC()
	session = &domain.NerdleSession{
		ID:        uuid.New(),
		GameID:    game.ID,
		ProfileID: profileID,
		Solved:    false,
		CreatedAt: now,
		Attempts:  []domain.NerdleAttempt{},
	}
	if err := s.repo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return session, nil
}

func (s *nerdleService) SubmitGuess(profileID uuid.UUID, guess string) (*GuessResponse, error) {
	today := domain.TodayUTC()
	game, err := s.repo.GetGameByDate(today)
	if err != nil {
		return nil, err
	}

	session, err := s.GetOrCreateSession(profileID)
	if err != nil {
		return nil, err
	}

	if session.Solved {
		return nil, &domain.AlreadySolvedError{ProfileID: profileID}
	}

	if len(session.Attempts) >= domain.MaxAttempts {
		return nil, &domain.MaxAttemptsReachedError{ProfileID: profileID}
	}

	if err := domain.ValidateFormula(guess); err != nil {
		return nil, err
	}

	results := domain.EvaluateGuess(guess, game.Formula)
	solved := domain.IsSolved(results)

	now := time.Now().UTC()
	attempt := &domain.NerdleAttempt{
		ID:        uuid.New(),
		SessionID: session.ID,
		Guess:     guess,
		Result:    results,
		CreatedAt: now,
	}
	if err := s.repo.AddAttempt(attempt); err != nil {
		return nil, fmt.Errorf("failed to save attempt: %w", err)
	}

	session.Attempts = append(session.Attempts, *attempt)

	kudosEarned := 0
	if solved {
		session.Solved = true
		session.SolvedAt = &now
		session.CompletedAt = &now

		_, kudos, err := s.profileService.AddKudosMinigame(profileID)
		if err != nil {
			return nil, err
		}
		kudosEarned = kudos
	}

	attemptsUsed := len(session.Attempts)
	gameOver := solved || attemptsUsed >= domain.MaxAttempts
	if gameOver && session.CompletedAt == nil {
		session.CompletedAt = &now
	}

	if err := s.repo.SaveSession(session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	attemptsLeft := domain.MaxAttempts - attemptsUsed

	return &GuessResponse{
		Attempt:      attempt,
		Session:      session,
		Solved:       solved,
		GameOver:     gameOver,
		AttemptsLeft: attemptsLeft,
		KudosEarned:  kudosEarned,
	}, nil
}
