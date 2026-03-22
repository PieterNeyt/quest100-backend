package domain

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type SudokuRepository interface {
	GetGameByDate(date time.Time) (*SudokuGame, error)
	CreateGame(game *SudokuGame) error

	GetSessionByProfileAndGame(profileID uuid.UUID, gameID uuid.UUID) (*SudokuSession, error)
	CreateSession(session *SudokuSession) error
	SaveSession(session *SudokuSession) error
}

const GridSize = 9

type SudokuGame struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Date     time.Time `gorm:"uniqueIndex;not null" json:"date"`
	Puzzle   [][]int   `gorm:"serializer:json" json:"-"`
	Solution [][]int   `gorm:"serializer:json" json:"-"`
}

type SudokuSession struct {
	GameID      uuid.UUID     `gorm:"type:uuid;primaryKey" json:"gameId"`
	ProfileID   uuid.UUID     `gorm:"type:uuid;primaryKey" json:"profileId"`
	PlayerBoard [][]int       `gorm:"serializer:json" json:"playerBoard"`
	CellStates  [][]CellState `gorm:"serializer:json" json:"cellStates"`
	MoveCount   int           `gorm:"default:0" json:"moveCount"`
	Solved      bool          `gorm:"default:false" json:"solved"`
	SolvedAt    *time.Time    `json:"solvedAt,omitempty"`
	CompletedAt *time.Time    `json:"completedAt,omitempty"`
	CreatedAt   time.Time     `json:"createdAt"`
}

type CellState string

const (
	CellGiven  CellState = "given"
	CellEmpty  CellState = "empty"
	CellFilled CellState = "filled"
)

func NewSession(profileID uuid.UUID, game *SudokuGame) *SudokuSession {
	playerBoard := make([][]int, GridSize)
	cellStates := make([][]CellState, GridSize)
	for r := 0; r < GridSize; r++ {
		playerBoard[r] = make([]int, GridSize)
		cellStates[r] = make([]CellState, GridSize)
		for c := 0; c < GridSize; c++ {
			playerBoard[r][c] = game.Puzzle[r][c]
			if game.Puzzle[r][c] != 0 {
				cellStates[r][c] = CellGiven
			} else {
				cellStates[r][c] = CellEmpty
			}
		}
	}
	return &SudokuSession{
		GameID:      game.ID,
		ProfileID:   profileID,
		PlayerBoard: playerBoard,
		CellStates:  cellStates,
		MoveCount:   0,
		Solved:      false,
		CreatedAt:   time.Now().UTC(),
	}
}

func (s *SudokuSession) IsCompleted() bool {
	return s.Solved
}

func (s *SudokuSession) PlaceValue(row, col, value int) error {
	if row < 0 || row >= GridSize || col < 0 || col >= GridSize {
		return &InvalidMoveError{Message: "coordinates out of bounds"}
	}
	if value < 0 || value > 9 {
		return &InvalidMoveError{Message: "value must be between 0 and 9"}
	}
	if s.CellStates[row][col] == CellGiven {
		return &InvalidMoveError{Message: "cannot modify a given cell"}
	}
	s.PlayerBoard[row][col] = value
	if value == 0 {
		s.CellStates[row][col] = CellEmpty
	} else {
		s.CellStates[row][col] = CellFilled
	}
	return nil
}

func (s *SudokuSession) CheckSolved(solution [][]int) bool {
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if s.PlayerBoard[r][c] != solution[r][c] {
				return false
			}
		}
	}
	return true
}

func (s *SudokuSession) ConflictMap() [][]bool {
	conflicts := make([][]bool, GridSize)
	for i := range conflicts {
		conflicts[i] = make([]bool, GridSize)
	}
	// Rows
	for r := 0; r < GridSize; r++ {
		seen := map[int][]int{}
		for c := 0; c < GridSize; c++ {
			if v := s.PlayerBoard[r][c]; v != 0 {
				seen[v] = append(seen[v], c)
			}
		}
		for _, cols := range seen {
			if len(cols) > 1 {
				for _, c := range cols {
					conflicts[r][c] = true
				}
			}
		}
	}
	// Columns
	for c := 0; c < GridSize; c++ {
		seen := map[int][]int{}
		for r := 0; r < GridSize; r++ {
			if v := s.PlayerBoard[r][c]; v != 0 {
				seen[v] = append(seen[v], r)
			}
		}
		for _, rows := range seen {
			if len(rows) > 1 {
				for _, r := range rows {
					conflicts[r][c] = true
				}
			}
		}
	}
	// 3×3 boxes
	for boxR := 0; boxR < 3; boxR++ {
		for boxC := 0; boxC < 3; boxC++ {
			seen := map[int][][2]int{}
			for dr := 0; dr < 3; dr++ {
				for dc := 0; dc < 3; dc++ {
					r, c := boxR*3+dr, boxC*3+dc
					if v := s.PlayerBoard[r][c]; v != 0 {
						seen[v] = append(seen[v], [2]int{r, c})
					}
				}
			}
			for _, cells := range seen {
				if len(cells) > 1 {
					for _, pos := range cells {
						conflicts[pos[0]][pos[1]] = true
					}
				}
			}
		}
	}
	return conflicts
}

func TodayBrussels() time.Time {
	loc, _ := time.LoadLocation("Europe/Brussels")
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func GenerateDailyPuzzle(date time.Time) (puzzle [][]int, solution [][]int) {
	seed := int64(date.Year())*10000 + int64(date.Month())*100 + int64(date.Day())
	rng := rand.New(rand.NewSource(seed))
	solution = generateFullBoard(rng)
	puzzle = digHoles(solution, rng)
	return puzzle, solution
}

func generateFullBoard(rng *rand.Rand) [][]int {
	board := make([][]int, GridSize)
	for i := range board {
		board[i] = make([]int, GridSize)
	}
	fillBoard(board, rng)
	return board
}

func fillBoard(board [][]int, rng *rand.Rand) bool {
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if board[r][c] == 0 {
				for _, d := range shuffled(rng) {
					if isValid(board, r, c, d) {
						board[r][c] = d
						if fillBoard(board, rng) {
							return true
						}
						board[r][c] = 0
					}
				}
				return false
			}
		}
	}
	return true
}

func shuffled(rng *rand.Rand) []int {
	digits := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	rng.Shuffle(len(digits), func(i, j int) { digits[i], digits[j] = digits[j], digits[i] })
	return digits
}

func isValid(board [][]int, row, col, val int) bool {
	for c := 0; c < GridSize; c++ {
		if board[row][c] == val {
			return false
		}
	}
	for r := 0; r < GridSize; r++ {
		if board[r][col] == val {
			return false
		}
	}
	boxR, boxC := (row/3)*3, (col/3)*3
	for dr := 0; dr < 3; dr++ {
		for dc := 0; dc < 3; dc++ {
			if board[boxR+dr][boxC+dc] == val {
				return false
			}
		}
	}
	return true
}

const targetGivens = 45

func digHoles(solution [][]int, rng *rand.Rand) [][]int {
	puzzle := copyBoard(solution)
	positions := rng.Perm(GridSize * GridSize)
	removed := 0
	goal := GridSize*GridSize - targetGivens

	for _, pos := range positions {
		if removed >= goal {
			break
		}
		r, c := pos/GridSize, pos%GridSize
		if puzzle[r][c] == 0 {
			continue
		}
		backup := puzzle[r][c]
		puzzle[r][c] = 0
		if !hasUniqueSolution(puzzle) {
			puzzle[r][c] = backup
		} else {
			removed++
		}
	}
	return puzzle
}

func copyBoard(src [][]int) [][]int {
	dst := make([][]int, GridSize)
	for i := range src {
		dst[i] = make([]int, GridSize)
		copy(dst[i], src[i])
	}
	return dst
}

func hasUniqueSolution(puzzle [][]int) bool {
	board := copyBoard(puzzle)
	count := 0
	countSolutions(board, &count)
	return count == 1
}

func countSolutions(board [][]int, count *int) {
	if *count > 1 {
		return
	}
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if board[r][c] == 0 {
				for d := 1; d <= 9; d++ {
					if isValid(board, r, c, d) {
						board[r][c] = d
						countSolutions(board, count)
						board[r][c] = 0
					}
				}
				return
			}
		}
	}
	*count++
}
