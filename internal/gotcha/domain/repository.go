package domain

import (
	"time"

	"github.com/google/uuid"
)

type GameRepository interface {
	SaveGame(game *GotchaGame) error
	GetGameByCampus(campus string) (*GotchaGame, error)
	GetGameByID(id uuid.UUID) (*GotchaGame, error)
	GetAllActiveGames() ([]*GotchaGame, error)
	GetAllOptInGames() ([]*GotchaGame, error)
}

type ParticipantRepository interface {
	SaveParticipant(p *Participant) error
	GetParticipantsByGame(gameID uuid.UUID) ([]*Participant, error)
	GetParticipant(gameID, profileID uuid.UUID) (*Participant, error)
	GetExpiredParticipants(before time.Time) ([]*Participant, error)
	DeleteParticipant(gameID, profileID uuid.UUID) error
}

type KillRepository interface {
	SaveKill(kill *GotchaKill) error
	GetKillByID(id uuid.UUID) (*GotchaKill, error)
	GetKillFeed(gameID uuid.UUID, limit, offset int) ([]*GotchaKill, error)
	GetPendingKills(gameID uuid.UUID) ([]*GotchaKill, error)
	SaveKillLike(like *GotchaKillLike) error
	DeleteKillLike(killID, profileID uuid.UUID) error
	HasLiked(killID, profileID uuid.UUID) (bool, error)
}

type PropRepository interface {
	GetRandomProp() (*GotchaProp, error)
	GetPropByID(id uuid.UUID) (*GotchaProp, error)
	SaveProp(prop *GotchaProp) error
	GetAllProps() ([]*GotchaProp, error)
}
