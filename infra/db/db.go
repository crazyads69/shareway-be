package db

// Import GORM and Postgres driver
import (
	"fmt"
	"golang_template/util"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Create DB connection instance and return it

func NewDataBaseInstance(cfg util.Config) *gorm.DB {
	// Define connection string
	psqlconn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.DatabaseHost,
		cfg.DatabaseUsername,
		cfg.DatabasePassword,
		cfg.DatabaseName,
		cfg.DatabasePort,
	)

	// Connect to DB
	db, err := gorm.Open(postgres.Open(psqlconn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to DB")
	}

	// Connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to DB")
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(10 * time.Second)

	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error; err != nil {
		log.Fatal().Err(err).Msg("Failed to create UUID extension")
	}

	return db
}
