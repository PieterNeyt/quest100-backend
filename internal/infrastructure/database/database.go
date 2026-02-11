package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

type Service interface {
	Close() error
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("QUEST_DB_DATABASE")
	password   = os.Getenv("QUEST_DB_PASSWORD")
	username   = os.Getenv("QUEST_DB_USERNAME")
	port       = os.Getenv("QUEST_DB_PORT")
	host       = os.Getenv("QUEST_DB_HOST")
	schema     = os.Getenv("QUEST_DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}
