package schemas

import (
	"time"

	"github.com/google/uuid"
)

type DashboardGeneralDataResponse struct {
	TotalUsers        int64   `json:"total_users"`
	UserChange        float64 `json:"user_change"`
	TotalRides        int64   `json:"total_rides"`
	RideChange        float64 `json:"ride_change"`
	TotalTransactions int64   `json:"total_transactions"`
	TransactionChange float64 `json:"transaction_change"`
	TotalVehicles     int64   `json:"total_vehicles"`
	VehicleChange     float64 `json:"vehicle_change"`
}

type UserDashboardDataResponse struct {
	UserStats []StatPoint `json:"user_stats"`
}

type RideDashboardDataResponse struct {
	RideStats []StatPoint `json:"ride_stats"`
}

type TransactionDashboardDataResponse struct {
	TransactionStats []StatPoint `json:"transaction_stats"`
}

type VehicleDashboardDataResponse struct {
	VehicleStats []StatPoint `json:"vehicle_stats"`
}

type StatPoint struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
	Total int64     `json:"total"` // For transaction total amount
}

type FilterDashboardDataRequest struct {
	Filter    string    `form:"filter" binding:"required"`
	StartDate time.Time `form:"start_date" time_format:"2006-01-02"` // Use time.Time for date parsing
	EndDate   time.Time `form:"end_date" time_format:"2006-01-02"`   // Use time.Time for date parsing
}

type UserDetail struct {
	ID                uuid.UUID       `json:"id" binding:"required"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	PhoneNumber       string          `json:"phone_number"`
	Email             string          `json:"email,omitempty"`
	CCCDNumber        string          `json:"cccd_number,omitempty"`
	AvatarURL         string          `json:"avatar_url"`
	FullName          string          `json:"full_name"`
	IsVerified        bool            `json:"is_verified"`
	IsActivated       bool            `json:"is_activated"`
	IsMomoLinked      bool            `json:"is_momo_linked"`
	BalanceInApp      int64           `json:"balance_in_app"`
	Role              string          `json:"role"`
	Gender            string          `json:"gender"`
	AverageRating     float64         `json:"average_rating"`
	TotalRatings      int64           `json:"total_ratings"`
	TotalRides        int64           `json:"total_rides"`
	TotalTransactions int64           `json:"total_transactions"`
	TotalVehicles     int64           `json:"total_vehicles"`
	Vehicles          []VehicleDetail `json:"vehicles"`
}

// Define UserManagementRequest schema
type UserListRequest struct {
	Page           int       `form:"page" binding:"required,min=1"`          // Page number for pagination
	Limit          int       `form:"limit" binding:"required,min=1,max=100"` // Limit number for pagination (max 100)
	StartDate      time.Time `form:"start_date" time_format:"2006-01-02"`    // Use time.Time for date parsing
	EndDate        time.Time `form:"end_date" time_format:"2006-01-02"`      // Use time.Time for date parsing
	IsActivated    *bool     `form:"is_activated"`                           // Optional filter for is_activated
	IsVerified     *bool     `form:"is_verified"`                            // Optional filter for is_verified
	SearchFullName string    `form:"search_full_name"`                       // Optional filter for full name
}

type UserListResponse struct {
	TotalPages  int64 `json:"total_pages"`
	CurrentPage int   `json:"current_page"`
	Limit       int   `json:"limit"`
	TotalUsers  int64 `json:"total_users"`
	// The detail user response
	Users []UserDetail `json:"users"`
}

type RideListRequest struct {
	Page           int       `form:"page" binding:"required,min=1"`            // Page number for pagination
	Limit          int       `form:"limit" binding:"required,min=1,max=100"`   // Limit number for pagination (max 100)
	StartDate      time.Time `form:"start_date_time" time_format:"2006-01-02"` // Use time.Time for date parsing
	EndDate        time.Time `form:"end_date_time" time_format:"2006-01-02"`   // Use time.Time for date parsing
	SearchFullName string    `form:"search_full_name"`                         // Optional filter for full name
	SearchRoute    string    `form:"search_route"`                             // Optional filter for route
	SearchVehicle  string    `form:"search_vehicle"`                           // Optional filter for vehicle
	RideStatus     []string  `form:"ride_status"`                              // Optional filter for ride status
}

type RideDetail struct {
	ID                     uuid.UUID         `json:"ride_id"`
	RideOfferID            uuid.UUID         `json:"ride_offer_id"`
	Driver                 UserInfo          `json:"driver"`
	Hitcher                UserInfo          `json:"hitcher"`
	RideRequestID          uuid.UUID         `json:"ride_request_id"`
	Status                 string            `json:"status"`
	StartTime              time.Time         `json:"start_time"`
	EndTime                time.Time         `json:"end_time"`
	StartAddress           string            `json:"start_address"`
	EndAddress             string            `json:"end_address"`
	Fare                   int64             `json:"fare"`
	EncodedPolyline        string            `json:"encoded_polyline"`
	Distance               float64           `json:"distance"`
	Duration               int               `json:"duration"`
	Transaction            TransactionDetail `json:"transaction"`
	StartLatitude          float64           `json:"start_latitude"`
	StartLongitude         float64           `json:"start_longitude"`
	EndLatitude            float64           `json:"end_latitude"`
	EndLongitude           float64           `json:"end_longitude"`
	Vehicle                VehicleDetail     `json:"vehicle"`
	DriverCurrentLatitude  float64           `json:"driver_current_latitude"`
	DriverCurrentLongitude float64           `json:"driver_current_longitude"`
	RiderCurrentLatitude   float64           `json:"rider_current_latitude"`
	RiderCurrentLongitude  float64           `json:"rider_current_longitude"`
	Waypoints              []Waypoint        `json:"waypoints"`
}

type RideListResponse struct {
	TotalPages  int64        `json:"total_pages"`
	CurrentPage int          `json:"current_page"`
	Limit       int          `json:"limit"`
	TotalRides  int64        `json:"total_rides"`
	Rides       []RideDetail `json:"rides"`
}

type VehicleListRequest struct {
	Page              int       `form:"page" binding:"required,min=1"`          // Page number for pagination
	Limit             int       `form:"limit" binding:"required,min=1,max=100"` // Limit number for pagination (max 100)
	StartDate         time.Time `form:"start_date" time_format:"2006-01-02"`    // Use time.Time for date parsing
	EndDate           time.Time `form:"end_date" time_format:"2006-01-02"`      // Use time.Time for date parsing
	SearchOwner       string    `form:"search_owner"`                           // Optional filter for owner
	SearchPlate       string    `form:"search_plate"`                           // Optional filter for plate
	SearchVehicleName string    `form:"search_vehicle_name"`                    // Optional filter for vehicle name
	SearchCavet       string    `form:"search_cavet"`                           // Optional filter for cavet
}

type VehicleListDetail struct {
	Owner        UserInfo  `json:"owner"` // Owner info of the vehicle
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	LicensePlate string    `json:"license_plate"`
	VehicleName  string    `json:"vehicle_name"`
	CaVet        string    `json:"cavet"` // Certificate of vehicle registration each vehicle has a unique number
	FuelConsumed float64   `json:"fuel_consumed"`
	TotalRides   int64     `json:"total_rides"` // Total rides of the vehicle that have been taken
}

type VehicleListResponse struct {
	TotalPages    int64               `json:"total_pages"`
	CurrentPage   int                 `json:"current_page"`
	Limit         int                 `json:"limit"`
	TotalVehicles int64               `json:"total_vehicles"`
	Vehicles      []VehicleListDetail `json:"vehicles"`
}

type TransactionListRequest struct {
	Page           int       `form:"page" binding:"required,min=1"`          // Page number for pagination
	Limit          int       `form:"limit" binding:"required,min=1,max=100"` // Limit number for pagination (max 100)
	StartDate      time.Time `form:"start_date" time_format:"2006-01-02"`    // Use time.Time for date parsing
	EndDate        time.Time `form:"end_date" time_format:"2006-01-02"`      // Use time.Time for date parsing
	SearchSender   string    `form:"search_sender"`                          // Optional filter for sender
	SearchReceiver string    `form:"search_receiver"`                        // Optional filter for receiver
	PaymentMethod  []string  `form:"payment_method"`                         // Optional filter for payment method
	PaymentStatus  []string  `form:"payment_status"`                         // Optional filter for payment status
	MinAmount      int64     `form:"min_amount"`                             // Optional filter for min amount
	MaxAmount      int64     `form:"max_amount"`                             // Optional filter for max amount
}

type TransactionListDetail struct {
	ID            uuid.UUID `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	Sender        UserInfo  `json:"sender"`
	Receiver      UserInfo  `json:"receiver"`
	Amount        int64     `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	PaymentStatus string    `json:"payment_status"`
}

type TransactionListResponse struct {
	TotalPages        int64                   `json:"total_pages"`
	CurrentPage       int                     `json:"current_page"`
	Limit             int                     `json:"limit"`
	TotalTransactions int64                   `json:"total_transactions"`
	Transactions      []TransactionListDetail `json:"transactions"`
}
