package migration

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Migrate creates all necessary tables in the database
func Migrate(db *gorm.DB) error {
	// AutoMigrate will create tables, foreign key constraints, and missing columns/indexes
	return db.AutoMigrate(
		&User{},
		&Admin{},
		&OTP{},
		&PasetoToken{},
		&Transaction{},
		&Vehicle{},
		&RideRequest{},
		&RideOffer{},
		// &Waypoint{},
		&Ride{},
		&Rating{},
		&Notification{},
		&Chat{},
		&FavoriteLocation{},
		&FuelPrice{},
		&VehicleType{},
	)
}

// DropAllTables removes all tables from the database
func DropAllTables(db *gorm.DB) error {
	// Drop tables in reverse order of dependencies to avoid foreign key constraint issues
	return db.Migrator().DropTable(&User{},
		&Admin{},
		&OTP{},
		&PasetoToken{},
		&Transaction{},
		&Vehicle{},
		&RideRequest{},
		&RideOffer{},
		// &Waypoint{},
		&Ride{},
		&Rating{},
		&Notification{},
		&Chat{},
		&FavoriteLocation{},
		&FuelPrice{},
		&VehicleType{})
}

// SeedAdmin creates an admin user if it doesn't already exist
func SeedAdmin(db *gorm.DB) error {
	// Check if admin already exists
	var count int64
	if err := db.Model(&Admin{}).Where("username = ?", "admin").Count(&count).Error; err != nil {
		return err
	}

	// If admin exists, no need to create
	if count > 0 {
		return nil
	}

	// Generate hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Create new admin
	admin := &Admin{
		ID:       uuid.New(),
		Username: "admin",
		Password: string(hashedPassword),
		Role:     "admin",
		FullName: "Quản trị viên tối cao",
	}

	// Insert admin into database
	return db.Create(admin).Error
}
