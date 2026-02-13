package domain

import (
	"encoding/json"

	"github.com/google/uuid"
)

type GraphProfile struct {
	Id      uuid.UUID `json:"id"`
	Name    string    `json:"givenName"`
	Surname string    `json:"surname"`
	Mail    string    `json:"mail"`
}

func (g *GraphProfile) UnmarshalJSON(data []byte) error {
	type Alias GraphProfile
	aux := &struct {
		Id string `json:"id"`
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
	return nil
}
