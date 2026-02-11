package domain

import (
	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfileById(id uuid.UUID) (*Profile, error)
	UpdateProfile(profile *Profile) error
}

type Profile struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;"`
	FirstName string
	LastName  string
	Email     string `gorm:"uniqueIndex"`
	Kudos     int
}

func (p *Profile) AddKudos(kudos int) error {
	if kudos < 0 {
		return &NegativeKudosError{Arg: kudos, Message: "Kudos can't be negative"}
	}

	p.Kudos += kudos
	return nil
}
