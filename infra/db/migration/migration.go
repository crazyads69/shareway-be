package migration

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Define function to migrate tables

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&User{}, &Admin{}, &OTP{}, &PasetoToken{})
	if err != nil {
		return err
	}
	return nil
}

// Define function to drop tables
func Drop(db *gorm.DB) error {
	err := db.Migrator().DropTable(&User{}, &Admin{}, &OTP{}, &PasetoToken{})
	if err != nil {
		return err
	}
	return nil
}

// Define function to seed Admin table
func SeedAdmin(db *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &Admin{
		ID:       uuid.New(),
		Username: "admin",
		Password: string(hashedPassword),
	}

	return db.FirstOrCreate(&admin, Admin{Username: "admin"}).Error
}
