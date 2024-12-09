package repository

import (
	"shareway/helper"
	"shareway/infra/db/migration"
	"shareway/schemas"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewAdminRepository(db *gorm.DB, redis *redis.Client) IAdminRepository {
	return &AdminRepository{
		db:    db,
		redis: redis,
	}
}

type IAdminRepository interface {
	CheckAdminExists(req schemas.LoginAdminRequest) (migration.Admin, error)
	GetAdminProfile(adminID uuid.UUID) (migration.Admin, error)
	GetDashboardGeneralData() (schemas.DashboardGeneralDataResponse, error)
	GetUserDashboardData(startDate time.Time, endDate time.Time) (schemas.UserDashboardDataResponse, error)
	GetRideDashboardData(startDate time.Time, endDate time.Time) (schemas.RideDashboardDataResponse, error)
	GetTransactionDashboardData(startDate time.Time, endDate time.Time) (schemas.TransactionDashboardDataResponse, error)
	GetVehicleDashboardData(startDate time.Time, endDate time.Time) (schemas.VehicleDashboardDataResponse, error)
}

// CheckAdminExists checks if the admin exists in the database
func (r *AdminRepository) CheckAdminExists(req schemas.LoginAdminRequest) (migration.Admin, error) {
	var admin migration.Admin
	if err := r.db.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		return admin, err
	}
	return admin, nil
}

// GetAdminProfile gets the profile of the admin
func (r *AdminRepository) GetAdminProfile(adminID uuid.UUID) (migration.Admin, error) {
	var admin migration.Admin
	if err := r.db.Where("id = ?", adminID).First(&admin).Error; err != nil {
		return admin, err
	}
	return admin, nil
}

// GetDashboardGeneralData gets the general data for the dashboard
func (r *AdminRepository) GetDashboardGeneralData() (schemas.DashboardGeneralDataResponse, error) {
	var dashboardGeneralData schemas.DashboardGeneralDataResponse
	var err error

	now := time.Now()
	startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfLastMonth := startOfThisMonth.AddDate(0, -1, 0)

	// Get total number of users
	if err = r.db.Model(&migration.User{}).Count(&dashboardGeneralData.TotalUsers).Error; err != nil {
		return dashboardGeneralData, err
	}

	// Get number of users this month and last month
	var usersThisMonth, usersLastMonth int64
	if err = r.db.Model(&migration.User{}).Where("created_at >= ?", startOfThisMonth).Count(&usersThisMonth).Error; err != nil {
		return dashboardGeneralData, err
	}
	if err = r.db.Model(&migration.User{}).Where("created_at >= ? AND created_at < ?", startOfLastMonth, startOfThisMonth).Count(&usersLastMonth).Error; err != nil {
		return dashboardGeneralData, err
	}

	// Calculate user change percentage
	dashboardGeneralData.UserChange = helper.CalculatePercentageChange(usersThisMonth, usersLastMonth)

	// Get total number of completed rides
	if err = r.db.Model(&migration.Ride{}).Where("status = ?", "completed").Count(&dashboardGeneralData.TotalRides).Error; err != nil {
		return dashboardGeneralData, err
	}

	// Get number of completed rides this month and last month
	var ridesThisMonth, ridesLastMonth int64
	if err = r.db.Model(&migration.Ride{}).Where("status = ? AND created_at >= ?", "completed", startOfThisMonth).Count(&ridesThisMonth).Error; err != nil {
		return dashboardGeneralData, err
	}
	if err = r.db.Model(&migration.Ride{}).Where("status = ? AND created_at >= ? AND created_at < ?", "completed", startOfLastMonth, startOfThisMonth).Count(&ridesLastMonth).Error; err != nil {
		return dashboardGeneralData, err
	}

	// Calculate ride change percentage
	dashboardGeneralData.RideChange = helper.CalculatePercentageChange(ridesThisMonth, ridesLastMonth)

	// Get total transactions amount
	if err = r.db.Model(&migration.Transaction{}).Where("status = ?", "completed").Select("COALESCE(SUM(amount), 0)").Scan(&dashboardGeneralData.TotalTransactions).Error; err != nil {
		return dashboardGeneralData, err
	}

	// Get transactions amount this month and last month
	var transactionsThisMonth, transactionsLastMonth int64
	if err = r.db.Model(&migration.Transaction{}).Where("status = ? AND created_at >= ?", "completed", startOfThisMonth).Select("COALESCE(SUM(amount), 0)").Scan(&transactionsThisMonth).Error; err != nil {
		return dashboardGeneralData, err
	}
	if err = r.db.Model(&migration.Transaction{}).Where("status = ? AND created_at >= ? AND created_at < ?", "completed", startOfLastMonth, startOfThisMonth).Select("COALESCE(SUM(amount), 0)").Scan(&transactionsLastMonth).Error; err != nil {
		return dashboardGeneralData, err
	}

	// Calculate transaction change percentage
	dashboardGeneralData.TransactionChange = helper.CalculatePercentageChange(transactionsThisMonth, transactionsLastMonth)

	// Get total number of vehicles
	if err = r.db.Model(&migration.Vehicle{}).Count(&dashboardGeneralData.TotalVehicles).Error; err != nil {
		return dashboardGeneralData, err
	}

	// Get number of vehicles this month and last month
	var vehiclesThisMonth, vehiclesLastMonth int64
	if err = r.db.Model(&migration.Vehicle{}).Where("created_at >= ?", startOfThisMonth).Count(&vehiclesThisMonth).Error; err != nil {
		return dashboardGeneralData, err
	}
	if err = r.db.Model(&migration.Vehicle{}).Where("created_at >= ? AND created_at < ?", startOfLastMonth, startOfThisMonth).Count(&vehiclesLastMonth).Error; err != nil {
		return dashboardGeneralData, err
	}

	// Calculate vehicle change percentage
	dashboardGeneralData.VehicleChange = helper.CalculatePercentageChange(vehiclesThisMonth, vehiclesLastMonth)

	return dashboardGeneralData, nil
}

// GetUserDashboardData gets the data for the user dashboard
func (r *AdminRepository) GetUserDashboardData(startDate time.Time, endDate time.Time) (schemas.UserDashboardDataResponse, error) {
	var userDashboardData schemas.UserDashboardDataResponse
	// Get user from the database and group by created_at
	err := r.db.Model(&migration.User{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&userDashboardData.UserStats).Error
	if err != nil {
		return userDashboardData, err
	}

	return userDashboardData, nil
}

// GetRideDashboardData gets the data for the ride dashboard
func (r *AdminRepository) GetRideDashboardData(startDate time.Time, endDate time.Time) (schemas.RideDashboardDataResponse, error) {
	var rideDashboardData schemas.RideDashboardDataResponse
	// Get ride from the database and group by created_at
	err := r.db.Model(&migration.Ride{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&rideDashboardData.RideStats).Error
	if err != nil {
		return rideDashboardData, err
	}

	return rideDashboardData, nil
}

// GetTransactionDashboardData gets the data for the transaction dashboard
func (r *AdminRepository) GetTransactionDashboardData(startDate time.Time, endDate time.Time) (schemas.TransactionDashboardDataResponse, error) {
	var transactionDashboardData schemas.TransactionDashboardDataResponse
	// Get transaction from the database and group by created_at
	err := r.db.Model(&migration.Transaction{}).
		Select("DATE(created_at) as date, COUNT(*) as count, COALESCE(SUM(amount), 0) as total").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&transactionDashboardData.TransactionStats).Error
	if err != nil {
		return transactionDashboardData, err
	}

	return transactionDashboardData, nil
}

// GetVehicleDashboardData gets the data for the vehicle dashboard
func (r *AdminRepository) GetVehicleDashboardData(startDate time.Time, endDate time.Time) (schemas.VehicleDashboardDataResponse, error) {
	var vehicleDashboardData schemas.VehicleDashboardDataResponse
	// Get vehicle from the database and group by created_at
	err := r.db.Model(&migration.Vehicle{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&vehicleDashboardData.VehicleStats).Error
	if err != nil {
		return vehicleDashboardData, err
	}

	return vehicleDashboardData, nil
}

// Ensure that the AdminRepository implements the IAdminRepository interface
var _ IAdminRepository = (*AdminRepository)(nil)
