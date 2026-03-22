package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

const MaxAttempts = 6

type NerdleGame struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	Date    time.Time `gorm:"uniqueIndex;not null" json:"date"`
	Formula string    `gorm:"type:text;not null" json:"-"`
}

type NerdleSession struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;" json:"id"`
	GameID      uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_nerdle_session_game_profile" json:"gameId"`
	ProfileID   uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_nerdle_session_game_profile" json:"profileId"`
	Attempts    []NerdleAttempt `gorm:"foreignKey:SessionID" json:"attempts"`
	Solved      bool            `gorm:"default:false" json:"solved"`
	SolvedAt    *time.Time      `json:"solvedAt,omitempty"`
	CompletedAt *time.Time      `json:"completedAt,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
}

func NewSession(profileID uuid.UUID, gameID uuid.UUID) *NerdleSession {
	return &NerdleSession{
		ID:        uuid.New(),
		GameID:    gameID,
		ProfileID: profileID,
		Solved:    false,
		CreatedAt: time.Now().UTC(),
		Attempts:  []NerdleAttempt{},
	}
}

func (s *NerdleSession) IsCompleted() bool {
	return s.Solved || len(s.Attempts) >= MaxAttempts
}

func (s *NerdleSession) AttemptsLeft() int {
	left := MaxAttempts - len(s.Attempts)
	if left < 0 {
		return 0
	}
	return left
}

type NerdleAttempt struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey;" json:"id"`
	SessionID uuid.UUID    `gorm:"type:uuid;not null;index" json:"sessionId"`
	Guess     string       `gorm:"type:varchar(32);not null" json:"guess"`
	Result    []TileResult `gorm:"serializer:json" json:"result"`
	CreatedAt time.Time    `json:"createdAt"`
}

type TileResult struct {
	Char   string     `json:"char"`
	Status TileStatus `json:"status"`
}

type TileStatus string

const (
	TileCorrect TileStatus = "correct"
	TilePresent TileStatus = "present"
	TileAbsent  TileStatus = "absent"
)

type NerdleRepository interface {
	GetGameByDate(date time.Time) (*NerdleGame, error)
	CreateGame(game *NerdleGame) error

	GetSessionByProfileAndGame(profileID uuid.UUID, gameID uuid.UUID) (*NerdleSession, error)
	CreateSession(session *NerdleSession) error
	SaveSession(session *NerdleSession) error

	AddAttempt(attempt *NerdleAttempt) error
	GetAttemptsBySession(sessionID uuid.UUID) ([]NerdleAttempt, error)
}

func TodayBrussels() time.Time {
	loc, _ := time.LoadLocation("Europe/Brussels")
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func ValidateFormula(guess string) error {
	if guess == "" {
		return &InvalidGuessError{Guess: guess, Message: "guess cannot be empty"}
	}

	allowed := "0123456789+-*/="
	for _, ch := range guess {
		if !strings.ContainsRune(allowed, ch) {
			return &InvalidGuessError{Guess: guess, Message: fmt.Sprintf("invalid character '%c'", ch)}
		}
	}

	parts := strings.SplitN(guess, "=", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return &InvalidGuessError{Guess: guess, Message: "must contain exactly one '=' with non-empty sides"}
	}

	rVal, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return &InvalidGuessError{Guess: guess, Message: "right-hand side must be a plain integer"}
	}

	lVal, err := evalExpression(parts[0])
	if err != nil {
		return &InvalidGuessError{Guess: guess, Message: err.Error()}
	}

	if lVal != rVal {
		return &InvalidGuessError{
			Guess:   guess,
			Message: fmt.Sprintf("left side evaluates to %g, not %s", lVal, parts[1]),
		}
	}

	return nil
}

func EvaluateGuess(guess, formula string) []TileResult {
	g := []rune(guess)
	f := []rune(formula)
	n := len(f)

	results := make([]TileResult, n)
	fUsed := make([]bool, n)
	gMatched := make([]bool, n)

	for i := 0; i < n; i++ {
		if i < len(g) && g[i] == f[i] {
			results[i] = TileResult{Char: string(g[i]), Status: TileCorrect}
			fUsed[i] = true
			gMatched[i] = true
		}
	}

	for i := 0; i < n; i++ {
		if gMatched[i] || i >= len(g) {
			if i >= len(g) {
				results[i] = TileResult{Char: " ", Status: TileAbsent}
			}
			continue
		}
		found := false
		for j := 0; j < n; j++ {
			if !fUsed[j] && g[i] == f[j] {
				results[i] = TileResult{Char: string(g[i]), Status: TilePresent}
				fUsed[j] = true
				found = true
				break
			}
		}
		if !found {
			results[i] = TileResult{Char: string(g[i]), Status: TileAbsent}
		}
	}

	return results
}

func IsSolved(results []TileResult) bool {
	if len(results) == 0 {
		return false
	}
	for _, r := range results {
		if r.Status != TileCorrect {
			return false
		}
	}
	return true
}

type exprParser struct {
	input string
	pos   int
}

func evalExpression(expr string) (float64, error) {
	p := &exprParser{input: expr}
	val, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	if p.pos != len(p.input) {
		return 0, fmt.Errorf("unexpected token near position %d", p.pos)
	}
	return val, nil
}

func (p *exprParser) peek() (rune, bool) {
	if p.pos >= len(p.input) {
		return 0, false
	}
	return rune(p.input[p.pos]), true
}

func (p *exprParser) consume() rune {
	r := rune(p.input[p.pos])
	p.pos++
	return r
}

func (p *exprParser) parseExpr() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		ch, ok := p.peek()
		if !ok || (ch != '+' && ch != '-') {
			break
		}
		op := p.consume()
		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if op == '+' {
			left += right
		} else {
			left -= right
		}
	}
	return left, nil
}

func (p *exprParser) parseTerm() (float64, error) {
	left, err := p.parseNumber()
	if err != nil {
		return 0, err
	}
	for {
		ch, ok := p.peek()
		if !ok || (ch != '*' && ch != '/') {
			break
		}
		op := p.consume()
		right, err := p.parseNumber()
		if err != nil {
			return 0, err
		}
		if op == '*' {
			left *= right
		} else {
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left /= right
		}
	}
	return left, nil
}

func (p *exprParser) parseNumber() (float64, error) {
	start := p.pos
	for {
		ch, ok := p.peek()
		if !ok || !unicode.IsDigit(ch) {
			break
		}
		p.consume()
	}
	if p.pos == start {
		ch, _ := p.peek()
		return 0, fmt.Errorf("expected number, got '%c'", ch)
	}
	s := p.input[start:p.pos]
	if len(s) > 1 && s[0] == '0' {
		return 0, fmt.Errorf("leading zeros not allowed in '%s'", s)
	}
	return strconv.ParseFloat(s, 64)
}
