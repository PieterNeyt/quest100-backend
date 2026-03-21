package schedular

import (
	"Quest100Backend/internal/minigame/nerdle/application"
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/gorm"
)

func StartDailyTableCleanup(db *gorm.DB, tables []string, atTime string) {

	loc, _ := time.LoadLocation("Europe/Brussels")
	scheduler := gocron.NewScheduler(loc)

	_, err := scheduler.Every(1).Day().At(atTime).Do(func() {
		fmt.Println("Starting table cleanup:", time.Now())
		emptyTables(db, tables)
	})
	if err != nil {
		log.Fatalf("Failed to add scheduler job: %v", err)
	}

	scheduler.StartAsync()
}

func emptyTables(db *gorm.DB, tables []string) {
	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s;", table)).Error
		if err != nil {
			log.Printf("Error truncating %s: %v", table, err)
		} else {
			log.Printf("Table %s successfully truncated", table)
		}
	}
}

func StartDailyNerdleGame(nerdleServ application.NerdleService) {
	loc, _ := time.LoadLocation("Europe/Brussels")
	scheduler := gocron.NewScheduler(loc)

	_, err := scheduler.Every(1).Day().At("00:01").Do(func() {
		log.Println("Nerdle: generating daily puzzle for", time.Now().Format("2006-01-02"))
		if _, err := nerdleServ.GetTodayGame(); err != nil {
			log.Printf("Nerdle: failed to generate daily game: %v", err)
		} else {
			log.Println("Nerdle: daily puzzle ready")
		}
	})
	if err != nil {
		log.Fatalf("Nerdle: failed to register daily scheduler job: %v", err)
	}

	scheduler.StartAsync()
}
