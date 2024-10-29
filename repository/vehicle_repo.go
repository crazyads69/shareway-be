package repository

import (
	"shareway/infra/db/migration"
	"shareway/schemas"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IVehicleRepository interface {
	GetVehicles() ([]schemas.Vehicle, error)
	RegisterVehicle(userID uuid.UUID, vehicleID uuid.UUID, licensePlate string, caVet string) error
	LicensePlateExists(licensePlate string) (bool, error)
	CaVetExists(caVet string) (bool, error)
	GetVehicleFromID(vehicleID uuid.UUID) (schemas.VehicleDetail, error)
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
			VehicleID:    vehicle.ID,
			Name:         vehicle.Name,
			FuelConsumed: vehicle.FuelConsumed,
		}
	}

	return schemaVehicles, nil
}

// RegisterVehicle registers a vehicle for a user
func (r *VehicleRepository) RegisterVehicle(userID uuid.UUID, vehicleID uuid.UUID, licensePlate string, caVet string) error {
	// Get the vehicle type from the database
	var vehicle migration.VehicleType
	if err := r.db.First(&vehicle, vehicleID).Error; err != nil {
		return err
	}

	// Create a new vehicle registration record
	vehicleRegistration := migration.Vehicle{
		UserID:        userID,
		VehicleTypeID: vehicleID,
		LicensePlate:  licensePlate,
		Name:          vehicle.Name,
		FuelConsumed:  vehicle.FuelConsumed,
		CaVet:         caVet,
	}

	// Insert the new record into the database
	if err := r.db.Create(&vehicleRegistration).Error; err != nil {
		return err
	}

	return nil
}

// LicensePlateExists checks if a given license plate already exists in the database
func (r *VehicleRepository) LicensePlateExists(licensePlate string) (bool, error) {
	var count int64
	err := r.db.Model(&migration.Vehicle{}).Where("license_plate = ?", licensePlate).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CaVetExists checks if a given CA VET already exists in the database
func (r *VehicleRepository) CaVetExists(caVet string) (bool, error) {
	var count int64
	err := r.db.Model(&migration.Vehicle{}).Where("ca_vet = ?", caVet).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetVehicleFromID retrieves a vehicle from the database using the vehicle ID
func (r *VehicleRepository) GetVehicleFromID(vehicleID uuid.UUID) (schemas.VehicleDetail, error) {
	var vehicle migration.Vehicle
	if err := r.db.First(&vehicle, vehicleID).Error; err != nil {
		return schemas.VehicleDetail{}, err
	}

	return schemas.VehicleDetail{
		VehicleID:    vehicle.ID,
		Name:         vehicle.Name,
		FuelConsumed: vehicle.FuelConsumed,
		LicensePlate: vehicle.LicensePlate,
	}, nil
}
