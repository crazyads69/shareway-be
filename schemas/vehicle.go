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
