package database

import (
	databaseProfile "Quest100Backend/internal/profile/infrastructure/database"
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Service interface {
	Close() error
	GetDB() *gorm.DB
}

type service struct {
	db *gorm.DB
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

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, username, password, database, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	autoMigration(db)

	log.Printf("Connected to database: %s", database)

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func autoMigration(db *gorm.DB) {
	databaseProfile.AutoMigration(db)
}

func (s *service) GetDB() *gorm.DB {
	return s.db
}

func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)

	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
