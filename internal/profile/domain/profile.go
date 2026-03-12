package domain

import (
	"fmt"
	"log"
	"slices"
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
	GetAllAssets() (*[]Asset, error)
	GetAssetById(assetId string) (*Asset, error)
	GetProfileAssets(profileId uuid.UUID) (*[]Asset, error)
	GetProfileAvatar(profileId uuid.UUID) (*Avatar, error)
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
	Avatar               Avatar             `gorm:"foreignKey:ProfileID;references:ID" json:"avatar"`
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

// TODO nadenken hoe opslaan equipped item
type Asset struct {
	ID         string `gorm:"primaryKey"`
	Name       string
	Category   string
	LayerOrder int
	Price      int `gorm:"default:0"`
	Link       string
}

type Avatar struct {
	ProfileID uuid.UUID `gorm:"type:uuid;primaryKey;" json:"profileId"`
	BodyID    string    `gorm:"not null"`
	Body      Asset     `gorm:"foreignKey:BodyID"`

	EyesID string `gorm:"not null"`
	Eyes   Asset  `gorm:"foreignKey:EyesID"`

	ShirtsID *string
	Shirts   *Asset `gorm:"foreignKey:ShirtsID"`

	HairID *string
	Hair   *Asset `gorm:"foreignKey:HairID"`

	FacialHairID *string
	FacialHair   *Asset `gorm:"foreignKey:FacialHairID"`

	GlassesID *string
	Glasses   *Asset `gorm:"foreignKey:GlassesID"`

	AccessoriesID *string
	Accessories   *Asset `gorm:"foreignKey:AccessoriesID"`

	ExtrasID *string
	Extras   *Asset `gorm:"foreignKey:ExtrasID"`
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

func CreateProfile(graph *GraphProfile, defaultBodyID string, defaultEyesID string) *Profile {
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
		Avatar: Avatar{
			ProfileID: graph.Id,
			BodyID:    defaultBodyID,
			EyesID:    defaultEyesID,
		},

		Assets: []Asset{
			{ID: defaultBodyID},
			{ID: defaultEyesID},
		},
	}
}

func (p *Profile) OwnsAsset(assetId string) bool {
	if slices.ContainsFunc(p.Assets, func(asset Asset) bool {
		return asset.ID == assetId
	}) {
		return true
	}
	return false
}

func (p *Profile) BuyAsset(asset *Asset) error {
	if p.Kudos < asset.Price {
		return fmt.Errorf("kudos can't be less than price")
	}
	if p.OwnsAsset(asset.ID) {
		return fmt.Errorf("asset with id %s is already owned", asset.ID)
	}
	p.Assets = append(p.Assets, *asset)
	return nil
}

func (p *Profile) EquipAsset(asset *Asset) {
	switch asset.Category {
	case "Body":
		p.Avatar.BodyID = asset.ID
	case "Eyes":
		p.Avatar.EyesID = asset.ID
	case "Shirts":
		p.Avatar.ShirtsID = &asset.ID
	case "Hair":
		p.Avatar.HairID = &asset.ID
	case "FacialHair":
		p.Avatar.FacialHairID = &asset.ID
	case "Glasses":
		p.Avatar.GlassesID = &asset.ID
	case "Accessories":
		p.Avatar.AccessoriesID = &asset.ID
	case "Extras":
		p.Avatar.ExtrasID = &asset.ID

	}
}

func (a *Avatar) AvatarAsArray() []Asset {
	assets := make([]Asset, 0, 8)
	assets = append(assets, a.Body, a.Eyes)

	if a.Shirts != nil {
		assets = append(assets, *a.Shirts)
	}
	if a.Hair != nil {
		assets = append(assets, *a.Hair)
	}
	if a.FacialHair != nil {
		assets = append(assets, *a.FacialHair)
	}
	if a.Glasses != nil {
		assets = append(assets, *a.Glasses)
	}
	if a.Accessories != nil {
		assets = append(assets, *a.Accessories)
	}
	if a.Extras != nil {
		assets = append(assets, *a.Extras)
	}
	return assets
}
