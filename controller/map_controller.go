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

type MapController struct {
	MapsService service.IMapService
	validate    *validator.Validate
}

func NewMapController(mapsService service.IMapService, validate *validator.Validate) *MapController {
	return &MapController{
		MapsService: mapsService,
		validate:    validate,
	}
}

// GetAutoComplete returns a list of places that match the query string
// GetAutoComplete godoc
// @Summary Get autocomplete suggestions for places
// @Description Returns a list of places that match the query string
// @Tags map
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer <access_token>"
// @Param input query string true "Input string to search for"
// @Param limit query int false "Limit the number of results"
// @Param location query string false "Location coordinates (lat,lng)"
// @Param radius query int false "Search radius in meters"
// @Param more_compound query bool false "Include more compound results"
// @Success 200 {object} helper.Response{data=schemas.GoongAutoCompleteResponse} "Successfully retrieved autocomplete data"
// @Failure 400 {object} helper.Response "Invalid request query"
// @Failure 500 {object} helper.Response "Failed to get autocomplete data"
// @Router /map/autocomplete [get]
func (ctrl *MapController) GetAutoComplete(ctx *gin.Context) {

	var req schemas.AutoCompleteRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request query",
			"Câu truy vấn không hợp lệ",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	places, err := ctrl.MapsService.GetAutoComplete(ctx.Request.Context(), req.Input, req.Limit, req.Location, req.Radius, req.MoreCompound)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get autocomplete data",
			"Không thể lấy dữ liệu gợi ý",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	response := helper.SuccessResponse(
		places,
		"Successfully retrieved autocomplete data",
		"Lấy dữ liệu gợi ý thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// CreateGiveRide receives a list of points and returns a route and polyline encoded string for the driver
// CreateGiveRide godoc
// @Summary Create a route for a driver's give ride
// @Description Receives a list of points and returns a route and polyline encoded string for the driver
// @Tags map
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer <access_token>"
// @Param request body schemas.GiveRideRequest true "Give ride request details"
// @Success 200 {object} helper.Response{data=schemas.GoongDirectionsResponse} "Successfully created route"
// @Failure 400 {object} helper.Response "Invalid request body"
// @Failure 500 {object} helper.Response "Failed to create route"
// @Router /map/give-ride [post]
func (ctrl *MapController) CreateGiveRide(ctx *gin.Context) {
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
	var req schemas.GiveRideRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Dữ liệu không hợp lệ",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate the request body
	if err := ctrl.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Dữ liệu không hợp lệ",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Create a route for the driver
	route, err := ctrl.MapsService.CreateGiveRide(ctx.Request.Context(), req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create route",
			"Không thể tạo tuyến đường",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	response := helper.SuccessResponse(
		route,
		"Successfully created route",
		"Tạo tuyến đường thành công",
	)
	helper.GinResponse(ctx, 200, response)

}

// CreateHitchRide receives a list of points and returns a route and polyline encoded string for the hitcher
// CreateHitchRide godoc
// @Summary Create a route for a passenger's hitch ride
// @Description Receives a list of points and returns a route and polyline encoded string for the passenger
// @Tags map
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer <access_token>"
// @Param request body schemas.HitchRideRequest true "Hitch ride request details"
// @Success 200 {object} helper.Response{data=schemas.GoongDirectionsResponse} "Successfully created route"
// @Failure 400 {object} helper.Response "Invalid request body"
// @Failure 500 {object} helper.Response "Failed to create route"
// @Router /map/hitch-ride [post]
func (ctrl *MapController) CreateHitchRide(ctx *gin.Context) {
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
	var req schemas.HitchRideRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Dữ liệu không hợp lệ",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate the request body
	if err := ctrl.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Dữ liệu không hợp lệ",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Create a route for the hitcher
	route, err := ctrl.MapsService.CreateHitchRide(ctx.Request.Context(), req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create route",
			"Không thể tạo tuyến đường",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	response := helper.SuccessResponse(
		route,
		"Successfully created route",
		"Tạo tuyến đường thành công",
	)
	helper.GinResponse(ctx, 200, response)
}
