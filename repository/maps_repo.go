package repository

import "gorm.io/gorm"

type IMapsRepository interface{}

type MapsRepository struct {
	db *gorm.DB
}

func NewMapsRepository(db *gorm.DB) IMapsRepository {
	return &MapsRepository{db: db}
}
