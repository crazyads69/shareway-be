package schemas

import "github.com/google/uuid"

// Define the GetVehiclesResponse schema
type GetVehiclesResponse struct {
	Vehicles []Vehicle `json:"vehicles"`
}

// Define the Vehicle schema
type Vehicle struct {
	VehicleID uuid.UUID `json:"vehicle_id" binding:"required"`
	Name      string    `json:"name"`
}

// Define the RegisterVehicleRequest schema
type RegisterVehicleRequest struct {
	VehicleID    uuid.UUID `json:"vehicle_id" binding:"required,uuid" validate:"required,uuid"` // this id from the vehicle_type table in the database
	UserID       uuid.UUID `json:"user_id" binding:"required,uuid" validate:"required,uuid"`
	LicensePlate string    `json:"license_plate" binding:"required" validate:"required"`
	CaVet        string    `json:"ca_vet" binding:"required" validate:"required"`
}
