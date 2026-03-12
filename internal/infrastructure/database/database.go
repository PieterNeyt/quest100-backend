package database

import (
	databaseComm "Quest100Backend/internal/communication/infrastructure/database"
	databaseEvent "Quest100Backend/internal/event/infrastructure/database"
	databaseProfile "Quest100Backend/internal/gotcha/infrastructure/database"
	databaseGotcha "Quest100Backend/internal/profile/infrastructure/database"
	"fmt"
	"log"
	"os"
	"strings"

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
	resetDB    = os.Getenv("RESET_DATABASE")
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

	if strings.ToLower(resetDB) == "true" {
		log.Println("dropping all tables")
		dropAllTables(db)
	}

	autoMigration(db)

	log.Printf("Connected to database: %s", database)

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func dropAllTables(db *gorm.DB) {
	var tables []string

	schemaName := schema
	if schemaName == "" {
		schemaName = "public"
	}

	err := db.Raw(`
       SELECT tablename 
       FROM pg_tables 
       WHERE schemaname = ?
    `, schemaName).Scan(&tables).Error

	if err != nil {
		log.Printf("Error fetching tables: %v", err)
		return
	}

	// Drop alle tabellen
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS \"%s\" CASCADE", table)).Error; err != nil {
			log.Printf("Error dropping table %s: %v", table, err)
		}
	}
}

func autoMigration(db *gorm.DB) {
	databaseProfile.AutoMigration(db)
	databaseEvent.AutoMigration(db)
	databaseGotcha.AutoMigration(db)
	databaseComm.AutoMigration(db)
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
