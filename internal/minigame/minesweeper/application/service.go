package application

import (
	"Quest100Backend/internal/minigame/minesweeper/domain"
	profileApp "Quest100Backend/internal/profile/application"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MinesweeperService interface {
	GetOrCreateSession(profileID uuid.UUID) (*domain.MinesweeperSession, error)
	GetSessionResponse(profileID uuid.UUID) (*SessionResponse, error)
	SubmitMove(profileID uuid.UUID, action domain.MoveAction, row, col int) (*MoveResponse, error)
}

type SessionResponse struct {
	*domain.MinesweeperSession
	BoardHints    [][]int  `json:"boardHints"`
	MineLocations [][]bool `json:"mineLocations,omitempty"`
}

type MoveResponse struct {
	Session       *domain.MinesweeperSession `json:"session"`
	Solved        bool                       `json:"solved"`
	GameOver      bool                       `json:"gameOver"`
	HitMine       bool                       `json:"hitMine"`
	KudosEarned   int                        `json:"kudosEarned"`
	BoardHints    [][]int                    `json:"boardHints"`
	MineLocations [][]bool                   `json:"mineLocations,omitempty"`
}

type minesweeperService struct {
	repo           domain.MinesweeperRepository
	profileService profileApp.ProfileService
}

func NewMinesweeperService(repo domain.MinesweeperRepository, profileService profileApp.ProfileService) MinesweeperService {
	return &minesweeperService{
		repo:           repo,
		profileService: profileService,
	}
}

func (s *minesweeperService) GetOrCreateSession(profileID uuid.UUID) (*domain.MinesweeperSession, error) {
	today := domain.TodayUTC()

	session, err := s.repo.GetSessionByProfileAndDate(profileID, today)
	if err == nil {
		return session, nil
	}

	session = domain.NewSession(profileID, today)
	if err := s.repo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return session, nil
}

func (s *minesweeperService) GetSessionResponse(profileID uuid.UUID) (*SessionResponse, error) {
	session, err := s.GetOrCreateSession(profileID)
	if err != nil {
		return nil, err
	}

	resp := &SessionResponse{
		MinesweeperSession: session,
		BoardHints:         session.BuildBoardHints(),
	}
	if session.IsCompleted() && session.BoardReady {
		resp.MineLocations = session.MineBoard
	}
	return resp, nil
}

func (s *minesweeperService) SubmitMove(profileID uuid.UUID, action domain.MoveAction, row, col int) (*MoveResponse, error) {
	session, err := s.GetOrCreateSession(profileID)
	if err != nil {
		return nil, err
	}

	if session.IsCompleted() {
		return nil, &domain.AlreadyCompletedError{ProfileID: profileID}
	}

	if row < 0 || row >= domain.GridSize || col < 0 || col >= domain.GridSize {
		return nil, &domain.InvalidMoveError{
			Message: fmt.Sprintf("coordinates (%d,%d) out of bounds", row, col),
		}
	}

	if !domain.ValidActions[action] {
		return nil, &domain.InvalidMoveError{Message: fmt.Sprintf("unknown action %q", action)}
	}

	hitMine := false
	now := time.Now().UTC()

	switch action {
	case domain.ActionReveal:
		if session.RevealedMap[row][col] == domain.CellRevealed {
			return nil, &domain.InvalidMoveError{Message: "cell is already revealed"}
		}
		if session.FlagMap[row][col] {
			return nil, &domain.InvalidMoveError{Message: "unflag cell before revealing"}
		}
		if !session.BoardReady {
			session.MineBoard = domain.GenerateBoardAvoidingCell(session.GameDate, row, col)
			session.BoardReady = true
		}
		if session.MineBoard[row][col] {
			hitMine = true
			session.GameOver = true
			session.CompletedAt = &now
			domain.RevealAllMines(session.MineBoard, session.RevealedMap)
		} else {
			domain.FloodReveal(session.MineBoard, session.RevealedMap, row, col)
		}

	case domain.ActionChord:
		if session.RevealedMap[row][col] != domain.CellRevealed {
			return nil, &domain.InvalidMoveError{Message: "can only chord a revealed cell"}
		}
		if !session.BoardReady {
			return nil, &domain.InvalidMoveError{Message: "no moves made yet"}
		}
		_, hit := domain.ChordReveal(session.MineBoard, session.RevealedMap, session.FlagMap, row, col)
		if hit {
			hitMine = true
			session.GameOver = true
			session.CompletedAt = &now
			domain.RevealAllMines(session.MineBoard, session.RevealedMap)
		}

	case domain.ActionFlag:
		if session.RevealedMap[row][col] == domain.CellRevealed {
			return nil, &domain.InvalidMoveError{Message: "cannot flag a revealed cell"}
		}
		session.FlagMap[row][col] = true

	case domain.ActionUnflag:
		session.FlagMap[row][col] = false
	}

	session.MoveCount++

	kudosEarned := 0
	solved := false
	if !hitMine && (action == domain.ActionReveal || action == domain.ActionChord) && session.BoardReady {
		solved = domain.IsSolved(session.MineBoard, session.RevealedMap)
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
	}

	if err := s.repo.SaveSession(session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	resp := &MoveResponse{
		Session:     session,
		Solved:      solved,
		GameOver:    session.GameOver,
		HitMine:     hitMine,
		KudosEarned: kudosEarned,
		BoardHints:  session.BuildBoardHints(),
	}
	if session.IsCompleted() {
		resp.MineLocations = session.MineBoard
	}
	return resp, nil
}
