package db

import (
	"fmt"
	"time"

	"shareway/db/migration"
	"shareway/util"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDatabaseInstance creates and configures a new database connection
func NewDatabaseInstance(cfg util.Config) *gorm.DB {
	// Construct the PostgreSQL connection string
	psqlconn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.DatabaseHost, cfg.DatabaseUsername, cfg.DatabasePassword, cfg.DatabaseName, cfg.DatabasePort)

	// Open a connection to the database
	db, err := gorm.Open(postgres.Open(psqlconn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Configure connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get database instance")
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(10 * time.Second)

	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error; err != nil {
		log.Fatal().Err(err).Msg("Failed to create UUID extension")
	}

	// Uncomment the following block to drop all tables before migration (use with caution)
	/*
		if err := migration.DropAllTables(db); err != nil {
			log.Fatal().Err(err).Msg("Failed to drop tables")
		}
	*/

	// Perform database migration
	if err := migration.Migrate(db); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	// Seed admin user
	if err := migration.SeedAdmin(db); err != nil {
		log.Fatal().Err(err).Msg("Failed to seed admin user")
	}

	return db
}
