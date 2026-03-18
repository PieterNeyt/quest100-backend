package database

import (
	"Quest100Backend/internal/profile/domain"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(&domain.Profile{}, &domain.KudosEntry{}, &domain.ProfileStats{},
		&domain.AwardHistoryEntry{}, &domain.Avatar{}, &domain.AttendanceRecord{}, &domain.Asset{},
		&domain.Course{}, &domain.Class{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}

	seedDatabase(db)
}

func seedDatabase(db *gorm.DB) {
	if err := syncGopherAssets(db); err != nil {
		log.Fatalf("Failed to sync assets: %v", err)
	}

	var defaultBody domain.Asset
	if err := db.Where("category = ? AND name = ?", "Body", "blue gopher").First(&defaultBody).Error; err != nil {
		log.Fatalf("Could not find default Body asset: %v", err)
	}

	var defaultEyes domain.Asset
	if err := db.Where("category = ? AND name = ?", "Eyes", "crazy eyes").First(&defaultEyes).Error; err != nil {
		log.Fatalf("Could not find default Eyes asset: %v", err)
	}

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

	if err := db.Where(domain.Profile{ID: hardcodedID}).FirstOrCreate(&hugo).Error; err != nil {
		log.Printf("Could not seed database: %v", err)
	}

	extraProfiles := []domain.Profile{
		{
			ID:                uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			FirstName:         "Wout",
			LastName:          "Brandhout",
			Email:             "wout.Brandhout@student.kdg.be",
			Kudos:             0,
			ArchetypeID:       2,
			Campus:            "Campus Stad",
			PreferredLanguage: domain.NL,
		},
		{
			ID:                uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			FirstName:         "Bart",
			LastName:          "slechtinbiljart",
			Email:             "bart.slechtinbiljart@student.kdg.be",
			Kudos:             0,
			ArchetypeID:       3,
			Campus:            "Campus Stad",
			PreferredLanguage: domain.NL,
		},
		{
			ID:                uuid.MustParse("00000000-0000-0000-0000-000000000004"),
			FirstName:         "Miezel",
			LastName:          "De Kiezel",
			Email:             "steen.dekiezel@student.kdg.be",
			Kudos:             0,
			ArchetypeID:       1,
			Campus:            "Campus Stad",
			PreferredLanguage: domain.NL,
		},
		{
			ID:                uuid.MustParse("00000000-0000-0000-0000-000000000005"),
			FirstName:         "Tim",
			LastName:          "Simsalabim",
			Email:             "tim.Simsalabim@student.kdg.be",
			Kudos:             0,
			ArchetypeID:       2,
			Campus:            "Campus Hoboken",
			PreferredLanguage: domain.NL,
		},
		{
			ID:                uuid.MustParse("00000000-0000-0000-0000-000000000006"),
			FirstName:         "Lien",
			LastName:          "Bijnaderinzien",
			Email:             "lien.Bijnaderinzien@student.kdg.be",
			Kudos:             0,
			ArchetypeID:       3,
			Campus:            "Campus Stad",
			PreferredLanguage: domain.NL,
		},
		{
			ID:                uuid.MustParse("00000000-0000-0000-0000-000000000007"),
			FirstName:         "Luis",
			LastName:          "Steengruis",
			Email:             "Luis.Steengruis@student.kdg.be",
			Kudos:             0,
			ArchetypeID:       1,
			Campus:            "Campus Hoboken",
			PreferredLanguage: domain.NL,
		},
		{
			ID:                uuid.MustParse("00000000-0000-0000-0000-000000000008"),
			FirstName:         "Ann",
			LastName:          "De Mortelman",
			Email:             "ann.demortelman@student.kdg.be",
			Kudos:             0,
			ArchetypeID:       2,
			Campus:            "Campus Stad",
			PreferredLanguage: domain.NL,
		},
		{
			ID:                uuid.MustParse("00000000-0000-0000-0000-000000000009"),
			FirstName:         "Rein",
			LastName:          "Azijn",
			Email:             "rein.Azijn@student.kdg.be",
			Kudos:             0,
			ArchetypeID:       3,
			Campus:            "Campus Hoboken",
			PreferredLanguage: domain.NL,
		},
		{
			ID:                uuid.MustParse("00000000-0000-0000-0000-000000000010"),
			FirstName:         "Noel",
			LastName:          "Zonderdoel",
			Email:             "noel.Zonderdoel@student.kdg.be",
			Kudos:             0,
			ArchetypeID:       1,
			Campus:            "Campus Stad",
			PreferredLanguage: domain.NL,
		},
		{
			ID:                uuid.MustParse("00000000-0000-0000-0000-000000000011"),
			FirstName:         "Mark",
			LastName:          "Benchmark",
			Email:             "mark.Benchmark@student.kdg.be",
			Kudos:             0,
			ArchetypeID:       2,
			Campus:            "Campus Hoboken",
			PreferredLanguage: domain.NL,
		},
	}

	for _, profile := range extraProfiles {
		p := profile
		if err := db.Where(domain.Profile{ID: p.ID}).FirstOrCreate(&p).Error; err != nil {
			log.Printf("Could not seed profile %s %s: %v", p.FirstName, p.LastName, err)
		}
	}

	seedAvatars(db, defaultBody.ID, defaultEyes.ID)
	seedCourses(db)
}
func seedCourses(db *gorm.DB) {
	courses := []domain.Course{
		{
			Id:   uuid.MustParse("10000000-0000-0000-0000-000000000001"),
			Name: "INF",
			Classes: []*domain.Class{
				{Id: uuid.MustParse("20000000-0000-0000-0000-000000000001"), Name: "101"},
				{Id: uuid.MustParse("20000000-0000-0000-0000-000000000002"), Name: "102"},
				{Id: uuid.MustParse("20000000-0000-0000-0000-000000000003"), Name: "103"},
				{Id: uuid.MustParse("20000000-0000-0000-0000-000000000004"), Name: "104"},
				{Id: uuid.MustParse("20000000-0000-0000-0000-000000000005"), Name: "105"},
				{Id: uuid.MustParse("20000000-0000-0000-0000-000000000006"), Name: "106"},
			},
		},
		{
			Id:   uuid.MustParse("10000000-0000-0000-0000-000000000002"),
			Name: "ACS",
			Classes: []*domain.Class{
				{Id: uuid.MustParse("20000000-0000-0000-0000-000000000007"), Name: "101"},
				{Id: uuid.MustParse("20000000-0000-0000-0000-000000000008"), Name: "102"},
				{Id: uuid.MustParse("20000000-0000-0000-0000-000000000009"), Name: "103"},
			},
		},
	}

	for _, course := range courses {
		c := course
		if err := db.Where(domain.Course{Id: c.Id}).FirstOrCreate(&c).Error; err != nil {
			log.Printf("Could not seed course %s: %v", c.Name, err)
			continue
		}

		for _, class := range c.Classes {
			cl := class
			cl.CourseId = c.Id
			if err := db.Where(domain.Class{Id: cl.Id}).FirstOrCreate(cl).Error; err != nil {
				log.Printf("Could not seed class %s for course %s: %v", cl.Name, c.Name, err)
			}
		}
	}
}

func seedAvatars(db *gorm.DB, bodyID string, eyesID string) {
	profileIDs := []uuid.UUID{
		uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		uuid.MustParse("00000000-0000-0000-0000-000000000004"),
		uuid.MustParse("00000000-0000-0000-0000-000000000005"),
		uuid.MustParse("00000000-0000-0000-0000-000000000006"),
		uuid.MustParse("00000000-0000-0000-0000-000000000007"),
		uuid.MustParse("00000000-0000-0000-0000-000000000008"),
		uuid.MustParse("00000000-0000-0000-0000-000000000009"),
		uuid.MustParse("00000000-0000-0000-0000-000000000010"),
		uuid.MustParse("00000000-0000-0000-0000-000000000011"),
	}

	for _, profileID := range profileIDs {
		avatar := domain.Avatar{
			ProfileID: profileID,
			BodyID:    bodyID,
			EyesID:    eyesID,
		}
		if err := db.Where(domain.Avatar{ProfileID: profileID}).FirstOrCreate(&avatar).Error; err != nil {
			log.Printf("Could not seed avatar for profile %s: %v", profileID, err)
		}
	}
}

func syncGopherAssets(db *gorm.DB) error {
	resp, err := http.Get("https://gopherize.me/api/artwork/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var apiResponse struct {
		Categories []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Images []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				Link      string `json:"href"`
				Thumbnail string `json:"thumbnail_href"`
			} `json:"images"`
		} `json:"categories"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return err
	}

	var assets []domain.Asset
	for _, category := range apiResponse.Categories {
		for i, image := range category.Images {
			asset := domain.Asset{
				ID:        uuid.NewSHA1(uuid.NameSpaceURL, []byte(image.Link)).String(),
				Name:      image.Name,
				Category:  category.Name,
				Price:     i * 10,
				Link:      image.Link,
				Thumbnail: image.Thumbnail,
			}

			if category.Name == "Body" || category.Name == "Eyes" {
				asset.Price = 0
			}

			assets = append(assets, asset)
		}
	}

	if result := db.Save(&assets); result.Error != nil {
		return result.Error
	}
	return nil
}
