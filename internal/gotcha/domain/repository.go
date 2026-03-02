package domain

import (
	"time"

	"github.com/google/uuid"
)

type GameRepository interface {
	SaveGame(game *Game) error
	GetGameByCampus(campus string) (*Game, error)
	GetGameByID(id uuid.UUID) (*Game, error)
	GetAllActiveGames() ([]*Game, error)
}

type ParticipantRepository interface {
	SaveParticipant(p *Participant) error
	GetParticipantsByGame(gameID uuid.UUID) ([]*Participant, error)
	GetParticipant(gameID, profileID uuid.UUID) (*Participant, error)
	GetExpiredParticipants(before time.Time) ([]*Participant, error)
}

type KillRepository interface {
	SaveKill(kill *Kill) error
	GetKillByID(id uuid.UUID) (*Kill, error)
	GetKillFeed(gameID uuid.UUID, limit, offset int) ([]*Kill, error)
	GetPendingKills(gameID uuid.UUID) ([]*Kill, error)
	SaveKillLike(like *KillLike) error
	DeleteKillLike(killID, profileID uuid.UUID) error
}

type PropRepository interface {
	GetRandomProp() (*Prop, error)
	SaveProp(prop *Prop) error
	GetAllProps() ([]*Prop, error)
}
