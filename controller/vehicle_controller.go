package controller

import (
	"fmt"
	"shareway/helper"
	"shareway/middleware"
	"shareway/schemas"
	"shareway/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type VehicleController struct {
	service  service.IVehicleService
	validate *validator.Validate
}

func NewVehicleController(service service.IVehicleService, validate *validator.Validate) *VehicleController {
	return &VehicleController{
		service:  service,
		validate: validate,
	}
}

// GetVehicles retrieves and returns the list of vehicles for user selected their vehicle when register vehicle
// GetVehicles godoc
// @Summary Get list of vehicles
// @Description Retrieves and returns the list of vehicles for user to select when registering a vehicle
// @Tags vehicle
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.Response{data=schemas.GetVehiclesResponse} "Successfully retrieved vehicles"
// @Failure 500 {object} helper.Response "Failed to get vehicles"
// @Router /vehicle/vehicles [get]
func (ctrl *VehicleController) GetVehicles(ctx *gin.Context) {
	// Get list of vehicles from database
	vehicles, err := ctrl.service.GetVehicles()
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get vehicles",
			"Không thể lấy danh sách xe",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.GetVehiclesResponse{
		Vehicles: vehicles,
	}

	response := helper.SuccessResponse(res, "Successfully retrieved vehicles", "Đã lấy danh sách xe thành công")
	helper.GinResponse(ctx, 200, response)

}

// RegisterVehicle registers a new vehicle for the user in the database
// RegisterVehicle godoc
// @Summary Register a new vehicle
// @Description Register a new vehicle for the authenticated user
// @Tags vehicle
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.RegisterVehicleRequest true "Vehicle registration details"
// @Success 200 {object} helper.Response "Successfully registered vehicle"
// @Failure 400 {object} helper.Response "Bad request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /vehicle/register-vehicle [post]
func (ctrl *VehicleController) RegisterVehicle(ctx *gin.Context) {
	// Fetch the user ID from the payload from middleware
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert payload to map
	data, err := helper.ConvertToPayload(payload)

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

	// Bind the request body to the RegisterVehicleRequest schema
	var req schemas.RegisterVehicleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to bind request"),
			"Failed to bind request",
			"Không thể bind request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate the request body
	if err := ctrl.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Check if the license plate already exists
	exists, err := ctrl.service.LicensePlateExists(req.LicensePlate)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to check if license plate exists",
			"Không thể kiểm tra biển số xe đã tồn tại",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}
	if exists {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("license plate already exists"),
			"License plate already exists",
			"Biển số xe đã tồn tại",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Check if the Cavet already exists
	exists, err = ctrl.service.CaVetExists(req.CaVet)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to check if Cavet exists",
			"Không thể kiểm tra Cavet đã tồn tại",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}
	if exists {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("Cavet already exists"),
			"Cavet already exists",
			"Cavet đã tồn tại",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Register the vehicle
	err = ctrl.service.RegisterVehicle(data.UserID, req.VehicleID, req.LicensePlate, req.CaVet)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to register vehicle",
			"Không thể đăng ký xe",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	response := helper.SuccessResponse(nil, "Successfully registered vehicle", "Đăng ký xe thành công")
	helper.GinResponse(ctx, 200, response)
}
