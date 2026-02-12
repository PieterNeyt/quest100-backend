package domain

import (
	"time"

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
	KudosHistory []KudosEntry `gorm:"foreignKey:ProfileID;references:ID"`
}

type KudosEntry struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;"`
	ProfileID uuid.UUID `gorm:"type:uuid;index;"`
	Amount    int
	Reason    string
	Date      time.Time `gorm:"autoCreateTime"`
}

func (p *Profile) AddKudos(kudos int, reason string) error {
	if kudos < 0 {
		return &NegativeKudosError{Arg: kudos, Message: "Kudos can't be negative"}
	}

	p.Kudos += kudos
	p.KudosHistory = append(p.KudosHistory, KudosEntry{
		ID:        uuid.New(),
		ProfileID: p.ID,
		Amount:    kudos,
		Reason:    reason,
	})
	return nil
}
