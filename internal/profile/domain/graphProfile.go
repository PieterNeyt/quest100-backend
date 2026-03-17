package domain

import (
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
)

type GraphProfile struct {
	Id                uuid.UUID `json:"id"`
	EmployeeID        int       `json:"employeeId"`
	Name              string    `json:"givenName"`
	Surname           string    `json:"surname"`
	Mail              string    `json:"mail"`
	PreferredLanguage Language  `json:"preferredLanguage"`
	OfficeLocation    string    `json:"officeLocation"`
}

func (g *GraphProfile) UnmarshalJSON(data []byte) error {
	type Alias GraphProfile
	aux := &struct {
		Id                string `json:"id"`
		EmployeeID        string `json:"employeeId"`
		PreferredLanguage string `json:"preferredLanguage"`
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

	parsedEmployeeID, err := strconv.Atoi(aux.EmployeeID)
	g.EmployeeID = parsedEmployeeID

	switch aux.PreferredLanguage[0:2] {
	case "en":
		g.PreferredLanguage = ENG
	case "nl":
		g.PreferredLanguage = NL
	default:
		g.PreferredLanguage = ENG
	}
	return nil
}
