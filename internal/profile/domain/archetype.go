package domain

import (
	"github.com/google/uuid"
)

type Archetype struct {
	ID          int
	Name        string
	Description string
}

const (
	Wizard = iota
	TeamCatalyst
	AtmosphereMaker
	CampusExplorer
	AcademicGuardian
)

func (p *Profile) CalculateArcheType() error {
	if p.PlayerStats.ProfileID == uuid.Nil {
		return &NoPlayerStatsError{
			Message: "Player stats are required to calculate archetype",
		}
	}

	stats := p.PlayerStats

	maxKudo := stats.KudoKnowledge
	selected := Wizard

	if stats.KudoAttendance > maxKudo {
		maxKudo = stats.KudoAttendance
		selected = AcademicGuardian
	}

	if stats.KudoTeamwork > maxKudo {
		maxKudo = stats.KudoTeamwork
		selected = TeamCatalyst
	}

	if stats.KudoAtmosphere > maxKudo {
		maxKudo = stats.KudoAtmosphere
		selected = AtmosphereMaker
	}

	if stats.KudoEngagement > maxKudo {
		maxKudo = stats.KudoEngagement
		selected = CampusExplorer
	}

	p.ArchetypeID = selected
	return nil
}

func (p *Profile) GetArchetype() Archetype {
	return Archetypes[p.ArchetypeID]
}

var Archetypes = map[int]Archetype{
	Wizard: {
		ID:          Wizard,
		Name:        "Wizard",
		Description: "Academisch sterke student die inhoudelijke vragen stelt, actief participeert en vaak kudos ontvangt van docenten en peers.",
	},
	TeamCatalyst: {
		ID:          TeamCatalyst,
		Name:        "Team Catalyst",
		Description: "Verbindende teamspeler die groepsdynamiek versterkt en vaak peer-to-peer kudos ontvangt voor samenwerking.",
	},
	AtmosphereMaker: {
		ID:          AtmosphereMaker,
		Name:        "Atmosphere Maker",
		Description: "Sociale energiebron die sfeer brengt op campus en betrokkenheid stimuleert via events en interacties.",
	},
	CampusExplorer: {
		ID:          CampusExplorer,
		Name:        "Campus Explorer",
		Description: "Avontuurlijke student die quests voltooit, collectibles verzamelt en actief deelneemt aan campus games.",
	},
	AcademicGuardian: {
		ID:          AcademicGuardian,
		Name:        "Academic Guardian",
		Description: "Consistente en punctuele student die uitblinkt in aanwezigheid en het tijdig indienen van opdrachten.",
	},
}
