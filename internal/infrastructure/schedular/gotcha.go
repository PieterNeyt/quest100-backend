package schedular

import (
	"Quest100Backend/internal/gotcha/application"
	gotchaDB "Quest100Backend/internal/gotcha/infrastructure/database"
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/gorm"
)

func StartGotchaScheduler(db *gorm.DB) {
	loc, _ := time.LoadLocation("Europe/Brussels")
	scheduler := gocron.NewScheduler(loc)

	gameRepo := gotchaDB.NewGameRepository(db)
	participantRepo := gotchaDB.NewParticipantRepository(db)
	killRepo := gotchaDB.NewKillRepository(db)
	propRepo := gotchaDB.NewPropRepository(db)
	service := application.NewGotchaService(gameRepo, participantRepo, killRepo, propRepo)

	// Elke dag controleren of een spel moet starten (startDate bereikt)
	_, err := scheduler.Every(1).Day().At("00:05").Do(func() {
		games, err := gameRepo.GetAllActiveGames()
		_ = games
		if err != nil {
			log.Printf("gotcha schedular: error fetching games: %v", err)
			return
		}
		// start alle spelen waarvan startDate <= now en status == OPT_IN
		// (uitbreiden: aparte methode GetGamesToStart in repository)
	})
	if err != nil {
		log.Printf("Failed to register gotcha start job: %v", err)
	}

	// Elk uur: verwerk timeouts
	_, err = scheduler.Every(1).Hour().Do(func() {
		if err := service.ProcessTimeouts(); err != nil {
			log.Printf("gotcha schedular: timeout processing error: %v", err)
		}
	})
	if err != nil {
		log.Printf("Failed to register gotcha timeout job: %v", err)
	}

	scheduler.StartAsync()
}
