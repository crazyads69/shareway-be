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

	places, err := ctrl.MapsService.GetAutoComplete(ctx.Request.Context(), req.Input, req.Limit, req.Location, req.Radius, req.MoreCompound, req.CurrentLocation)
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
// @Param request body schemas.GiveRideRequest true "Give ride request details"
// @Success 200 {object} helper.Response{data=schemas.GiveRideResponse} "Successfully created route"
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
	route, rideOfferID, err := ctrl.MapsService.CreateGiveRide(ctx.Request.Context(), req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create route",
			"Không thể tạo tuyến đường",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Create a response with the route and ride offer ID
	res := schemas.GiveRideResponse{
		Route:       route,
		RideOfferID: rideOfferID,
	}

	response := helper.SuccessResponse(
		res,
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
// @Param request body schemas.HitchRideRequest true "Hitch ride request details"
// @Success 200 {object} helper.Response{data=schemas.HitchRideResponse} "Successfully created route"
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
	route, rideRequestID, err := ctrl.MapsService.CreateHitchRide(ctx.Request.Context(), req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create route",
			"Không thể tạo tuyến đường",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Create a response with the route and ride request ID
	res := schemas.HitchRideResponse{
		Route:         route,
		RideRequestID: rideRequestID,
	}

	response := helper.SuccessResponse(
		res,
		"Successfully created route",
		"Tạo tuyến đường thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// GetGeoCode returns the geocode information for a given point
// GetGeoCode godoc
// @Summary Get geocode data for a given point
// @Description Retrieves geocode information for a specified latitude and longitude
// @Tags map
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.GeoCodeRequest true "Geocode request parameters"
// @Success 200 {object} helper.Response{data=schemas.GeoCodeLocationResponse} "Successfully retrieved geocode data"
// @Failure 400 {object} helper.Response "Invalid request body"
// @Failure 500 {object} helper.Response "Failed to get geocode data"
// @Router /map/geocode [post]
func (ctrl *MapController) GetGeoCode(ctx *gin.Context) {
	var req schemas.GeoCodeRequest
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

	geocode, err := ctrl.MapsService.GetGeoCode(ctx.Request.Context(), req.Point)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get geocode data",
			"Không thể lấy dữ liệu geocode",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	optimizedResults := schemas.GeoCodeLocationResponse{
		Results: make([]schemas.GeoCodeLocation, 0, len(geocode.Results)),
	}

	for _, result := range geocode.Results {
		optimizedResult := schemas.GeoCodeLocation{
			PlaceID:          result.PlaceID,
			FormattedAddress: result.FormattedAddress,
			Latitude:         result.Geometry.Location.Lat,
			Longitude:        result.Geometry.Location.Lng,
		}
		optimizedResults.Results = append(optimizedResults.Results, optimizedResult)
	}

	response := helper.SuccessResponse(
		optimizedResults,
		"Successfully retrieved geocode data",
		"Lấy dữ liệu geocode thành công",
	)
	helper.GinResponse(ctx, 200, response)
}
