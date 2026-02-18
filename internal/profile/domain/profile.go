package domain

import (
	"log"
	"time"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfileById(id uuid.UUID) (*Profile, error)
	UpdateProfile(profile *Profile) error
	SaveProfile(profile *Profile) error
}

type Language string

const (
	NL  Language = "NL"
	ENG          = "EN"
)

type Profile struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	FirstName            string    `json:"firstName"`
	LastName             string    `json:"lastName"`
	Email                string    `gorm:"uniqueIndex" json:"email"`
	Kudos                int       `json:"kudos"`
	CustomProfilePicture *string   `gorm:"type:text" json:"customProfilePicture"`
	ArchetypeID          int
	PreferredLanguage    Language           `gorm:"type:varchar(2);check:preferred_language IN ('NL','EN')" json:"preferredLanguage"`
	PlayerStats          PlayerStats        `gorm:"foreignKey:ProfileID;references:ID"`
	KudosHistory         []KudosEntry       `gorm:"foreignKey:ProfileID;references:ID"`
	AttendanceRecords    []AttendanceRecord `gorm:"foreignKey:ProfileID;references:ID" json:"attendanceRecords"`
}

type PlayerStats struct {
	ProfileID      uuid.UUID `gorm:"type:uuid;primaryKey;"`
	KudoKnowledge  int
	KudoAttendance int
	KudoTeamwork   int
	KudoAtmosphere int
	KudoEngagement int
}

type AttendanceRecord struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	ProfileID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_profile_class_unique" json:"profileId"`
	ClassID   uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_profile_class_unique" json:"classId"`
	Timestamp time.Time `json:"timestamp"`
}

func (p *Profile) HasAttendedClass(classId uuid.UUID) bool {
	for _, record := range p.AttendanceRecords {
		if record.ClassID == classId {
			return true
		}
	}
	return false
}

func (p *Profile) RecordAttendance(classId uuid.UUID) error {
	if p.HasAttendedClass(classId) {
		return &DuplicateAttendanceError{
			ProfileID: p.ID,
			ClassID:   classId,
			Message:   "Attendance already recorded for this class",
		}
	}

	record := AttendanceRecord{
		ID:        uuid.New(),
		ProfileID: p.ID,
		ClassID:   classId,
		Timestamp: time.Now(),
	}

	p.AttendanceRecords = append(p.AttendanceRecords, record)
	return nil
}

func (p *Profile) AddKudos(kudos int, reason string, kudoType KudoType) error {
	if kudos < 0 {
		return &NegativeKudosError{Arg: kudos, Message: "Kudos can't be negative"}
	}

	p.Kudos += kudos
	p.KudosHistory = append(p.KudosHistory, KudosEntry{
		ID:        uuid.New(),
		ProfileID: p.ID,
		Amount:    kudos,
		Reason:    reason,
		Type:      kudoType,
	})

	p.UpdatePlayerStats(kudos, kudoType)
	return nil
}

func (p *Profile) UpdatePlayerStats(amount int, kudoType KudoType) {
	switch kudoType {
	case KudoKnowledge:
		p.PlayerStats.KudoKnowledge += amount
	case KudoAttendance:
		p.PlayerStats.KudoAttendance += amount
	case KudoTeamwork:
		p.PlayerStats.KudoTeamwork += amount
	case KudoAtmosphere:
		p.PlayerStats.KudoAtmosphere += amount
	case KudoEngagement:
		p.PlayerStats.KudoEngagement += amount
	}

	if err := p.CalculateArcheType(); err != nil {
		log.Printf("Error calculating archetype: %v", err)
	}
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
		ID:                graph.Id,
		FirstName:         graph.Name,
		LastName:          graph.Surname,
		Email:             graph.Mail,
		Kudos:             0,
		ArchetypeID:       1,
		PlayerStats:       PlayerStats{},
		KudosHistory:      []KudosEntry{},
		AttendanceRecords: []AttendanceRecord{},
		PreferredLanguage: graph.PreferredLanguage,
	}
}
