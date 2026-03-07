package domain

import "github.com/google/uuid"

type GameRepository interface {
	SaveGame(game *GotchaGame) error
	GetGameByCampus(campus string) (*GotchaGame, error)
	GetGameByFinishedCampus(campus string) (*GotchaGame, error)
	GetGameByID(id uuid.UUID) (*GotchaGame, error)
	GetAllActiveGames() ([]*GotchaGame, error)
	GetAllOptInGames() ([]*GotchaGame, error)
}

type ParticipantRepository interface {
	SaveParticipant(p *Participant) error
	GetParticipant(gameID, profileID uuid.UUID) (*Participant, error)
	GetParticipantsByGame(gameID uuid.UUID) ([]*Participant, error)
	GetExpiredParticipants(before interface{}) ([]*Participant, error)
	DeleteParticipant(gameID, profileID uuid.UUID) error
}

type KillRepository interface {
	SaveKill(kill *GotchaKill) error
	GetKillByID(id uuid.UUID) (*GotchaKill, error)
	GetKillFeed(gameID uuid.UUID, limit, offset int) ([]*GotchaKill, error)
	GetApprovedKillFeed(gameID uuid.UUID, limit, offset int) ([]*GotchaKill, error)
	GetApprovedKillsByGame(gameID uuid.UUID) ([]*GotchaKill, error)
	GetPendingKills(gameID uuid.UUID) ([]*GotchaKill, error)

	GetOldestPendingKill(gameID uuid.UUID) (*GotchaKill, error)
	HasPendingKill(gameID, hunterID uuid.UUID) (bool, error)
	CountPendingKills(gameID uuid.UUID) (int, error)

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
