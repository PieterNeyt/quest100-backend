package domain

import (
	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfileById(id uuid.UUID) (*Profile, error)
	UpdateProfile(profile *Profile) error
	SaveProfile(profile *Profile) error
}

type Profile struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `gorm:"uniqueIndex" json:"email"`
	Kudos     int       `json:"kudos"`
}

func (p *Profile) AddKudos(kudos int) error {
	if kudos < 0 {
		return &NegativeKudosError{Arg: kudos, Message: "Kudos can't be negative"}
	}

	p.Kudos += kudos
	return nil
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
		ID:        graph.Id,
		FirstName: graph.Name,
		LastName:  graph.Surname,
		Email:     graph.Mail,
		Kudos:     0,
	}
}
