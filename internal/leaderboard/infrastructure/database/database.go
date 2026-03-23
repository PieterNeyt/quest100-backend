package database

import (
	"Quest100Backend/internal/leaderboard/domain"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(
		&domain.Leaderboard{},
		&domain.LeaderboardClass{},
	); err != nil {
		log.Printf("Failed to migrate leaderboard database: %v", err)
	}
	seedLeaderboards(db)
}

func seedLeaderboards(db *gorm.DB) {
	infCourseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	acsCourseID := uuid.MustParse("10000000-0000-0000-0000-000000000002")

	inf101 := uuid.MustParse("20000000-0000-0000-0000-000000000001")
	inf102 := uuid.MustParse("20000000-0000-0000-0000-000000000002")
	inf103 := uuid.MustParse("20000000-0000-0000-0000-000000000003")
	inf104 := uuid.MustParse("20000000-0000-0000-0000-000000000004")
	inf105 := uuid.MustParse("20000000-0000-0000-0000-000000000005")
	inf106 := uuid.MustParse("20000000-0000-0000-0000-000000000006")

	acs101 := uuid.MustParse("20000000-0000-0000-0000-000000000007")
	acs102 := uuid.MustParse("20000000-0000-0000-0000-000000000008")
	acs103 := uuid.MustParse("20000000-0000-0000-0000-000000000009")

	leaderboards := []struct {
		lb      domain.Leaderboard
		classes []uuid.UUID
		kudos   []int
	}{
		// ── INF: oud ────────────────────────────────────────────────────
		{
			lb: domain.Leaderboard{
				ID:        uuid.MustParse("30000000-0000-0000-0000-000000000001"),
				CourseId:  infCourseID,
				StartDate: time.Now().AddDate(0, -4, 0),
				EndDate:   time.Now().AddDate(0, -2, 0),
				Prize: domain.Prize{
					Name:        "KdG Goodie Bag",
					Description: "Een goodie bag vol KdG merchandise voor de winnende klas!",
					PhotoURL:    "https://example.com/goodiebag.png",
				},
			},
			classes: []uuid.UUID{inf101, inf102, inf103, inf104, inf105, inf106},
			kudos:   []int{340, 210, 290, 175, 420, 310},
		},
		// ── INF: huidig actief ───────────────────────────────────────────
		{
			lb: domain.Leaderboard{
				ID:        uuid.MustParse("30000000-0000-0000-0000-000000000002"),
				CourseId:  infCourseID,
				StartDate: time.Now().AddDate(0, -1, 0),
				EndDate:   time.Now().AddDate(0, 1, 0),
				Prize: domain.Prize{
					Name:        "Bowling met de klas",
					Description: "De winnende klas gaat een avond bowlen op kosten van KdG!",
					PhotoURL:    "https://example.com/bowling.png",
				},
			},
			classes: []uuid.UUID{inf101, inf102, inf103, inf104, inf105, inf106},
			kudos:   []int{120, 95, 140, 60, 185, 110},
		},
		// ── INF: aankomend ──────────────────────────────────────────────
		{
			lb: domain.Leaderboard{
				ID:        uuid.MustParse("30000000-0000-0000-0000-000000000006"),
				CourseId:  infCourseID,
				StartDate: time.Now().AddDate(0, 1, 0),
				EndDate:   time.Now().AddDate(0, 3, 0),
				Prize: domain.Prize{
					Name:        "Laser Shoot met de klas",
					Description: "De winnende klas gaat een avond lasershooten op kosten van KdG!",
					PhotoURL:    "https://example.com/lasershoot.png",
				},
			},
			classes: []uuid.UUID{inf101, inf102, inf103, inf104, inf105, inf106},
			kudos:   []int{0, 0, 0, 0, 0, 0},
		},
		// ── ACS: oud 1 ──────────────────────────────────────────────────
		{
			lb: domain.Leaderboard{
				ID:        uuid.MustParse("30000000-0000-0000-0000-000000000003"),
				CourseId:  acsCourseID,
				StartDate: time.Now().AddDate(0, -6, 0),
				EndDate:   time.Now().AddDate(0, -4, 0),
				Prize: domain.Prize{
					Name:        "KdG Rekenmachine",
					Description: "Elke student van de winnende klas krijgt een KdG rekenmachine!",
					PhotoURL:    "https://example.com/calculator.png",
				},
			},
			classes: []uuid.UUID{acs101, acs102, acs103},
			kudos:   []int{530, 390, 460},
		},
		// ── ACS: oud 2 ──────────────────────────────────────────────────
		{
			lb: domain.Leaderboard{
				ID:        uuid.MustParse("30000000-0000-0000-0000-000000000004"),
				CourseId:  acsCourseID,
				StartDate: time.Now().AddDate(0, -3, 0),
				EndDate:   time.Now().AddDate(0, -1, 0),
				Prize: domain.Prize{
					Name:        "KdG Trui",
					Description: "Elke student van de winnende klas krijgt een exclusieve KdG trui!",
					PhotoURL:    "https://example.com/sweater.png",
				},
			},
			classes: []uuid.UUID{acs101, acs102, acs103},
			kudos:   []int{275, 410, 320},
		},
		// ── ACS: aankomend ──────────────────────────────────────────────
		{
			lb: domain.Leaderboard{
				ID:        uuid.MustParse("30000000-0000-0000-0000-000000000005"),
				CourseId:  acsCourseID,
				StartDate: time.Now().AddDate(0, 1, 0),
				EndDate:   time.Now().AddDate(0, 3, 0),
				Prize: domain.Prize{
					Name:        "Lasershooting met de klas",
					Description: "De winnende klas gaat een avond lasershooten op kosten van KdG!",
					PhotoURL:    "https://example.com/lasershoot.png",
				},
			},
			classes: []uuid.UUID{acs101, acs102, acs103},
			kudos:   []int{0, 0, 0},
		},
	}

	for _, entry := range leaderboards {
		lb := entry.lb

		var existing domain.Leaderboard
		if err := db.First(&existing, "id = ?", lb.ID).Error; err == nil {
			continue
		}

		for i, classID := range entry.classes {
			lb.Classes = append(lb.Classes, &domain.LeaderboardClass{
				LeaderboardID: lb.ID,
				ClassId:       classID,
				TotalKudos:    entry.kudos[i],
			})
		}

		if err := db.Create(&lb).Error; err != nil {
			log.Printf("Could not seed leaderboard %s: %v", lb.ID, err)
		}
	}
}
