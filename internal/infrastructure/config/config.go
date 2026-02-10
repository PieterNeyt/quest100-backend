package config

import (
	"log"
	"os"
)

type Config struct {
	Port        string
	DatabaseUrl string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("PORT not set, using default 8080")
		port = "8080"
	}

	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Println("PORT not set, using default 8080")
		databaseUrl = ""
	}

	return &Config{
		Port:        port,
		DatabaseUrl: databaseUrl,
	}
}
