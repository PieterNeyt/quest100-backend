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
	GetOrCreateSession(profileID uuid.UUID) (*domain.NerdleSession, error)
	SubmitGuess(profileID uuid.UUID, guess string) (*GuessResponse, error)
	PrepareDailyGame() error
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

func (s *nerdleService) getOrCreateGame() (*domain.NerdleGame, error) {
	today := domain.TodayUTC()
	game, err := s.repo.GetGameByDate(today)
	if err == nil {
		return game, nil
	}

	var notFound *domain.GameNotFoundError
	if !errors.As(err, &notFound) {
		return nil, fmt.Errorf("failed to fetch today's game: %w", err)
	}

	game = &domain.NerdleGame{
		ID:      uuid.New(),
		Date:    today,
		Formula: domain.GenerateDailyFormula(today),
	}
	if err := s.repo.CreateGame(game); err != nil {
		return nil, fmt.Errorf("failed to create today's game: %w", err)
	}
	return game, nil
}

func (s *nerdleService) getOrCreateSessionForGame(profileID uuid.UUID, game *domain.NerdleGame) (*domain.NerdleSession, error) {
	session, err := s.repo.GetSessionByProfileAndGame(profileID, game.ID)
	if err == nil {
		attempts, aErr := s.repo.GetAttemptsBySession(session.ID)
		if aErr != nil {
			return nil, fmt.Errorf("failed to load attempts: %w", aErr)
		}
		session.Attempts = attempts
		return session, nil
	}

	session = domain.NewSession(profileID, game.ID)
	if err := s.repo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return session, nil
}

func (s *nerdleService) GetOrCreateSession(profileID uuid.UUID) (*domain.NerdleSession, error) {
	game, err := s.getOrCreateGame()
	if err != nil {
		return nil, err
	}
	return s.getOrCreateSessionForGame(profileID, game)
}

func (s *nerdleService) SubmitGuess(profileID uuid.UUID, guess string) (*GuessResponse, error) {
	game, err := s.getOrCreateGame()
	if err != nil {
		return nil, err
	}

	session, err := s.getOrCreateSessionForGame(profileID, game)
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

	gameOver := session.IsCompleted()
	if gameOver && session.CompletedAt == nil {
		session.CompletedAt = &now
	}

	if err := s.repo.SaveSession(session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &GuessResponse{
		Attempt:      attempt,
		Session:      session,
		Solved:       solved,
		GameOver:     gameOver,
		AttemptsLeft: session.AttemptsLeft(),
		KudosEarned:  kudosEarned,
	}, nil
}

func (s *nerdleService) PrepareDailyGame() error {
	_, err := s.getOrCreateGame()
	return err
}
