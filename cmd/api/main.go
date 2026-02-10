package main

import (
	"Quest100Backend/internal/infrastructure/config"
	"Quest100Backend/internal/infrastructure/database"
	"Quest100Backend/internal/infrastructure/router"
	"fmt"
	"log"
)

func main() {
	cfg := config.LoadConfig()

	r := router.SetupRoutes()

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server is running on port %s\n", cfg.Port)

	database.SetupDatabase(cfg.DatabaseUrl)

	if err := r.Run(addr); err != nil {
		log.Fatal("Server failed:", err)
	}
}
