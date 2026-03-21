package domain

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

const (
	GridSize  = 16
	MineCount = 40
)

type MinesweeperSession struct {
	ProfileID   uuid.UUID     `gorm:"type:uuid;primaryKey" json:"profileId"`
	GameDate    time.Time     `gorm:"primaryKey" json:"gameDate"`
	RevealedMap [][]CellState `gorm:"serializer:json" json:"revealedMap"`
	FlagMap     [][]bool      `gorm:"serializer:json" json:"flagMap"`
	MineBoard   [][]bool      `gorm:"serializer:json" json:"-"`
	BoardReady  bool          `gorm:"default:false" json:"-"`
	MoveCount   int           `gorm:"default:0" json:"moveCount"`
	Solved      bool          `gorm:"default:false" json:"solved"`
	GameOver    bool          `gorm:"default:false" json:"gameOver"`
	SolvedAt    *time.Time    `json:"solvedAt,omitempty"`
	CompletedAt *time.Time    `json:"completedAt,omitempty"`
	CreatedAt   time.Time     `json:"createdAt"`
}

func NewSession(profileID uuid.UUID, gameDate time.Time) *MinesweeperSession {
	now := time.Now().UTC()
	board := make([][]bool, GridSize)
	for i := range board {
		board[i] = make([]bool, GridSize)
	}
	return &MinesweeperSession{
		ProfileID:   profileID,
		GameDate:    gameDate,
		RevealedMap: NewRevealedMap(),
		FlagMap:     NewFlagMap(),
		MineBoard:   board,
		BoardReady:  false,
		MoveCount:   0,
		Solved:      false,
		GameOver:    false,
		CreatedAt:   now,
	}
}

func (s *MinesweeperSession) IsCompleted() bool {
	return s.Solved || s.GameOver
}

func (s *MinesweeperSession) BuildBoardHints() [][]int {
	hints := make([][]int, GridSize)
	for r := 0; r < GridSize; r++ {
		hints[r] = make([]int, GridSize)
		for c := 0; c < GridSize; c++ {
			if s.RevealedMap[r][c] == CellRevealed {
				hints[r][c] = AdjacentMineCount(s.MineBoard, r, c)
			} else {
				hints[r][c] = -1
			}
		}
	}
	return hints
}

type CellState string

const (
	CellHidden   CellState = "hidden"
	CellRevealed CellState = "revealed"
)

type MoveAction string

const (
	ActionReveal MoveAction = "reveal"
	ActionFlag   MoveAction = "flag"
	ActionUnflag MoveAction = "unflag"
	ActionChord  MoveAction = "chord"
)

var ValidActions = map[MoveAction]bool{
	ActionReveal: true,
	ActionFlag:   true,
	ActionUnflag: true,
	ActionChord:  true,
}

type MinesweeperRepository interface {
	GetSessionByProfileAndDate(profileID uuid.UUID, date time.Time) (*MinesweeperSession, error)
	CreateSession(session *MinesweeperSession) error
	SaveSession(session *MinesweeperSession) error
}

func TodayUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func GenerateBoardAvoidingCell(date time.Time, safeRow, safeCol int) [][]bool {
	seed := int64(date.Year())*10000 + int64(date.Month())*100 + int64(date.Day())
	seed ^= int64(safeRow)*100 + int64(safeCol)
	rng := rand.New(rand.NewSource(seed))

	safe := make(map[[2]int]bool)
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			r, c := safeRow+dr, safeCol+dc
			if r >= 0 && r < GridSize && c >= 0 && c < GridSize {
				safe[[2]int{r, c}] = true
			}
		}
	}

	board := make([][]bool, GridSize)
	for i := range board {
		board[i] = make([]bool, GridSize)
	}

	placed := 0
	for placed < MineCount {
		r := rng.Intn(GridSize)
		c := rng.Intn(GridSize)
		if !board[r][c] && !safe[[2]int{r, c}] {
			board[r][c] = true
			placed++
		}
	}
	return board
}

func NewRevealedMap() [][]CellState {
	m := make([][]CellState, GridSize)
	for i := range m {
		m[i] = make([]CellState, GridSize)
		for j := range m[i] {
			m[i][j] = CellHidden
		}
	}
	return m
}

func NewFlagMap() [][]bool {
	m := make([][]bool, GridSize)
	for i := range m {
		m[i] = make([]bool, GridSize)
	}
	return m
}

func AdjacentMineCount(board [][]bool, row, col int) int {
	count := 0
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			r, c := row+dr, col+dc
			if r >= 0 && r < GridSize && c >= 0 && c < GridSize && board[r][c] {
				count++
			}
		}
	}
	return count
}

func AdjacentFlagCount(flags [][]bool, row, col int) int {
	count := 0
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			r, c := row+dr, col+dc
			if r >= 0 && r < GridSize && c >= 0 && c < GridSize && flags[r][c] {
				count++
			}
		}
	}
	return count
}

func FloodReveal(board [][]bool, revealed [][]CellState, row, col int) {
	if row < 0 || row >= GridSize || col < 0 || col >= GridSize {
		return
	}
	if revealed[row][col] == CellRevealed {
		return
	}
	if board[row][col] {
		return
	}
	revealed[row][col] = CellRevealed
	if AdjacentMineCount(board, row, col) == 0 {
		for dr := -1; dr <= 1; dr++ {
			for dc := -1; dc <= 1; dc++ {
				if dr == 0 && dc == 0 {
					continue
				}
				FloodReveal(board, revealed, row+dr, col+dc)
			}
		}
	}
}

func RevealAllMines(board [][]bool, revealed [][]CellState) {
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if board[r][c] {
				revealed[r][c] = CellRevealed
			}
		}
	}
}

func ChordReveal(board [][]bool, revealed [][]CellState, flags [][]bool, row, col int) (anyRevealed bool, hitMine bool) {
	if revealed[row][col] != CellRevealed {
		return false, false
	}
	hint := AdjacentMineCount(board, row, col)
	if hint == 0 {
		return false, false
	}
	if AdjacentFlagCount(flags, row, col) != hint {
		return false, false
	}
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			r, c := row+dr, col+dc
			if r < 0 || r >= GridSize || c < 0 || c >= GridSize {
				continue
			}
			if revealed[r][c] == CellRevealed || flags[r][c] {
				continue
			}
			if board[r][c] {
				hitMine = true
				continue
			}
			FloodReveal(board, revealed, r, c)
			anyRevealed = true
		}
	}
	return anyRevealed, hitMine
}

func IsSolved(board [][]bool, revealed [][]CellState) bool {
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if !board[r][c] && revealed[r][c] != CellRevealed {
				return false
			}
		}
	}
	return true
}
