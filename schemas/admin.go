package schemas

import "time"

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
