package repository

import (
	"context"
	"errors"
	"shareway/infra/db/migration"
	"shareway/schemas"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type IVehicleRepository interface {
	GetVehicles(ctx context.Context, limit int, page int, input string) ([]schemas.Vehicle, error)
	RegisterVehicle(userID uuid.UUID, vehicleID uuid.UUID, licensePlate string, caVet string) error
	LicensePlateExists(licensePlate string) (bool, error)
	CaVetExists(caVet string) (bool, error)
	GetVehicleFromID(vehicleID uuid.UUID) (schemas.VehicleDetail, error)
	GetAllVehiclesFromUserID(userID uuid.UUID) ([]schemas.VehicleDetail, error)
	GetTotalVehiclesForUser(userID uuid.UUID) (int64, error)
	GetVehiclesForUser(userID uuid.UUID) ([]schemas.VehicleDetail, error)
}

type VehicleRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewVehicleRepository(db *gorm.DB, redis *redis.Client) IVehicleRepository {
	return &VehicleRepository{db: db, redis: redis}
}

var (
	ErrVehicleNotFound = errors.New("vehicle not found")
)

// GetVehicles retrieves all vehicles from the database and converts them to schema format
// func (r *VehicleRepository) GetVehicles(ctx context.Context, limit int, page int, input string) ([]schemas.Vehicle, error) {
// 	var vehicles []migration.VehicleType
// 	input = strings.ToLower(input)
// 	query := r.db.Model(&migration.VehicleType{}).
// 		Select("id", "name", "fuel_consumed").
// 		Limit(limit).
// 		Offset(page * limit)

// 	if input != "" {
// 		query = query.Where("LOWER(name) LIKE ?", "%"+input+"%")
// 	}

// 	err := query.Find(&vehicles).Error
// 	if err != nil {
// 		return nil, err
// 	}

// 	schemaVehicles := make([]schemas.Vehicle, len(vehicles))
// 	for i, vehicle := range vehicles {
// 		schemaVehicles[i] = schemas.Vehicle{
// 			VehicleID:    vehicle.ID,
// 			Name:         vehicle.Name,
// 			FuelConsumed: vehicle.FuelConsumed,
// 		}
// 	}

// 	return schemaVehicles, nil
// }

// GetVehicles retrieves all vehicles from the database and converts them to schema format
func (r *VehicleRepository) GetVehicles(ctx context.Context, limit int, page int, input string) ([]schemas.Vehicle, error) {
	var vehicles []migration.VehicleType
	query := r.db.Model(&migration.VehicleType{}).
		Select("id", "name", "fuel_consumed").
		Limit(limit).
		Offset(page * limit)

	if input != "" {
		// Use ILIKE for case-insensitive search in PostgreSQL
		query = query.Where("name ILIKE ?", "%"+input+"%")
	}

	err := query.Find(&vehicles).Error
	if err != nil {
		return nil, err
	}

	schemaVehicles := make([]schemas.Vehicle, len(vehicles))
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
func (r *VehicleRepository) RegisterVehicle(userID, vehicleID uuid.UUID, licensePlate, caVet string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var vehicle migration.VehicleType
		if err := tx.Select("name", "fuel_consumed").First(&vehicle, vehicleID).Error; err != nil {
			return err
		}

		vehicleRegistration := migration.Vehicle{
			UserID:        userID,
			VehicleTypeID: vehicleID,
			LicensePlate:  licensePlate,
			Name:          vehicle.Name,
			FuelConsumed:  vehicle.FuelConsumed,
			CaVet:         caVet,
		}

		return tx.Create(&vehicleRegistration).Error
	})
}

// LicensePlateExists checks if a given license plate already exists in the database
func (r *VehicleRepository) LicensePlateExists(licensePlate string) (bool, error) {
	var exists bool
	err := r.db.Model(&migration.Vehicle{}).
		Select("1").
		Where("license_plate = ?", licensePlate).
		Limit(1).
		Find(&exists).
		Error
	return exists, err
}

// CaVetExists checks if a given CA VET already exists in the database
func (r *VehicleRepository) CaVetExists(caVet string) (bool, error) {
	var exists bool
	err := r.db.Model(&migration.Vehicle{}).
		Select("1").
		Where("ca_vet = ?", caVet).
		Limit(1).
		Find(&exists).
		Error
	return exists, err
}

// GetVehicleFromID retrieves a vehicle from the database using the vehicle ID
func (r *VehicleRepository) GetVehicleFromID(vehicleID uuid.UUID) (schemas.VehicleDetail, error) {
	var vehicle migration.Vehicle
	err := r.db.Select("id", "name", "fuel_consumed", "license_plate").
		First(&vehicle, vehicleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return schemas.VehicleDetail{}, ErrVehicleNotFound
		}
		return schemas.VehicleDetail{}, err
	}

	return schemas.VehicleDetail{
		VehicleID:    vehicle.ID,
		Name:         vehicle.Name,
		FuelConsumed: vehicle.FuelConsumed,
		LicensePlate: vehicle.LicensePlate,
	}, nil
}

// GetAllVehiclesFromUserID retrieves all vehicles for a user using the user ID
func (r *VehicleRepository) GetAllVehiclesFromUserID(userID uuid.UUID) ([]schemas.VehicleDetail, error) {
	var vehicles []migration.Vehicle
	err := r.db.Select("id", "name", "fuel_consumed", "license_plate").
		Where("user_id = ?", userID).
		Find(&vehicles).Error
	if err != nil {
		return nil, err
	}

	schemaVehicles := make([]schemas.VehicleDetail, len(vehicles))
	for i, vehicle := range vehicles {
		schemaVehicles[i] = schemas.VehicleDetail{
			VehicleID:    vehicle.ID,
			Name:         vehicle.Name,
			FuelConsumed: vehicle.FuelConsumed,
			LicensePlate: vehicle.LicensePlate,
		}
	}
	return schemaVehicles, nil
}

// GetTotalVehiclesForUser retrieves the total number of vehicles for a user
func (r *VehicleRepository) GetTotalVehiclesForUser(userID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.Model(&migration.Vehicle{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	return total, err
}

// GetVehiclesForUser retrieves all vehicles for a user using the user ID
func (r *VehicleRepository) GetVehiclesForUser(userID uuid.UUID) ([]schemas.VehicleDetail, error) {
	var vehicles []migration.Vehicle
	err := r.db.Select("id", "name", "fuel_consumed", "license_plate").
		Where("user_id = ?", userID).
		Find(&vehicles).Error
	if err != nil {
		return nil, err
	}

	schemaVehicles := make([]schemas.VehicleDetail, len(vehicles))
	for i, vehicle := range vehicles {
		schemaVehicles[i] = schemas.VehicleDetail{
			VehicleID:    vehicle.ID,
			Name:         vehicle.Name,
			FuelConsumed: vehicle.FuelConsumed,
			LicensePlate: vehicle.LicensePlate,
		}
	}
	return schemaVehicles, nil
}

// Make sure VehicleRepository implements IVehicleRepository
var _ IVehicleRepository = (*VehicleRepository)(nil)
