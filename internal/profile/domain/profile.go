package domain

import (
	"log"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfileById(id uuid.UUID) (*Profile, error)
	UpdateProfile(profile *Profile) error
	SaveProfile(profile *Profile) error
}

type Language string

const (
	NL  Language = "NL"
	ENG          = "EN"
)

type Profile struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	FirstName         string    `json:"firstName"`
	LastName          string    `json:"lastName"`
	Email             string    `gorm:"uniqueIndex" json:"email"`
	Kudos             int       `json:"kudos"`
	ArchetypeID       int
	PrefferedLanguage Language     `gorm:"type:varchar(2);check:preffered_language IN ('NL','EN')"`
	PlayerStats       PlayerStats  `gorm:"foreignKey:ProfileID;references:ID"`
	KudosHistory      []KudosEntry `gorm:"foreignKey:ProfileID;references:ID"`
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

func (p *Profile) Sync(graph *GraphProfile) error {
	if p.ID != graph.Id {
		return &ProfileIdError{Arg: graph.Id, Message: "Profile id doesn't match"}
	}

	p.FirstName = graph.Name
	p.LastName = graph.Surname
	p.Email = graph.Mail
	return nil
}

func CreateProfile(graph *GraphProfile) *Profile {
	return &Profile{
		ID:                graph.Id,
		FirstName:         graph.Name,
		LastName:          graph.Surname,
		Email:             graph.Mail,
		Kudos:             0,
		ArchetypeID:       1,
		PlayerStats:       PlayerStats{},
		KudosHistory:      []KudosEntry{},
		PrefferedLanguage: graph.PreferredLanguage,
	}
}
