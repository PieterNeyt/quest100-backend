package domain

import (
	"github.com/google/uuid"
)

type Archetype struct {
	ID           int
	Name         string
	Descriptions map[Language]string
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
	archetype := Archetypes[p.ArchetypeID]

	return Archetype{
		ID:   archetype.ID,
		Name: archetype.Name,
		Descriptions: map[Language]string{
			p.PreferredLanguage: archetype.Descriptions[p.PreferredLanguage],
		},
	}
}

var Archetypes = map[int]Archetype{
	Wizard: {
		ID:   Wizard,
		Name: "Wizard",
		Descriptions: map[Language]string{
			NL:  "Academisch sterke student die inhoudelijke vragen stelt, actief participeert en vaak kudos ontvangt van docenten en peers.",
			ENG: "Academically strong student asking insightful questions, actively participating and frequently receiving kudos from teachers and peers.",
		},
	},
	TeamCatalyst: {
		ID:   TeamCatalyst,
		Name: "Team Catalyst",
		Descriptions: map[Language]string{
			NL:  "Verbindende teamspeler die groepsdynamiek versterkt en vaak peer-to-peer kudos ontvangt voor samenwerking.",
			ENG: "Connecting team player who strengthens group dynamics and frequently receives peer-to-peer kudos for collaboration.",
		},
	},
	AtmosphereMaker: {
		ID:   AtmosphereMaker,
		Name: "Atmosphere Maker",
		Descriptions: map[Language]string{
			NL:  "Sociale energiebron die sfeer brengt op campus en betrokkenheid stimuleert via events en interacties.",
			ENG: "Social energy source who brings atmosphere to campus and encourages engagement through events and interactions.",
		},
	},
	CampusExplorer: {
		ID:   CampusExplorer,
		Name: "Campus Explorer",
		Descriptions: map[Language]string{
			NL:  "Avontuurlijke student die quests voltooit, collectibles verzamelt en actief deelneemt aan campus games.",
			ENG: "Adventurous student who completes quests, collects items, and actively participates in campus games.",
		},
	},
	AcademicGuardian: {
		ID:   AcademicGuardian,
		Name: "Academic Guardian",

		Descriptions: map[Language]string{
			NL:  "Consistente en punctuele student die uitblinkt in aanwezigheid en het tijdig indienen van opdrachten.",
			ENG: "Consistent and punctual student excelling in attendance and timely submission of assignments.",
		},
	},
}
