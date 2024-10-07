package schemas

import (
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID          uuid.UUID `json:"id" binding:"required"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email,omitempty"`
	FullName    string    `json:"full_name"`
	IsVerified  bool      `json:"is_verified"`
	IsActivated bool      `json:"is_activated"`
	Role        string    `json:"role"`
}

type AdminResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
}

type OTPResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	PhoneNumber string    `json:"phone_number"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type VehicleResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LicensePlate string    `json:"license_plate"`
	Brand        string    `json:"brand"`
	Model        string    `json:"model"`
	FuelConsumed float64   `json:"fuel_consumed"`
}

type RideOfferResponse struct {
	ID                     uuid.UUID `json:"id"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
	StartLatitude          float64   `json:"start_latitude"`
	StartLongitude         float64   `json:"start_longitude"`
	EndLatitude            float64   `json:"end_latitude"`
	EndLongitude           float64   `json:"end_longitude"`
	DriverCurrentLatitude  float64   `json:"driver_current_latitude"`
	DriverCurrentLongitude float64   `json:"driver_current_longitude"`
	Status                 string    `json:"status"`
}

type WaypointResponse struct {
	ID        uuid.UUID `json:"id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Order     int       `json:"order"`
}

type RideRequestResponse struct {
	ID                    uuid.UUID `json:"id"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	StartLatitude         float64   `json:"start_latitude"`
	StartLongitude        float64   `json:"start_longitude"`
	EndLatitude           float64   `json:"end_latitude"`
	EndLongitude          float64   `json:"end_longitude"`
	RiderCurrentLatitude  float64   `json:"rider_current_latitude"`
	RiderCurrentLongitude float64   `json:"rider_current_longitude"`
	Status                string    `json:"status"`
}

type RideResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status"`
}

type TransactionResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
}

type RatingResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Rating    float64   `json:"rating"`
	Comment   string    `json:"comment"`
}

type NotificationResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	IsRead    bool      `json:"is_read"`
}

type ChatResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Message     string    `json:"message"`
	MessageType string    `json:"message_type"`
	IsRead      bool      `json:"is_read"`
}

type FavoriteLocationResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

type FuelPriceResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	FuelType  string    `json:"fuel_type"`
	Price     float64   `json:"price"`
}

type VehicleTypeResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Type         string    `json:"type"`
	Model        string    `json:"model"`
	Brand        string    `json:"brand"`
	FuelConsumed float64   `json:"fuel_consumed"`
}
