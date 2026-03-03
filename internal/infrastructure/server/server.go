package server

import (
	"Quest100Backend/internal/infrastructure/database"
	"Quest100Backend/internal/infrastructure/schedular"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int
	db   database.Service
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,
		db:   database.New(),
	}

	schedular.StartDailyTableCleanup(
		NewServer.db.GetDB(),
		[]string{"award_history_entries"},
		"02:00",
	)

	schedular.StartGotchaScheduler(NewServer.db.GetDB())
	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
