package controller

import (
	"shareway/helper"
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
// @Param Authorization header string true "Bearer <access_token>"
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
