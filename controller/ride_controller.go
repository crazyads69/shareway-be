package controller

import (
	"fmt"
	"shareway/helper"
	"shareway/infra/ws"
	"shareway/middleware"
	"shareway/schemas"
	"shareway/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type RideController struct {
	validate       *validator.Validate
	hub            *ws.Hub
	RideService    service.IRideService
	MapsService    service.IMapService
	UserService    service.IUsersService
	VehicleService service.IVehicleService
}

func NewRideController(validate *validator.Validate, hub *ws.Hub, rideService service.IRideService,
	mapService service.IMapService, userService service.IUsersService, vehicleService service.IVehicleService) *RideController {
	return &RideController{
		validate:       validate,
		hub:            hub,
		RideService:    rideService,
		MapsService:    mapService,
		UserService:    userService,
		VehicleService: vehicleService,
	}
}

// SendGiveRideRequest sends a ride offer request from the driver to the hitcher
// SendGiveRideRequest godoc
// @Summary Send a ride offer request from the driver to the hitcher
// @Description Send a ride offer request from the driver to the hitcher
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.SendGiveRideRequestRequest true "Give ride request details"
// @Success 200 {object} helper.Response "Successfully sent ride offer request"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/give-ride-request [post]
func (ctrl *RideController) SendGiveRideRequest(ctx *gin.Context) {
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

	var req schemas.SendGiveRideRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind JSON",
			"Không thể bind JSON",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := ctrl.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Get user details from user_id
	user, err := ctrl.UserService.GetUserByID(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get user details",
			"Không thể lấy thông tin người dùng",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get vehicle details from user_id
	vehicle, err := ctrl.VehicleService.GetVehicleFromID(req.VehicleID)

	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get vehicle details",
			"Không thể lấy thông tin phương tiện",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get ride offer details from ride_offer_id
	rideOffer, err := ctrl.RideService.GetRideOfferByID(req.RideOfferID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride offer details",
			"Không thể lấy thông tin chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.SendGiveRideRequestResponse{
		ID: rideOffer.ID,
		User: schemas.UserInfo{
			ID:          user.ID,
			PhoneNumber: user.PhoneNumber,
			FullName:    user.FullName,
		},
		Vehicle:                vehicle,
		StartLatitude:          rideOffer.StartLatitude,
		StartLongitude:         rideOffer.StartLongitude,
		EndLatitude:            rideOffer.EndLatitude,
		EndLongitude:           rideOffer.EndLongitude,
		StartAddress:           rideOffer.StartAddress,
		EndAddress:             rideOffer.EndAddress,
		EncodedPolyline:        rideOffer.EncodedPolyline,
		Distance:               rideOffer.Distance,
		Duration:               rideOffer.Duration,
		DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
		DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
		StartTime:              rideOffer.StartTime,
		EndTime:                rideOffer.EndTime,
		Status:                 rideOffer.Status,
		Fare:                   rideOffer.Fare,
	}

	// Send ride offer request to the receiver
	ctrl.hub.SendToUser(req.ReceiverID.String(), "new-give-ride-request", res)

	// Return success response
	helper.GinResponse(ctx, 200, helper.SuccessResponse(
		nil,
		"Successfully sent ride offer request",
		"Đã gửi yêu cầu chia sẻ chuyến đi thành công",
	))

}

// SendHitchRideRequest sends a ride request from the hitcher to the driver
// SendHitchRideRequest godoc
// @Summary Send a ride request from the hitcher to the driver
// @Description Send a ride request from the hitcher to the driver
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.SendHitchRideRequestRequest true "Hitch ride request details"
// @Success 200 {object} helper.Response "Successfully sent ride request"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/hitch-ride-request [post]
func (ctrl *RideController) SendHitchRideRequest(ctx *gin.Context) {
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

	var req schemas.SendHitchRideRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind JSON",
			"Không thể bind JSON",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := ctrl.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Get user details from user_id
	user, err := ctrl.UserService.GetUserByID(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get user details",
			"Không thể lấy thông tin người dùng",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get ride request details from ride_request_id
	rideRequest, err := ctrl.RideService.GetRideRequestByID(req.RideRequestID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride request details",
			"Không thể lấy thông tin yêu cầu chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.SendHitchRideRequestResponse{
		ID: rideRequest.ID,
		User: schemas.UserInfo{
			ID:          user.ID,
			PhoneNumber: user.PhoneNumber,
			FullName:    user.FullName,
		},
		StartLatitude:         rideRequest.StartLatitude,
		StartLongitude:        rideRequest.StartLongitude,
		EndLatitude:           rideRequest.EndLatitude,
		EndLongitude:          rideRequest.EndLongitude,
		StartAddress:          rideRequest.StartAddress,
		EndAddress:            rideRequest.EndAddress,
		RiderCurrentLatitude:  rideRequest.RiderCurrentLatitude,
		RiderCurrentLongitude: rideRequest.RiderCurrentLongitude,
		Status:                rideRequest.Status,
		EncodedPolyline:       rideRequest.EncodedPolyline,
		Distance:              rideRequest.Distance,
		Duration:              rideRequest.Duration,
		StartTime:             rideRequest.StartTime,
		EndTime:               rideRequest.EndTime,
	}

	// Send ride request to the receiver
	ctrl.hub.SendToUser(req.ReceiverID.String(), "new-hitch-ride-request", res)

	// Return success response
	helper.GinResponse(ctx, 200, helper.SuccessResponse(
		nil,
		"Successfully sent ride request",
		"Đã gửi yêu cầu chuyến đi thành công",
	))
}

// AcceptGiveRideRequest accepts a ride offer request from the driver (the hitcher accepted the driver's offer)
// AcceptGiveRideRequest godoc
// @Summary Accept a ride offer request from the driver
// @Description Accept a ride offer request from the driver
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.AcceptGiveRideRequestRequest true "Accept give ride request details"
// @Success 200 {object} helper.Response{data=schemas.AcceptGiveRideRequestResponse} "Successfully accepted ride offer request"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/accept-give-ride-request [post]
func (ctrl *RideController) AcceptGiveRideRequest(ctx *gin.Context) {
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

	var req schemas.AcceptGiveRideRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind JSON",
			"Không thể bind JSON",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := ctrl.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Create ride between driver and hitcher (because the hitcher accepted the ride offer from the driver means ride is engaged)
	ride, err := ctrl.RideService.AcceptRideRequest(req.RideOfferID, req.RideRequestID, req.VehicleID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create ride",
			"Không thể tạo chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Create a transaction to store fare details
	transaction, err := ctrl.RideService.CreateRideTransaction(ride.ID, ride.Fare, req.ReceiverID, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create transaction",
			"Không thể tạo giao dịch",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	vehicle, err := ctrl.VehicleService.GetVehicleFromID(req.VehicleID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get vehicle details",
			"Không thể lấy thông tin phương tiện",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.AcceptGiveRideRequestResponse{
		ID:          ride.ID,
		RideOfferID: ride.RideOfferID,
		Transaction: schemas.TransactionDetail{
			ID:            transaction.ID,
			Amount:        transaction.Amount,
			Status:        transaction.Status,
			PaymentMethod: transaction.PaymentMethod,
		},
		Status:          ride.Status,
		StartTime:       ride.StartTime,
		EndTime:         ride.EndTime,
		StartAddress:    ride.StartAddress,
		EndAddress:      ride.EndAddress,
		Fare:            ride.Fare,
		EncodedPolyline: ride.EncodedPolyline,
		Distance:        ride.Distance,
		Duration:        ride.Duration,
		StartLatitude:   ride.StartLatitude,
		StartLongitude:  ride.StartLongitude,
		EndLatitude:     ride.EndLatitude,
		EndLongitude:    ride.EndLongitude,
		Vehicle:         vehicle,
	}

	// Send the accepted ride offer to the driver (match the ride successfully)
	ctrl.hub.SendToUser(req.ReceiverID.String(), "accept-give-ride-request", res)

	// Return success response
	helper.GinResponse(ctx, 200, helper.SuccessResponse(
		res,
		"Successfully accepted ride offer request",
		"Đã chấp nhận yêu cầu chia sẻ chuyến đi thành công",
	))
}

// AcceptHitchRideRequest accepts a ride request from the hitcher (the driver accepted the hitcher's request)
// AcceptHitchRideRequest godoc
// @Summary Accept a ride request from the hitcher
// @Description Accept a ride request from the hitcher
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.AcceptHitchRideRequestRequest true "Accept hitch ride request details"
// @Success 200 {object} helper.Response{data=schemas.AcceptHitchRideRequestResponse} "Successfully accepted ride request"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/accept-hitch-ride-request [post]
func (ctrl *RideController) AcceptHitchRideRequest(ctx *gin.Context) {
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

	var req schemas.AcceptHitchRideRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind JSON",
			"Không thể bind JSON",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}
	if err := ctrl.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Create ride between driver and hitcher (because the driver accepted the ride request from the hitcher means ride is engaged)
	ride, err := ctrl.RideService.AcceptRideRequest(req.RideOfferID, req.RideRequestID, req.VehicleID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create ride",
			"Không thể tạo chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Create a transaction to store fare details
	transaction, err := ctrl.RideService.CreateRideTransaction(ride.ID, ride.Fare, data.UserID, req.ReceiverID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create transaction",
			"Không thể tạo giao dịch",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	vehicle, err := ctrl.VehicleService.GetVehicleFromID(req.VehicleID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get vehicle details",
			"Không thể lấy thông tin phương tiện",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.AcceptHitchRideRequestResponse{
		ID:            ride.ID,
		RideRequestID: ride.RideRequestID,
		Transaction: schemas.TransactionDetail{
			ID:            transaction.ID,
			Amount:        transaction.Amount,
			Status:        transaction.Status,
			PaymentMethod: transaction.PaymentMethod,
		},
		Status:          ride.Status,
		StartTime:       ride.StartTime,
		EndTime:         ride.EndTime,
		StartAddress:    ride.StartAddress,
		EndAddress:      ride.EndAddress,
		Fare:            ride.Fare,
		EncodedPolyline: ride.EncodedPolyline,
		Distance:        ride.Distance,
		Duration:        ride.Duration,
		StartLatitude:   ride.StartLatitude,
		StartLongitude:  ride.StartLongitude,
		EndLatitude:     ride.EndLatitude,
		EndLongitude:    ride.EndLongitude,
		Vehicle:         vehicle,
	}

	// Send the accepted ride request to the hitcher (match the ride successfully)
	ctrl.hub.SendToUser(req.ReceiverID.String(), "accept-hitch-ride-request", res)

	// Return success response
	helper.GinResponse(ctx, 200, helper.SuccessResponse(
		res,
		"Successfully accepted ride request",
		"Đã chấp nhận yêu cầu chia sẻ chuyến đi thành công",
	))
}

// CancelGiveRideRequest cancels a ride offer request from the driver (the hitcher cancels the request)
// CancelGiveRideRequest godoc
// @Summary Cancel a ride offer request from the driver
// @Description Cancel a ride offer request from the driver
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.CancelGiveRideRequestRequest true "Cancel give ride request details"
// @Success 200 {object} helper.Response "Successfully canceled ride offer request"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/cancel-give-ride-request [post]
func (ctrl *RideController) CancelGiveRideRequest(ctx *gin.Context) {
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

	var req schemas.CancelGiveRideRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind JSON",
			"Không thể bind JSON",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}
	if err := ctrl.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	res := schemas.CancelGiveRideRequestResponse{
		RideOfferID:   req.RideOfferID,
		RideRequestID: req.RideRequestID,
		UserID:        data.UserID,
	}
	// Send the cancel notification to the driver
	ctrl.hub.SendToUser(req.ReceiverID.String(), "cancel-give-ride-request", res)

	// Return success response
	helper.GinResponse(ctx, 200, helper.SuccessResponse(
		nil,
		"Successfully canceled ride offer request",
		"Đã hủy yêu cầu chia sẻ chuyến đi thành công",
	))
}

// CancelHitchRideRequest cancels a ride request from the hitcher (the driver cancels the request)
// CancelHitchRideRequest godoc
// @Summary Cancel a ride request from the hitcher
// @Description Cancel a ride request from the hitcher
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.CancelHitchRideRequestRequest true "Cancel hitch ride request details"
// @Success 200 {object} helper.Response "Successfully canceled ride request"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/cancel-hitch-ride-request [post]
func (ctrl *RideController) CancelHitchRideRequest(ctx *gin.Context) {
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

	var req schemas.CancelHitchRideRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind JSON",
			"Không thể bind JSON",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}
	if err := ctrl.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	res := schemas.CancelHitchRideRequestResponse{
		RideOfferID:   req.RideOfferID,
		RideRequestID: req.RideRequestID,
		UserID:        data.UserID,
	}
	// Send the cancel notification to the hitcher
	ctrl.hub.SendToUser(req.ReceiverID.String(), "cancel-hitch-ride-request", res)

	// Return success response
	helper.GinResponse(ctx, 200, helper.SuccessResponse(
		nil,
		"Successfully canceled ride request",
		"Đã hủy yêu cầu chia sẻ chuyến đi thành công",
	))
}
