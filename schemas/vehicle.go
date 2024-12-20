package schemas

import "github.com/google/uuid"

// Define the GetVehiclesResponse schema

// Define the GetVehiclesRequest schema
type GetVehiclesRequest struct {
	Limit int `form:"limit" binding:"required" validate:"required"`
	Page  int `form:"page" binding:"required" validate:"required"`
	// Input is the search input for the vehicle name (optional)
	Input string `form:"input,omitempty"`
}

type GetVehiclesResponse struct {
	Vehicles []Vehicle `json:"vehicles"`
}

// Define the Vehicle schema
type Vehicle struct {
	VehicleID    uuid.UUID `json:"vehicle_id" binding:"required"`
	Name         string    `json:"name"`
	FuelConsumed float64   `json:"fuel_consumed"`
}

// Define the RegisterVehicleRequest schema
type RegisterVehicleRequest struct {
	VehicleID    uuid.UUID `json:"vehicle_id" binding:"required,uuid" validate:"required,uuid"` // this id from the vehicle_type table in the database
	UserID       uuid.UUID `json:"user_id" binding:"required,uuid" validate:"required,uuid"`
	LicensePlate string    `json:"license_plate" binding:"required" validate:"required"`
	CaVet        string    `json:"ca_vet" binding:"required" validate:"required"`
}

// Define the VehicleDetail schema
type VehicleDetail struct {
	VehicleID    uuid.UUID `json:"vehicle_id" binding:"required"`
	Name         string    `json:"name"`
	FuelConsumed float64   `json:"fuel_consumed"`
	LicensePlate string    `json:"license_plate"`
}

// Define the GetVehicleRequest schema
type GetVehicleRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required,uuid" validate:"required,uuid"`
}

// Define the GetVehicleResponse schema
type GetVehicleResponse struct {
	Vehicle []VehicleDetail `json:"vehicle"` // this is an array because a user can have multiple vehicles
}
