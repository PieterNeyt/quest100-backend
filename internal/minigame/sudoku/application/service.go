package application

import (
	"Quest100Backend/internal/minigame/sudoku/domain"
	profileApp "Quest100Backend/internal/profile/application"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SudokuService interface {
	GetOrCreateSession(profileID uuid.UUID) (*domain.SudokuSession, *domain.SudokuGame, error)
	GetSessionResponse(profileID uuid.UUID) (*SessionResponse, error)
	SubmitMove(profileID uuid.UUID, row, col, value int) (*MoveResponse, error)
	PrepareDailyGame() error
}

type SessionResponse struct {
	*domain.SudokuSession
	Puzzle      [][]int  `json:"puzzle"`
	ConflictMap [][]bool `json:"conflictMap"`
}

type MoveResponse struct {
	Session     *domain.SudokuSession `json:"session"`
	Solved      bool                  `json:"solved"`
	KudosEarned int                   `json:"kudosEarned"`
	ConflictMap [][]bool              `json:"conflictMap"`
}

type sudokuService struct {
	repo           domain.SudokuRepository
	profileService profileApp.ProfileService
}

func NewSudokuService(repo domain.SudokuRepository, profileService profileApp.ProfileService) SudokuService {
	return &sudokuService{
		repo:           repo,
		profileService: profileService,
	}
}

// getOrCreateGame fetches today's game or generates and persists a new one.
func (s *sudokuService) getOrCreateGame() (*domain.SudokuGame, error) {
	today := domain.TodayBrussels()

	game, err := s.repo.GetGameByDate(today)
	if err == nil {
		return game, nil
	}

	var notFound *domain.GameNotFoundError
	if !errors.As(err, &notFound) {
		return nil, fmt.Errorf("failed to fetch today's game: %w", err)
	}

	puzzle, solution := domain.GenerateDailyPuzzle(today)
	game = &domain.SudokuGame{
		ID:       uuid.New(),
		Date:     today,
		Puzzle:   puzzle,
		Solution: solution,
	}
	if err := s.repo.CreateGame(game); err != nil {
		return nil, fmt.Errorf("failed to create today's game: %w", err)
	}
	return game, nil
}

func (s *sudokuService) GetOrCreateSession(profileID uuid.UUID) (*domain.SudokuSession, *domain.SudokuGame, error) {
	game, err := s.getOrCreateGame()
	if err != nil {
		return nil, nil, err
	}

	session, err := s.repo.GetSessionByProfileAndGame(profileID, game.ID)
	if err == nil {
		return session, game, nil
	}

	session = domain.NewSession(profileID, game)
	if err := s.repo.CreateSession(session); err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}
	return session, game, nil
}

func (s *sudokuService) GetSessionResponse(profileID uuid.UUID) (*SessionResponse, error) {
	session, game, err := s.GetOrCreateSession(profileID)
	if err != nil {
		return nil, err
	}

	return &SessionResponse{
		SudokuSession: session,
		Puzzle:        game.Puzzle,
		ConflictMap:   session.ConflictMap(),
	}, nil
}

func (s *sudokuService) SubmitMove(profileID uuid.UUID, row, col, value int) (*MoveResponse, error) {
	session, game, err := s.GetOrCreateSession(profileID)
	if err != nil {
		return nil, err
	}

	if session.IsCompleted() {
		return nil, &domain.AlreadyCompletedError{ProfileID: profileID}
	}

	if err := session.PlaceValue(row, col, value); err != nil {
		return nil, err
	}

	session.MoveCount++

	kudosEarned := 0
	solved := false

	if session.CheckSolved(game.Solution) {
		now := time.Now().UTC()
		solved = true
		session.Solved = true
		session.SolvedAt = &now
		session.CompletedAt = &now

		_, kudos, err := s.profileService.AddKudosMinigame(profileID)
		if err != nil {
			return nil, err
		}
		kudosEarned = kudos
	}

	if err := s.repo.SaveSession(session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &MoveResponse{
		Session:     session,
		Solved:      solved,
		KudosEarned: kudosEarned,
		ConflictMap: session.ConflictMap(),
	}, nil
}

func (s *sudokuService) PrepareDailyGame() error {
	_, err := s.getOrCreateGame()
	return err
}
