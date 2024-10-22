package repository

import (
	"shareway/infra/db/migration"
	"shareway/schemas"

	"gorm.io/gorm"
)

type IVehicleRepository interface {
	GetVehicles() ([]schemas.Vehicle, error)
}

type VehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository(db *gorm.DB) IVehicleRepository {
	return &VehicleRepository{db: db}
}

// GetVehicles retrieves all vehicles from the database and converts them to schema format
func (r *VehicleRepository) GetVehicles() ([]schemas.Vehicle, error) {
	var vehicles []migration.VehicleType

	// Fetch all vehicles from the database
	if err := r.db.Find(&vehicles).Error; err != nil {
		return nil, err
	}

	// Preallocate the slice for better performance
	schemaVehicles := make([]schemas.Vehicle, len(vehicles))

	// Convert the vehicles to the schema type
	for i, vehicle := range vehicles {
		schemaVehicles[i] = schemas.Vehicle{
			VehicleID: vehicle.ID,
			Name:      vehicle.Name,
		}
	}

	return schemaVehicles, nil
}
