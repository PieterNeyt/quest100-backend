package domain

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfileById(id uuid.UUID) (*Profile, error)
	SaveProfile(profile *Profile) error
	AddAwardHistoryEntry(senderId uuid.UUID, receiverId uuid.UUID) error
	GetProfiles() (*[]Profile, error)
	HasSentAward(senderId uuid.UUID, recieverId uuid.UUID) (bool, error)
	GetSentAwardReceivers(id uuid.UUID) ([]uuid.UUID, error)
	GetProfileStats(profileId uuid.UUID) (ProfileStats, error)
	GetAssets(profileId uuid.UUID) (*[]Asset, error)
	GetAssetById(assetId string) (*Asset, error)
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
	PlayerStats          ProfileStats       `gorm:"foreignKey:ProfileID;references:ID"`
	KudosHistory         []KudosEntry       `gorm:"foreignKey:ProfileID;references:ID"`
	AttendanceRecords    []AttendanceRecord `gorm:"foreignKey:ProfileID;references:ID" json:"attendanceRecords"`
	Assets               []Asset            `gorm:"many2many:user_assets;" json:"assets"`
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

type Asset struct {
	ID         string `gorm:"primaryKey"`
	Name       string
	Category   string
	LayerOrder int
	Price      int `gorm:"default:0"`
	Link       string
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
		ID:          graph.Id,
		FirstName:   graph.Name,
		LastName:    graph.Surname,
		Email:       graph.Mail,
		Kudos:       0,
		ArchetypeID: 1,
		PlayerStats: ProfileStats{
			ProfileID:      graph.Id,
			KudoKnowledge:  0,
			KudoAttendance: 0,
			KudoTeamwork:   0,
			KudoAtmosphere: 0,
			KudoEngagement: 0,
		},
		KudosHistory:      []KudosEntry{},
		PreferredLanguage: graph.PreferredLanguage,
		AttendanceRecords: []AttendanceRecord{},
	}
}

func (p *Profile) BuyAsset(asset *Asset) error {
	if p.Kudos < asset.Price {
		return fmt.Errorf("kudos can't be less than price")
	}
	p.Kudos -= asset.Price
	p.Assets = append(p.Assets, *asset)
	return nil
}
