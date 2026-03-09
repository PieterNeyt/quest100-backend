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
	if err := db.AutoMigrate(&domain.Profile{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}
	if err := db.AutoMigrate(&domain.KudosEntry{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}
	if err := db.AutoMigrate(&domain.ProfileStats{}); err != nil {
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
		PreferredLanguage:    domain.NL,
		CustomProfilePicture: &picture,
	}

	err := db.Where(domain.Profile{ID: hardcodedID}).FirstOrCreate(&hugo).Error
	if err != nil {
		log.Printf("Could not seed database: %v", err)
	}

	if err := syncGopherAssets(db); err != nil {
		log.Fatalf("Failed to sync assets: %v", err)
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

	for _, category := range apiResponse.Categories {
		for _, image := range category.Images {
			asset := domain.Asset{
				ID:         image.ID,
				Name:       image.Name,
				Category:   category.Name,
				LayerOrder: 0,
				Price:      100,
				Link:       image.Link,
			}

			if category.Name == "Body" || category.Name == "Eyes" {
				asset.Price = 0
			}

			db.Create(&asset)
		}
	}
	return nil
}
