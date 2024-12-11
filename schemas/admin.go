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
