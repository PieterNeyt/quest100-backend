package schedular

import (
	lbApp "Quest100Backend/internal/leaderboard/application"
	nerdleApp "Quest100Backend/internal/minigame/nerdle/application"
	sudokuApp "Quest100Backend/internal/minigame/sudoku/application"
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

func StartDailyNerdleGame(nerdleServ nerdleApp.NerdleService) {
	loc, _ := time.LoadLocation("Europe/Brussels")
	scheduler := gocron.NewScheduler(loc)

	_, err := scheduler.Every(1).Day().At("00:01").Do(func() {
		log.Println("Nerdle: generating daily puzzle for", time.Now().Format("2006-01-02"))
		if err := nerdleServ.PrepareDailyGame(); err != nil {
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

func StartDailySudokuGame(sudokuServ sudokuApp.SudokuService) {
	loc, _ := time.LoadLocation("Europe/Brussels")
	scheduler := gocron.NewScheduler(loc)

	_, err := scheduler.Every(1).Day().At("00:01").Do(func() {
		log.Println("Sudoku: generating daily puzzle for", time.Now().Format("2006-01-02"))
		if err := sudokuServ.PrepareDailyGame(); err != nil {
			log.Printf("Sudoku: failed to generate daily game: %v", err)
		} else {
			log.Println("Sudoku: daily puzzle ready")
		}
	})
	if err != nil {
		log.Fatalf("Sudoku: failed to register daily scheduler job: %v", err)
	}

	scheduler.StartAsync()
}

func StartLeaderboardStatusUpdater(lbServ lbApp.LeaderboardService) {
	loc, _ := time.LoadLocation("Europe/Brussels")
	scheduler := gocron.NewScheduler(loc)

	_, err := scheduler.Every(1).Day().At("00:00").Do(func() {
		log.Println("Leaderboard: updating active statuses for", time.Now().Format("2006-01-02"))
		if err := lbServ.UpdateLeaderboardStatuses(); err != nil {
			log.Printf("Leaderboard: failed to update statuses: %v", err)
		} else {
			log.Println("Leaderboard: statuses updated successfully")
		}
	})
	if err != nil {
		log.Fatalf("Leaderboard: failed to register scheduler job: %v", err)
	}

	scheduler.StartAsync()
}
