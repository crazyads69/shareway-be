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
	MapsService    service.IMapService
	validate       *validator.Validate
	VehicleService service.IVehicleService
	UserService    service.IUsersService
}

func NewMapController(mapsService service.IMapService, validate *validator.Validate, vehicleService service.IVehicleService, userService service.IUsersService) *MapController {
	return &MapController{
		MapsService:    mapsService,
		validate:       validate,
		VehicleService: vehicleService,
		UserService:    userService,
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
// @Param current_location query string false "Current location coordinates (lat,lng)"
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

	// Get the ride offer details
	rideOffer, err := ctrl.MapsService.GetRideOfferDetails(ctx.Request.Context(), rideOfferID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride offer details",
			"Không thể lấy thông tin chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get a list of waypoints for the ride offer
	waypoints, err := ctrl.MapsService.GetAllWaypoints(rideOfferID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get waypoints",
			"Không thể lấy danh sách điểm dừng",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get the vehicle details
	vehicle, err := ctrl.VehicleService.GetVehicleFromID(rideOffer.VehicleID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get vehicle details",
			"Không thể lấy thông tin xe",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	waypointDetails := make([]schemas.Waypoint, len(waypoints))
	for i, waypoint := range waypoints {
		waypointDetails[i] = schemas.Waypoint{
			ID:        waypoint.ID,
			Latitude:  waypoint.Latitude,
			Longitude: waypoint.Longitude,
			Address:   waypoint.Address,
			Order:     waypoint.Order,
		}
	}

	// Create a response with the route and ride offer ID
	res := schemas.GiveRideResponse{
		Route:       route,
		RideOfferID: rideOfferID,
		Distance:    rideOffer.Distance,
		Duration:    rideOffer.Duration,
		StartTime:   rideOffer.StartTime,
		EndTime:     rideOffer.EndTime,
		Fare:        rideOffer.Fare,
		Vehicle:     vehicle,
		Waypoints:   waypointDetails,
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

	// Get the ride request details
	rideRequest, err := ctrl.MapsService.GetRideRequestDetails(ctx.Request.Context(), rideRequestID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride request details",
			"Không thể lấy thông tin chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Create a response with the route and ride request ID
	res := schemas.HitchRideResponse{
		Route:         route,
		RideRequestID: rideRequestID,
		Distance:      rideRequest.Distance,
		Duration:      rideRequest.Duration,
		StartTime:     rideRequest.StartTime,
		EndTime:       rideRequest.EndTime,
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

	optimizedResults, err := ctrl.MapsService.GetGeoCode(ctx.Request.Context(), req.Point, req.CurrentLocation)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get geocode data",
			"Không thể lấy dữ liệu geocode",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	response := helper.SuccessResponse(
		optimizedResults,
		"Successfully retrieved geocode data",
		"Lấy dữ liệu geocode thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// SuggestHitchRides returns a list of ride requests that match the business rules for the rider (ride offer)
// SuggestHitchRides godoc
// @Summary Suggest ride requests for a rider (ride offer)
// @Description Returns a list of ride requests that match the business rules for the rider (ride offer)
// @Tags map
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.SuggestRideRequestRequest true "Ride offer ID"
// @Success 200 {object} helper.Response{data=schemas.SuggestRideRequestResponse} "Successfully retrieved suggested ride requests"
// @Failure 400 {object} helper.Response "Invalid request body"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /map/suggest-hitch-rides [post]
func (ctrl *MapController) SuggestHitchRides(ctx *gin.Context) {
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

	var req schemas.SuggestRideRequestRequest

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

	// Get the ride requests that match the business rules
	rideRequests, err := ctrl.MapsService.SuggestRideRequests(ctx.Request.Context(), data.UserID, req.RideOfferID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get suggested ride requests",
			"Không thể lấy danh sách chuyến đi gợi ý",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Convert the ride requests to the RideRequestDetail
	var rideRequestDetails []schemas.RideRequestDetail
	for _, rideRequest := range rideRequests {
		// Get the user details
		user, err := ctrl.UserService.GetUserByID(rideRequest.UserID)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to get user details",
				"Không thể lấy thông tin người dùng",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}
		rideRequestDetail := schemas.RideRequestDetail{
			ID: rideRequest.ID,
			User: schemas.UserInfo{
				ID:          user.ID,
				FullName:    user.FullName,
				PhoneNumber: user.PhoneNumber,
				AvatarURL:   user.AvatarURL,
				Gender:      user.Gender,
			},
			EncodedPolyline:       string(rideRequest.EncodedPolyline),
			Distance:              rideRequest.Distance,
			Duration:              rideRequest.Duration,
			StartTime:             rideRequest.StartTime,
			EndTime:               rideRequest.EndTime,
			StartLatitude:         rideRequest.StartLatitude,
			StartLongitude:        rideRequest.StartLongitude,
			EndLatitude:           rideRequest.EndLatitude,
			EndLongitude:          rideRequest.EndLongitude,
			StartAddress:          rideRequest.StartAddress,
			EndAddress:            rideRequest.EndAddress,
			RiderCurrentLatitude:  rideRequest.RiderCurrentLatitude,
			RiderCurrentLongitude: rideRequest.RiderCurrentLongitude,
		}
		rideRequestDetails = append(rideRequestDetails, rideRequestDetail)
	}

	res := schemas.SuggestRideRequestResponse{
		RideRequests: rideRequestDetails,
	}

	response := helper.SuccessResponse(
		res,
		"Successfully retrieved suggested ride requests",
		"Lấy danh sách chuyến đi gợi ý thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// SuggestGiveRides returns a list of ride offers that match the business rules for the hitcher (ride request)
// SuggestGiveRides godoc
// @Summary Suggest ride offers for a hitcher
// @Description Returns a list of ride offers that match the business rules for the hitcher (ride request)
// @Tags map
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.SuggestRideOfferRequest true "Ride request details"
// @Success 200 {object} helper.Response{data=schemas.SuggestRideOfferResponse} "Successfully retrieved suggested ride offers"
// @Failure 400 {object} helper.Response "Invalid request body"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /map/suggest-give-rides [post]
func (ctrl *MapController) SuggestGiveRides(ctx *gin.Context) {
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

	var req schemas.SuggestRideOfferRequest

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

	// Get the ride offers that match the business rules
	rideOffers, err := ctrl.MapsService.SuggestRideOffers(ctx.Request.Context(), data.UserID, req.RideRequestID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get suggested ride offers",
			"Không thể lấy danh sách chuyến đi gợi ý",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Convert the ride offers to the RideOfferDetail
	var rideOfferDetails []schemas.RideOfferDetail
	for _, rideOffer := range rideOffers {
		// Get the user details
		user, err := ctrl.UserService.GetUserByID(rideOffer.UserID)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to get user details",
				"Không thể lấy thông tin người dùng",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}
		// Get the vehicle details
		vehicle, err := ctrl.VehicleService.GetVehicleFromID(rideOffer.VehicleID)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to get vehicle details",
				"Không thể lấy thông tin xe",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}
		// Get a list of waypoints for the ride offer
		waypoints, err := ctrl.MapsService.GetAllWaypoints(rideOffer.ID)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to get waypoints",
				"Không thể lấy danh sách điểm dừng",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}
		waypointDetails := make([]schemas.Waypoint, len(waypoints))
		for i, waypoint := range waypoints {
			waypointDetails[i] = schemas.Waypoint{
				ID:        waypoint.ID,
				Latitude:  waypoint.Latitude,
				Longitude: waypoint.Longitude,
				Address:   waypoint.Address,
				Order:     waypoint.Order,
			}
		}
		rideOfferDetail := schemas.RideOfferDetail{
			ID: rideOffer.ID,
			User: schemas.UserInfo{
				ID:          user.ID,
				FullName:    user.FullName,
				PhoneNumber: user.PhoneNumber,
				AvatarURL:   user.AvatarURL,
				Gender:      user.Gender,
			},
			Vehicle:                vehicle,
			EncodedPolyline:        string(rideOffer.EncodedPolyline),
			Distance:               rideOffer.Distance,
			Duration:               rideOffer.Duration,
			StartTime:              rideOffer.StartTime,
			EndTime:                rideOffer.EndTime,
			StartLatitude:          rideOffer.StartLatitude,
			StartLongitude:         rideOffer.StartLongitude,
			EndLatitude:            rideOffer.EndLatitude,
			EndLongitude:           rideOffer.EndLongitude,
			StartAddress:           rideOffer.StartAddress,
			EndAddress:             rideOffer.EndAddress,
			DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
			DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
			Status:                 rideOffer.Status,
			Fare:                   rideOffer.Fare,
			Waypoints:              waypointDetails,
		}
		// Append the ride offer detail to the list
		rideOfferDetails = append(rideOfferDetails, rideOfferDetail)
	}

	res := schemas.SuggestRideOfferResponse{
		RideOffers: rideOfferDetails,
	}

	response := helper.SuccessResponse(
		res,
		"Successfully retrieved suggested ride offers",
		"Lấy danh sách chuyến đi gợi ý thành công",
	)
	helper.GinResponse(ctx, 200, response)
}
