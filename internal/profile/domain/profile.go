package domain

import (
	"log"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfileById(id uuid.UUID) (*Profile, error)
	UpdateProfile(profile *Profile) error
}

type Profile struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;"`
	FirstName    string
	LastName     string
	Email        string `gorm:"uniqueIndex"`
	Kudos        int
	ArchetypeID  int
	PlayerStats  PlayerStats  `gorm:"foreignKey:ProfileID;references:ID"`
	KudosHistory []KudosEntry `gorm:"foreignKey:ProfileID;references:ID"`
}

type PlayerStats struct {
	ProfileID      uuid.UUID `gorm:"type:uuid;primaryKey;"`
	KudoKnowledge  int
	KudoAttendance int
	KudoTeamwork   int
	KudoAtmosphere int
	KudoEngagement int
}

func (p *Profile) AddKudos(kudos int, reason string, kudoType KudoType) error {
	if kudos < 0 {
		return &NegativeKudosError{Arg: kudos, Message: "Kudos can't be negative"}
	}

	p.Kudos += kudos
	p.KudosHistory = append(p.KudosHistory, KudosEntry{
		ID:        uuid.New(),
		ProfileID: p.ID,
		Amount:    kudos,
		Reason:    reason,
		Type:      kudoType,
	})

	p.UpdatePlayerStats(kudos, kudoType)
	return nil
}

func (p *Profile) UpdatePlayerStats(amount int, kudoType KudoType) {
	switch kudoType {
	case KudoKnowledge:
		p.PlayerStats.KudoKnowledge += amount
	case KudoAttendance:
		p.PlayerStats.KudoAttendance += amount
	case KudoTeamwork:
		p.PlayerStats.KudoTeamwork += amount
	case KudoAtmosphere:
		p.PlayerStats.KudoAtmosphere += amount
	case KudoEngagement:
		p.PlayerStats.KudoEngagement += amount
	}

	if err := p.CalculateArcheType(); err != nil {
		log.Printf("Error calculating archetype: %v", err)
	}
}
