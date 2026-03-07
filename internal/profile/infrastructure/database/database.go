package database

import (
	"Quest100Backend/internal/profile/domain"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(&domain.Profile{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}
	if err := db.AutoMigrate(&domain.KudosEntry{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}
	if err := db.AutoMigrate(&domain.PlayerStats{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}
	if err := db.AutoMigrate(&domain.AwardHistoryEntry{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}

	errAttendance := db.AutoMigrate(&domain.AttendanceRecord{})
	if errAttendance != nil {
		log.Printf("Failed to migrate AttendanceRecord: %v", errAttendance)
	}

	seedDatabase(db)
}

func seedDatabase(db *gorm.DB) {
	hardcodedID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	picture := "https://media.licdn.com/dms/image/v2/D4D03AQFUJr-0NnhW_w/profile-displayphoto-shrink_200_200/profile-displayphoto-shrink_200_200/0/1715145874530?e=2147483647&v=beta&t=9w9T7AsoZPaZ3q9AOIRALdaVxel2rcC8BH0ynPczQqQ"
	hugo := domain.Profile{
		ID:                   hardcodedID,
		FirstName:            "Jon",
		LastName:             "Beton",
		Email:                "jon.beton@student.kdg.be",
		Kudos:                0,
		ArchetypeID:          1,
		Campus:               "Campus Stad",
		PreferredLanguage:    domain.NL,
		CustomProfilePicture: &picture,
	}

	err := db.Where(domain.Profile{ID: hardcodedID}).FirstOrCreate(&hugo).Error
	if err != nil {
		log.Printf("Could not seed database: %v", err)
	}

	extraProfiles := []domain.Profile{
		{
			ID:                   uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			FirstName:            "Wout",
			LastName:             "Hout",
			Email:                "wout.hout@student.kdg.be",
			Kudos:                0,
			ArchetypeID:          2,
			Campus:               "Campus Stad",
			PreferredLanguage:    domain.NL,
			CustomProfilePicture: nil,
		},
		{
			ID:                   uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			FirstName:            "Bart",
			LastName:             "Asfalt",
			Email:                "bart.asfalt@student.kdg.be",
			Kudos:                0,
			ArchetypeID:          3,
			Campus:               "Campus Stad",
			PreferredLanguage:    domain.NL,
			CustomProfilePicture: nil,
		},
		{
			ID:                   uuid.MustParse("00000000-0000-0000-0000-000000000004"),
			FirstName:            "Miezel",
			LastName:             "De Kiezel",
			Email:                "steen.koolsteen@student.kdg.be",
			Kudos:                0,
			ArchetypeID:          1,
			Campus:               "Campus Stad",
			PreferredLanguage:    domain.NL,
			CustomProfilePicture: nil,
		},
		{
			ID:                   uuid.MustParse("00000000-0000-0000-0000-000000000005"),
			FirstName:            "Tim",
			LastName:             "Morteltim",
			Email:                "tim.morteltim@student.kdg.be",
			Kudos:                0,
			ArchetypeID:          2,
			Campus:               "Campus Hoboken",
			PreferredLanguage:    domain.NL,
			CustomProfilePicture: nil,
		},
		{
			ID:                   uuid.MustParse("00000000-0000-0000-0000-000000000006"),
			FirstName:            "Lien",
			LastName:             "Porsellien",
			Email:                "lien.porsellien@student.kdg.be",
			Kudos:                0,
			ArchetypeID:          3,
			Campus:               "Campus Stad",
			PreferredLanguage:    domain.NL,
			CustomProfilePicture: nil,
		},
		{
			ID:                   uuid.MustParse("00000000-0000-0000-0000-000000000007"),
			FirstName:            "Luis",
			LastName:             "Steengruis",
			Email:                "kris.graniet@student.kdg.be",
			Kudos:                0,
			ArchetypeID:          1,
			Campus:               "Campus Hoboken",
			PreferredLanguage:    domain.NL,
			CustomProfilePicture: nil,
		},
		{
			ID:                   uuid.MustParse("00000000-0000-0000-0000-000000000008"),
			FirstName:            "Ann",
			LastName:             "De Mortelman",
			Email:                "ann.tegann@student.kdg.be",
			Kudos:                0,
			ArchetypeID:          2,
			Campus:               "Campus Stad",
			PreferredLanguage:    domain.NL,
			CustomProfilePicture: nil,
		},
		{
			ID:                   uuid.MustParse("00000000-0000-0000-0000-000000000009"),
			FirstName:            "Rein",
			LastName:             "Kalkstein",
			Email:                "rein.kalkstein@student.kdg.be",
			Kudos:                0,
			ArchetypeID:          3,
			Campus:               "Campus Hoboken",
			PreferredLanguage:    domain.NL,
			CustomProfilePicture: nil,
		},
		{
			ID:                   uuid.MustParse("00000000-0000-0000-0000-000000000010"),
			FirstName:            "Noel",
			LastName:             "Pannoel",
			Email:                "noel.pannoel@student.kdg.be",
			Kudos:                0,
			ArchetypeID:          1,
			Campus:               "Campus Stad",
			PreferredLanguage:    domain.NL,
			CustomProfilePicture: nil,
		},
		{
			ID:                   uuid.MustParse("00000000-0000-0000-0000-000000000011"),
			FirstName:            "Mark",
			LastName:             "Remark",
			Email:                "mark.remark@student.kdg.be",
			Kudos:                0,
			ArchetypeID:          2,
			Campus:               "Campus Hoboken",
			PreferredLanguage:    domain.NL,
			CustomProfilePicture: nil,
		},
	}

	for _, profile := range extraProfiles {
		p := profile
		if err := db.Where(domain.Profile{ID: p.ID}).FirstOrCreate(&p).Error; err != nil {
			log.Printf("Could not seed profile %s %s: %v", p.FirstName, p.LastName, err)
		}
	}
}
