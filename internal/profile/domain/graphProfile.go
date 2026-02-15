package domain

import (
	"encoding/json"

	"github.com/google/uuid"
)

type GraphProfile struct {
	Id                uuid.UUID `json:"id"`
	Name              string    `json:"givenName"`
	Surname           string    `json:"surname"`
	Mail              string    `json:"mail"`
	PreferredLanguage Language  `json:"preferredLanguage"`
	Role              Role      `json:"role"`
}

func (g *GraphProfile) UnmarshalJSON(data []byte) error {
	type Alias GraphProfile
	aux := &struct {
		Id                string `json:"id"`
		PreferredLanguage string `json:"preferredLanguage"`
		JobTitle          string `json:"jobTitle"`
		*Alias
	}{
		Alias: (*Alias)(g),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	parsedUUID, err := uuid.Parse(aux.Id)
	if err != nil {
		return err
	}
	g.Id = parsedUUID

	switch aux.PreferredLanguage[0:2] {
	case "en":
		g.PreferredLanguage = ENG
	case "nl":
		g.PreferredLanguage = NL
	default:
		g.PreferredLanguage = ENG
	}

	switch aux.JobTitle {
	case "Student":
		g.Role = Student
	case "Lector":
		g.Role = Lector
	default:
		g.Role = Student
	}
	return nil
}
