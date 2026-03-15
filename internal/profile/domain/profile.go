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
	AddAwardHistoryEntry(senderId uuid.UUID, receiverId uuid.UUID) error
	GetProfiles() (*[]Profile, error)
	HasSentAward(senderId uuid.UUID, recieverId uuid.UUID) (bool, error)
	GetSentAwardReceivers(id uuid.UUID) ([]uuid.UUID, error)
	GetProfileStats(profileId uuid.UUID) (ProfileStats, error)
	GetCampusByProfileID(id uuid.UUID) (string, error)
	GetLastKudosEntries(profileId uuid.UUID, limit int) ([]KudosEntry, error)
	GetKudoEntryById(id uuid.UUID) (*KudosEntry, error)
}

type Language string

const (
	NL  Language = "NL"
	ENG          = "EN"
)

type Profile struct {
	ID                   uuid.UUID          `gorm:"type:uuid;primaryKey;" json:"id"`
	FirstName            string             `json:"firstName"`
	LastName             string             `json:"lastName"`
	Email                string             `gorm:"uniqueIndex" json:"email"`
	Kudos                int                `json:"kudos"`
	CustomProfilePicture *string            `gorm:"type:text" json:"customProfilePicture"`
	Campus               string             `gorm:"type:varchar(100)" json:"campus"`
	ArchetypeID          int                `json:"archetypeId"`
	PreferredLanguage    Language           `gorm:"type:varchar(2);check:preferred_language IN ('NL','EN')" json:"preferredLanguage"`
	PlayerStats          ProfileStats       `gorm:"foreignKey:ProfileID;references:ID"`
	KudosHistory         []KudosEntry       `gorm:"foreignKey:ProfileID;references:ID"`
	AttendanceRecords    []AttendanceRecord `gorm:"foreignKey:ProfileID;references:ID" json:"attendanceRecords"`
}

type ProfileStats struct {
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

func (p *Profile) AddKudos(kudos int, reason string, kudoType KudoType, senderID ...uuid.UUID) error {
	if kudos < 0 {
		return &NegativeKudosError{Arg: kudos, Message: "Kudos can't be negative"}
	}

	var sender *uuid.UUID
	if len(senderID) > 0 {
		sender = &senderID[0]
	}

	p.Kudos += kudos
	p.KudosHistory = append(p.KudosHistory, KudosEntry{
		ID:        uuid.New(),
		ProfileID: p.ID,
		SenderID:  sender,
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

	if err := p.CalculateArcheType(); err != nil {
		log.Printf("Error calculating archetype during sync: %v", err)
	}

	return nil
}

func CreateProfile(graph *GraphProfile) *Profile {
	hardcodedID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")

	profile := &Profile{
		ID:          graph.Id,
		FirstName:   graph.Name,
		LastName:    graph.Surname,
		Email:       graph.Mail,
		Campus:      graph.OfficeLocation,
		Kudos:       285,
		ArchetypeID: 1,
		PlayerStats: ProfileStats{
			ProfileID:      graph.Id,
			KudoKnowledge:  80,
			KudoAttendance: 60,
			KudoTeamwork:   55,
			KudoAtmosphere: 45,
			KudoEngagement: 45,
		},
		PreferredLanguage: graph.PreferredLanguage,
		AttendanceRecords: []AttendanceRecord{},
		KudosHistory: []KudosEntry{
			{ID: uuid.New(), ProfileID: graph.Id, Amount: 50, Reason: "Aced the JavaScript quiz", Type: KudoKnowledge, SenderID: &hardcodedID},
			{ID: uuid.New(), ProfileID: graph.Id, Amount: 30, Reason: "Helped a teammate debug their code", Type: KudoTeamwork},
			{ID: uuid.New(), ProfileID: graph.Id, Amount: 40, Reason: "Active participation in class discussion", Type: KudoEngagement},
			{ID: uuid.New(), ProfileID: graph.Id, Amount: 25, Reason: "Organized a study group session", Type: KudoAtmosphere},
			{ID: uuid.New(), ProfileID: graph.Id, Amount: 35, Reason: "Perfect attendance this week", Type: KudoAttendance},
			{ID: uuid.New(), ProfileID: graph.Id, Amount: 60, Reason: "Submitted extra assignment", Type: KudoKnowledge},
			{ID: uuid.New(), ProfileID: graph.Id, Amount: 45, Reason: "Presented group project", Type: KudoTeamwork},
		},
	}

	return profile
}
