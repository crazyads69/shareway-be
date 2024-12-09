package controller

import (
	"fmt"
	"shareway/helper"
	"shareway/middleware"
	"shareway/schemas"
	"shareway/service"
	"shareway/util"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

// AdminController handles authentication-related requests
type AdminController struct {
	cfg          util.Config
	validate     *validator.Validate
	AdminService service.IAdminService
}

// NewAdminController creates a new AdminController instance
func NewAdminController(cfg util.Config, validate *validator.Validate, adminService service.IAdminService) *AdminController {
	return &AdminController{
		cfg:          cfg,
		validate:     validate,
		AdminService: adminService,
	}
}

// GetAdminProfile returns the profile of the admin
// @Summary Get the profile of the admin
// @Description Get the profile of the admin
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.Response{data=schemas.GetAdminProfileResponse} "Admin profile"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /admin/get-profile [get]
func (ac *AdminController) GetAdminProfile(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert the payload to a map of string and interface
	// Convert payload to map
	data, err := helper.ConvertToAdminPayload(payload)

	// If error occurs, return error response
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get the admin ID from the payload and use it to get the admin profile
	admin, err := ac.AdminService.GetAdminProfile(data.AdminID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get admin profile",
			"Không thể lấy thông tin admin",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Return the admin profile
	res := schemas.GetAdminProfileResponse{
		AdminInfo: schemas.AdminInfo{
			ID:        admin.ID,
			FullName:  admin.FullName,
			CreatedAt: admin.CreatedAt,
			UpdatedAt: admin.UpdatedAt,
			Username:  admin.Username,
			Role:      admin.Role,
		},
	}

	response := helper.SuccessResponse(res, "Admin profile retrieved successfully", "Lấy thông tin admin thành công")
	helper.GinResponse(ctx, 200, response)
}

// GetDashboardGeneralData returns the general data of the dashboard (total users, total rides, total transactions and their changes)
// @Summary Get the general data of the dashboard
// @Description Get the general data of the dashboard
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.Response{data=schemas.DashboardGeneralDataResponse} "Dashboard general data"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /admin/get-dashboard-general-data [get]
func (ac *AdminController) GetDashboardGeneralData(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert the payload to a map of string and interface
	// Convert payload to map
	data, err := helper.ConvertToAdminPayload(payload)

	// If error occurs, return error response
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	log.Info().Msgf("Admin ID: %s", data.AdminID)

	// Get the general data of the dashboard
	generalData, err := ac.AdminService.GetDashboardGeneralData()
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get dashboard general data",
			"Không thể lấy thông tin tổng quan của dashboard",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Return the general data of the dashboard
	res := generalData

	response := helper.SuccessResponse(res, "Dashboard general data retrieved successfully", "Lấy thông tin tổng quan của dashboard thành công")
	helper.GinResponse(ctx, 200, response)
}

// GetUserDashboardData returns the data of the dashboard for the user to visualize charts
// @Summary Get the data of the dashboard for the user
// @Description Get the data of the dashboard for the user
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param filter query string true "Filter for the data (all_time, last_week, last_month, last_year, custom)"
// @Param start_date query string false "Start date for custom filter (YYYY-MM-DD)"
// @Param end_date query string false "End date for custom filter (YYYY-MM-DD)"
// @Success 200 {object} helper.Response{data=schemas.UserDashboardDataResponse} "User dashboard data"
// @Failure 400 {object} helper.Response "Bad request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /admin/get-user-dashboard-data [get]
func (ac *AdminController) GetUserDashboardData(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert the payload to a map of string and interface
	// Convert payload to map
	data, err := helper.ConvertToAdminPayload(payload)

	// If error occurs, return error response
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	log.Info().Msgf("Admin ID: %s", data.AdminID)

	var req schemas.FilterDashboardDataRequest

	// Bind request to struct
	if err := ctx.ShouldBind(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind request",
			"Không thể bind request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := ac.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	var customStartDate, customEndDate time.Time

	if req.Filter == "custom" {
		// Parse the start date and end date from the query to UTC time
		customStartDate, err = time.Parse(time.RFC3339, req.StartDate.Format(time.RFC3339))
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to parse start date",
				"Không thể chuyển đổi ngày bắt đầu",
			)
			helper.GinResponse(ctx, 400, response)
			return
		}
		customEndDate, err = time.Parse(time.RFC3339, req.EndDate.Format(time.RFC3339))
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to parse end date",
				"Không thể chuyển đổi ngày kết thúc",
			)
			helper.GinResponse(ctx, 400, response)
			return
		}
	}

	// Get the data for the user dashboard
	userData, err := ac.AdminService.GetUserDashboardData(req.Filter, customStartDate, customEndDate)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get user dashboard data",
			"Không thể lấy thông tin dashboard của user",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Return the data for the user dashboard
	res := userData

	response := helper.SuccessResponse(res, "User dashboard data retrieved successfully", "Lấy thông tin dashboard của user thành công")
	helper.GinResponse(ctx, 200, response)
}

// GetRideDashboardData returns the data of the dashboard for the ride to visualize charts
// @Summary Get the data of the dashboard for the ride
// @Description Get the data of the dashboard for the ride
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param filter query string true "Filter for the data (all_time, last_week, last_month, last_year, custom)"
// @Param start_date query string false "Start date for custom filter (YYYY-MM-DD)"
// @Param end_date query string false "End date for custom filter (YYYY-MM-DD)"
// @Success 200 {object} helper.Response{data=schemas.RideDashboardDataResponse} "Ride dashboard data"
// @Failure 400 {object} helper.Response "Bad request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /admin/get-ride-dashboard-data [get]
func (ac *AdminController) GetRideDashboardData(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert the payload to a map of string and interface
	// Convert payload to map
	data, err := helper.ConvertToAdminPayload(payload)

	// If error occurs, return error response
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	log.Info().Msgf("Admin ID: %s", data.AdminID)

	var req schemas.FilterDashboardDataRequest

	// Bind request to struct
	if err := ctx.ShouldBind(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind request",
			"Không thể bind request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := ac.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	var customStartDate, customEndDate time.Time

	if req.Filter == "custom" {
		// Parse the start date and end date from the query to UTC time
		customStartDate, err = time.Parse(time.RFC3339, req.StartDate.Format(time.RFC3339))
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to parse start date",
				"Không thể chuyển đổi ngày bắt đầu",
			)
			helper.GinResponse(ctx, 400, response)
			return
		}
		customEndDate, err = time.Parse(time.RFC3339, req.EndDate.Format(time.RFC3339))
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to parse end date",
				"Không thể chuyển đổi ngày kết thúc",
			)
			helper.GinResponse(ctx, 400, response)
			return
		}
	}

	// Get the data for the ride dashboard
	rideData, err := ac.AdminService.GetRideDashboardData(req.Filter, customStartDate, customEndDate)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride dashboard data",
			"Không thể lấy thông tin dashboard của ride",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Return the data for the ride dashboard
	res := rideData

	response := helper.SuccessResponse(res, "Ride dashboard data retrieved successfully", "Lấy thông tin dashboard của ride thành công")
	helper.GinResponse(ctx, 200, response)
}

// GetTransactionDashboardData returns the data of the dashboard for the transaction to visualize charts
// @Summary Get the data of the dashboard for the transaction
// @Description Get the data of the dashboard for the transaction
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param filter query string true "Filter for the data (all_time, last_week, last_month, last_year, custom)"
// @Param start_date query string false "Start date for custom filter (YYYY-MM-DD)"
// @Param end_date query string false "End date for custom filter (YYYY-MM-DD)"
// @Success 200 {object} helper.Response{data=schemas.TransactionDashboardDataResponse} "Transaction dashboard data"
// @Failure 400 {object} helper.Response "Bad request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /admin/get-transaction-dashboard-data [get]
func (ac *AdminController) GetTransactionDashboardData(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert the payload to a map of string and interface
	// Convert payload to map
	data, err := helper.ConvertToAdminPayload(payload)

	// If error occurs, return error response
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	log.Info().Msgf("Admin ID: %s", data.AdminID)

	var req schemas.FilterDashboardDataRequest

	// Bind request to struct
	if err := ctx.ShouldBind(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind request",
			"Không thể bind request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := ac.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	var customStartDate, customEndDate time.Time

	if req.Filter == "custom" {
		// Parse the start date and end date from the query to UTC time
		customStartDate, err = time.Parse(time.RFC3339, req.StartDate.Format(time.RFC3339))
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to parse start date",
				"Không thể chuyển đổi ngày bắt đầu",
			)
			helper.GinResponse(ctx, 400, response)
			return
		}
		customEndDate, err = time.Parse(time.RFC3339, req.EndDate.Format(time.RFC3339))
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to parse end date",
				"Không thể chuyển đổi ngày kết thúc",
			)
			helper.GinResponse(ctx, 400, response)
			return
		}
	}

	// Get the data for the transaction dashboard
	transactionData, err := ac.AdminService.GetTransactionDashboardData(req.Filter, customStartDate, customEndDate)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get transaction dashboard data",
			"Không thể lấy thông tin dashboard của transaction",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Return the data for the transaction dashboard
	res := transactionData

	response := helper.SuccessResponse(res, "Transaction dashboard data retrieved successfully", "Lấy thông tin dashboard của transaction thành công")
	helper.GinResponse(ctx, 200, response)
}

// GetVehicleDashboardData returns the data of the dashboard for the vehicle to visualize charts
// @Summary Get the data of the dashboard for the vehicle
// @Description Get the data of the dashboard for the vehicle
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param filter query string true "Filter for the data (all_time, last_week, last_month, last_year, custom)"
// @Param start_date query string false "Start date for custom filter (YYYY-MM-DD)"
// @Param end_date query string false "End date for custom filter (YYYY-MM-DD)"
// @Success 200 {object} helper.Response{data=schemas.VehicleDashboardDataResponse} "Vehicle dashboard data"
// @Failure 400 {object} helper.Response "Bad request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /admin/get-vehicle-dashboard-data [get]
func (ac *AdminController) GetVehicleDashboardData(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert the payload to a map of string and interface
	// Convert payload to map
	data, err := helper.ConvertToAdminPayload(payload)

	// If error occurs, return error response
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	log.Info().Msgf("Admin ID: %s", data.AdminID)

	var req schemas.FilterDashboardDataRequest

	// Bind request to struct
	if err := ctx.ShouldBind(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind request",
			"Không thể bind request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := ac.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	var customStartDate, customEndDate time.Time

	if req.Filter == "custom" {
		// Parse the start date and end date from the query to UTC time
		customStartDate, err = time.Parse(time.RFC3339, req.StartDate.Format(time.RFC3339))
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to parse start date",
				"Không thể chuyển đổi ngày bắt đầu",
			)
			helper.GinResponse(ctx, 400, response)
			return
		}
		customEndDate, err = time.Parse(time.RFC3339, req.EndDate.Format(time.RFC3339))
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to parse end date",
				"Không thể chuyển đổi ngày kết thúc",
			)
			helper.GinResponse(ctx, 400, response)
			return
		}
	}

	// Get the data for the vehicle dashboard
	vehicleData, err := ac.AdminService.GetVehicleDashboardData(req.Filter, customStartDate, customEndDate)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get vehicle dashboard data",
			"Không thể lấy thông tin dashboard của vehicle",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Return the data for the vehicle dashboard
	res := vehicleData

	response := helper.SuccessResponse(res, "Vehicle dashboard data retrieved successfully", "Lấy thông tin dashboard của vehicle thành công")
	helper.GinResponse(ctx, 200, response)
}
