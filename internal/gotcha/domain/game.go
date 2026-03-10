package domain

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type GameStatus string

const (
	StatusOptIn    GameStatus = "OPT_IN"
	StatusActive   GameStatus = "ACTIVE"
	StatusFinished GameStatus = "FINISHED"
)

type GotchaGame struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Campus            string     `gorm:"type:varchar(100);index" json:"campus"`
	Status            GameStatus `gorm:"type:varchar(20)" json:"status"`
	StartDate         time.Time  `json:"startDate"`
	KillDeadlineHours int        `gorm:"default:72" json:"killDeadlineHours"`
	WinnerID          *uuid.UUID `gorm:"type:uuid" json:"winnerId,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`

	PrizePhotoBase64   string `gorm:"type:text" json:"prizePhotoBase64,omitempty"`
	PrizeDescriptionEN string `gorm:"type:text" json:"prizeDescriptionEN,omitempty"`
	PrizeDescriptionNL string `gorm:"type:text" json:"prizeDescriptionNL,omitempty"`

	Participants []Participant `gorm:"foreignKey:GameID" json:"participants,omitempty"`
}

func (g *GotchaGame) AssignTargets() error {
	active := g.activePlayers()
	if len(active) < 2 {
		return fmt.Errorf("need at least 2 participants to start")
	}
	rand.Shuffle(len(active), func(i, j int) { active[i], active[j] = active[j], active[i] })
	for i := 0; i < len(active)-1; i++ {
		targetID := active[i+1].ProfileID
		active[i].TargetID = &targetID
		active[i].KillDeadline = time.Now().Add(time.Duration(g.KillDeadlineHours) * time.Hour)
	}
	firstID := active[0].ProfileID
	active[len(active)-1].TargetID = &firstID
	active[len(active)-1].KillDeadline = time.Now().Add(time.Duration(g.KillDeadlineHours) * time.Hour)
	g.Status = StatusActive
	g.UpdatedAt = time.Now()
	return nil
}

func (g *GotchaGame) ProcessKill(hunterID, victimID uuid.UUID) error {
	hunter := g.FindParticipant(hunterID)
	victim := g.FindParticipant(victimID)
	if hunter == nil || victim == nil {
		return fmt.Errorf("hunter or victim not found")
	}
	if hunter.TargetID == nil || *hunter.TargetID != victimID {
		return fmt.Errorf("victim is not hunter's current target")
	}

	victim.IsAlive = false
	victim.KilledAt = timePtr(time.Now())
	victim.KilledBy = &hunterID

	hunter.TargetID = victim.TargetID
	if hunter.TargetID != nil {
		hunter.KillDeadline = time.Now().Add(time.Duration(g.KillDeadlineHours) * time.Hour)
	}

	g.UpdatedAt = time.Now()
	g.checkWinner()
	return nil
}

func (g *GotchaGame) ProcessTimeout(victimID uuid.UUID) error {
	victim := g.FindParticipant(victimID)
	if victim == nil || !victim.IsAlive {
		return nil
	}
	victim.IsAlive = false
	victim.KilledAt = timePtr(time.Now())

	hunter := g.findHunterOf(victimID)
	if hunter != nil {
		hunter.TargetID = victim.TargetID
		if hunter.TargetID != nil {
			hunter.KillDeadline = time.Now().Add(time.Duration(g.KillDeadlineHours) * time.Hour)
		}
	}
	g.UpdatedAt = time.Now()
	g.checkWinner()
	return nil
}

func (g *GotchaGame) checkWinner() {
	alive := g.activePlayers()
	if len(alive) == 1 {
		g.Status = StatusFinished
		g.WinnerID = &alive[0].ProfileID
	}
}

func (g *GotchaGame) activePlayers() []*Participant {
	var result []*Participant
	for i := range g.Participants {
		if g.Participants[i].IsAlive {
			result = append(result, &g.Participants[i])
		}
	}
	return result
}

func (g *GotchaGame) ValidateKillApproval(hunterID, victimID uuid.UUID) error {
	hunter := g.FindParticipant(hunterID)
	if hunter == nil || !hunter.IsAlive {
		return fmt.Errorf("hunter is not alive or not found")
	}
	if hunter.TargetID == nil || *hunter.TargetID != victimID {
		return fmt.Errorf("victim is not hunter's current target")
	}
	return nil
}

func (g *GotchaGame) FindParticipant(id uuid.UUID) *Participant {
	for i := range g.Participants {
		if g.Participants[i].ProfileID == id {
			return &g.Participants[i]
		}
	}
	return nil
}

func (g *GotchaGame) findHunterOf(targetID uuid.UUID) *Participant {
	for i := range g.Participants {
		p := &g.Participants[i]
		if p.IsAlive && p.TargetID != nil && *p.TargetID == targetID {
			return p
		}
	}
	return nil
}

func timePtr(t time.Time) *time.Time { return &t }
