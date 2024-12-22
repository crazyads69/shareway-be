package controller

import (
	"fmt"
	"log"

	"shareway/helper"
	"shareway/infra/task"
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
	PaymentService service.IPaymentService
	asyncClient    *task.AsyncClient
}

func NewRideController(validate *validator.Validate, hub *ws.Hub, rideService service.IRideService,
	mapService service.IMapService, userService service.IUsersService, vehicleService service.IVehicleService, paymentService service.IPaymentService,
	asyncClient *task.AsyncClient) *RideController {
	return &RideController{
		validate:       validate,
		hub:            hub,
		RideService:    rideService,
		MapsService:    mapService,
		UserService:    userService,
		VehicleService: vehicleService,
		PaymentService: paymentService,
		asyncClient:    asyncClient,
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

	// Get waypoints details from ride_offer_id
	waypoints, err := ctrl.MapsService.GetAllWaypoints(rideOffer.ID)
	if err != nil {
		helper.GinResponse(ctx, 500, helper.ErrorResponseWithMessage(
			err,
			"Failed to get waypoints",
			"Không thể lấy thông tin waypoints",
		))
		return
	}

	var waypointDetails []schemas.Waypoint
	if waypoints != nil {
		waypointDetails = make([]schemas.Waypoint, 0, len(waypoints))
		for _, waypoint := range waypoints {
			waypointDetails = append(waypointDetails, schemas.Waypoint{
				Latitude:  waypoint.Latitude,
				Longitude: waypoint.Longitude,
				Address:   waypoint.Address,
				ID:        waypoint.ID,
				Order:     waypoint.WaypointOrder,
			})
		}
	}

	res := schemas.SendGiveRideRequestResponse{
		ID: rideOffer.ID,
		User: schemas.UserInfo{
			ID:            user.ID,
			FullName:      user.FullName,
			PhoneNumber:   user.PhoneNumber,
			AvatarURL:     user.AvatarURL,
			Gender:        user.Gender,
			IsMomoLinked:  user.IsMomoLinked,
			BalanceInApp:  user.BalanceInApp,
			AverageRating: user.AverageRating,
		},
		Vehicle:                vehicle,
		StartLatitude:          rideOffer.StartLatitude,
		StartLongitude:         rideOffer.StartLongitude,
		EndLatitude:            rideOffer.EndLatitude,
		EndLongitude:           rideOffer.EndLongitude,
		StartAddress:           rideOffer.StartAddress,
		EndAddress:             rideOffer.EndAddress,
		EncodedPolyline:        string(rideOffer.EncodedPolyline),
		Distance:               rideOffer.Distance,
		Duration:               rideOffer.Duration,
		DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
		DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
		StartTime:              rideOffer.StartTime,
		EndTime:                rideOffer.EndTime,
		Status:                 rideOffer.Status,
		Fare:                   rideOffer.Fare,
		ReceiverID:             req.ReceiverID,
		RideRequestID:          req.RideRequestID,
		Waypoints:              waypointDetails,
	}

	// Send ride offer request to the receiver
	// ctrl.hub.SendToUser(req.ReceiverID.String(), "new-give-ride-request", res)

	// Get receiver device token to send notification
	receiver, err := ctrl.UserService.GetUserByID(req.ReceiverID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get receiver details",
			"Không thể lấy thông tin người nhận",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the WebSocket message
	wsMessage := schemas.WebSocketMessage{
		UserID:  req.ReceiverID.String(),
		Type:    "new-give-ride-request",
		Payload: res,
	}

	// Convert res to map[string]string
	resMap, err := helper.ConvertToStringMap(res)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Append type to the notification payload
	notificationPayload := schemas.NotificationPayload{
		Type: "new-give-ride-request",
		Data: resMap,
	}

	// Convert notificationPayload to map[string]string
	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the notification message
	notification := schemas.Notification{
		Title: "Bạn nhận được một lời mời đi nhờ mới",
		Body:  "Bạn nhận được một lời mời đi nhờ mới, hãy xem chi tiết và chấp nhận hoặc từ chối",
		Token: receiver.DeviceToken,
		Data:  notificationPayloadMap,
	}

	// Send the WebSocket message
	go func() {
		err := ctrl.asyncClient.EnqueueWebsocketMessage(wsMessage)
		if err != nil {
			log.Printf("Failed to enqueue websocket message: %v", err)
		}
	}()

	// Send the notification message
	go func() {
		err = ctrl.asyncClient.EnqueueFCMNotification(notification)
		if err != nil {
			log.Printf("Failed to enqueue FCM notification: %v", err)
		}
	}()

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

	// Get vehicle details from user_id
	vehicle, err := ctrl.VehicleService.GetVehicleFromID(rideOffer.VehicleID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get vehicle details",
			"Không thể lấy thông tin phương tiện",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.SendHitchRideRequestResponse{
		ID: rideRequest.ID,
		User: schemas.UserInfo{
			ID:            user.ID,
			FullName:      user.FullName,
			PhoneNumber:   user.PhoneNumber,
			AvatarURL:     user.AvatarURL,
			Gender:        user.Gender,
			IsMomoLinked:  user.IsMomoLinked,
			BalanceInApp:  user.BalanceInApp,
			AverageRating: user.AverageRating,
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
		EncodedPolyline:       string(rideRequest.EncodedPolyline),
		Distance:              rideRequest.Distance,
		Duration:              rideRequest.Duration,
		StartTime:             rideRequest.StartTime,
		EndTime:               rideRequest.EndTime,
		ReceiverID:            req.ReceiverID,
		RideOfferID:           req.RideOfferID,
		Vehicle:               vehicle,
	}

	// Send ride request to the receiver
	// ctrl.hub.SendToUser(req.ReceiverID.String(), "new-hitch-ride-request", res)

	// Get receiver device token to send notification
	receiver, err := ctrl.UserService.GetUserByID(req.ReceiverID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get receiver details",
			"Không thể lấy thông tin người nhận",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the WebSocket message
	wsMessage := schemas.WebSocketMessage{
		UserID:  req.ReceiverID.String(),
		Type:    "new-hitch-ride-request",
		Payload: res,
	}

	// Convert res to map[string]string
	resMap, err := helper.ConvertToStringMap(res)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	notificationPayload := schemas.NotificationPayload{
		Type: "new-hitch-ride-request",
		Data: resMap,
	}

	// Convert notificationPayload to map[string]string
	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the notification message
	notification := schemas.Notification{
		Title: "Bạn nhận được một yêu cầu đi nhờ mới",
		Body:  "Bạn nhận được một yêu cầu đi nhờ mới, hãy xem chi tiết và chấp nhận hoặc từ chối",
		Token: receiver.DeviceToken,
		Data:  notificationPayloadMap,
	}

	// Send the WebSocket message using the async client
	go func() {
		err := ctrl.asyncClient.EnqueueWebsocketMessage(wsMessage)
		if err != nil {
			log.Printf("Failed to enqueue websocket message: %v", err)
		}
	}()

	// Send the notification message using the async client
	go func() {
		err = ctrl.asyncClient.EnqueueFCMNotification(notification)
		if err != nil {
			log.Printf("Failed to enqueue FCM notification: %v", err)
		}
	}()

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

	waypoints, err := ctrl.MapsService.GetAllWaypoints(rideOffer.ID)
	if err != nil {
		helper.GinResponse(ctx, 500, helper.ErrorResponseWithMessage(
			err,
			"Failed to get waypoints",
			"Không thể lấy thông tin waypoints",
		))
		return
	}

	var waypointDetails []schemas.Waypoint
	if waypoints != nil {
		waypointDetails = make([]schemas.Waypoint, 0, len(waypoints))
		for _, waypoint := range waypoints {
			waypointDetails = append(waypointDetails, schemas.Waypoint{
				Latitude:  waypoint.Latitude,
				Longitude: waypoint.Longitude,
				Address:   waypoint.Address,
				ID:        waypoint.ID,
				Order:     waypoint.WaypointOrder,
			})
		}
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

	// Create a transaction to store fare details
	transaction, err := ctrl.RideService.CreateRideTransaction(ride.ID, ride.Fare, "cash", req.ReceiverID, data.UserID)
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

	// Get receiver device token to send notification
	receiver, err := ctrl.UserService.GetUserByID(req.ReceiverID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get receiver details",
			"Không thể lấy thông tin người nhận",
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
		DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
		DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
		RiderCurrentLatitude:   rideRequest.RiderCurrentLatitude,
		RiderCurrentLongitude:  rideRequest.RiderCurrentLongitude,
		Status:                 ride.Status,
		StartTime:              ride.StartTime,
		EndTime:                ride.EndTime,
		StartAddress:           ride.StartAddress,
		EndAddress:             ride.EndAddress,
		Fare:                   ride.Fare,
		EncodedPolyline:        string(ride.EncodedPolyline),
		Distance:               ride.Distance,
		Duration:               ride.Duration,
		StartLatitude:          ride.StartLatitude,
		StartLongitude:         ride.StartLongitude,
		EndLatitude:            ride.EndLatitude,
		EndLongitude:           ride.EndLongitude,
		Vehicle:                vehicle,
		ReceiverID:             req.ReceiverID,
		UserInfo: schemas.UserInfo{
			ID:            receiver.ID,
			PhoneNumber:   receiver.PhoneNumber,
			FullName:      receiver.FullName,
			AvatarURL:     receiver.AvatarURL,
			Gender:        receiver.Gender,
			IsMomoLinked:  receiver.IsMomoLinked,
			BalanceInApp:  receiver.BalanceInApp,
			AverageRating: receiver.AverageRating,
		},
		RideRequestID: req.RideRequestID,
		Waypoints:     waypointDetails,
	}

	// Send the accepted ride offer to the driver (match the ride successfully)
	// ctrl.hub.SendToUser(req.ReceiverID.String(), "accept-give-ride-request", res)

	// Get accepter user details from user_id
	accepter, err := ctrl.UserService.GetUserByID(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get accepter details",
			"Không thể lấy thông tin người chấp nhận",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	wsRes := schemas.AcceptGiveRideRequestResponse{
		ID:          ride.ID,
		RideOfferID: ride.RideOfferID,
		Transaction: schemas.TransactionDetail{
			ID:            transaction.ID,
			Amount:        transaction.Amount,
			Status:        transaction.Status,
			PaymentMethod: transaction.PaymentMethod,
		},
		DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
		DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
		RiderCurrentLatitude:   rideRequest.RiderCurrentLatitude,
		RiderCurrentLongitude:  rideRequest.RiderCurrentLongitude,
		Status:                 ride.Status,
		StartTime:              ride.StartTime,
		EndTime:                ride.EndTime,
		StartAddress:           ride.StartAddress,
		EndAddress:             ride.EndAddress,
		Fare:                   ride.Fare,
		EncodedPolyline:        string(ride.EncodedPolyline),
		Distance:               ride.Distance,
		Duration:               ride.Duration,
		StartLatitude:          ride.StartLatitude,
		StartLongitude:         ride.StartLongitude,
		EndLatitude:            ride.EndLatitude,
		EndLongitude:           ride.EndLongitude,
		Vehicle:                vehicle,
		ReceiverID:             req.ReceiverID,
		UserInfo: schemas.UserInfo{
			ID:            accepter.ID,
			PhoneNumber:   accepter.PhoneNumber,
			FullName:      accepter.FullName,
			AvatarURL:     accepter.AvatarURL,
			Gender:        accepter.Gender,
			IsMomoLinked:  accepter.IsMomoLinked,
			BalanceInApp:  accepter.BalanceInApp,
			AverageRating: accepter.AverageRating,
		},
		RideRequestID: req.RideRequestID,
		Waypoints:     waypointDetails,
	}

	// Prepare the WebSocket message
	wsMessage := schemas.WebSocketMessage{
		UserID:  req.ReceiverID.String(),
		Type:    "accept-give-ride-request",
		Payload: wsRes,
	}

	// Convert res to map[string]string
	resMap, err := helper.ConvertToStringMap(wsRes)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	notificationPayload := schemas.NotificationPayload{
		Type: "accept-give-ride-request",
		Data: resMap,
	}

	// Convert notificationPayload to map[string]string
	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the notification message
	notification := schemas.Notification{
		Title: "Yêu cầu đi nhờ của bạn đã được chấp nhận",
		Body:  "Chuyến đi của bạn đã được chấp nhận, hãy chuẩn bị sẵn sàng để bắt đầu chuyến đi",
		Token: receiver.DeviceToken,
		Data:  notificationPayloadMap,
	}

	// Send the WebSocket message using the async client
	go func() {
		err := ctrl.asyncClient.EnqueueWebsocketMessage(wsMessage)
		if err != nil {
			log.Printf("Failed to enqueue websocket message: %v", err)
		}
	}()

	// Send the notification message using the async client
	go func() {
		err = ctrl.asyncClient.EnqueueFCMNotification(notification)
		if err != nil {
			log.Printf("Failed to enqueue FCM notification: %v", err)
		}
	}()

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

	// Get waypoints details from ride_offer_id
	waypoints, err := ctrl.MapsService.GetAllWaypoints(rideOffer.ID)
	if err != nil {
		helper.GinResponse(ctx, 500, helper.ErrorResponseWithMessage(
			err,
			"Failed to get waypoints",
			"Không thể lấy thông tin waypoints",
		))
		return
	}

	var waypointDetails []schemas.Waypoint
	if waypoints != nil {
		waypointDetails = make([]schemas.Waypoint, 0, len(waypoints))
		for _, waypoint := range waypoints {
			waypointDetails = append(waypointDetails, schemas.Waypoint{
				Latitude:  waypoint.Latitude,
				Longitude: waypoint.Longitude,
				Address:   waypoint.Address,
				ID:        waypoint.ID,
				Order:     waypoint.WaypointOrder,
			})
		}
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

	// Create a transaction to store fare details
	transaction, err := ctrl.RideService.CreateRideTransaction(ride.ID, ride.Fare, "cash", data.UserID, req.ReceiverID)
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

	// Get receiver device token to send notification
	receiver, err := ctrl.UserService.GetUserByID(req.ReceiverID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get receiver details",
			"Không thể lấy thông tin người nhận",
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
		Status:                 ride.Status,
		StartTime:              ride.StartTime,
		RideOfferID:            ride.RideOfferID,
		DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
		DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
		RiderCurrentLatitude:   rideRequest.RiderCurrentLatitude,
		RiderCurrentLongitude:  rideRequest.RiderCurrentLongitude,
		EndTime:                ride.EndTime,
		StartAddress:           ride.StartAddress,
		EndAddress:             ride.EndAddress,
		Fare:                   ride.Fare,
		EncodedPolyline:        string(ride.EncodedPolyline),
		Distance:               ride.Distance,
		Duration:               ride.Duration,
		StartLatitude:          ride.StartLatitude,
		StartLongitude:         ride.StartLongitude,
		EndLatitude:            ride.EndLatitude,
		EndLongitude:           ride.EndLongitude,
		ReceiverID:             req.ReceiverID,
		UserInfo: schemas.UserInfo{
			ID:            receiver.ID,
			PhoneNumber:   receiver.PhoneNumber,
			FullName:      receiver.FullName,
			AvatarURL:     receiver.AvatarURL,
			Gender:        receiver.Gender,
			IsMomoLinked:  receiver.IsMomoLinked,
			BalanceInApp:  receiver.BalanceInApp,
			AverageRating: receiver.AverageRating,
		},
		Vehicle:   vehicle,
		Waypoints: waypointDetails,
	}

	// Send the accepted ride request to the hitcher (match the ride successfully)
	// ctrl.hub.SendToUser(req.ReceiverID.String(), "accept-hitch-ride-request", res)

	// Get accepter user details from user_id
	accepter, err := ctrl.UserService.GetUserByID(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get accepter details",
			"Không thể lấy thông tin người chấp nhận",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	wsRes := schemas.AcceptHitchRideRequestResponse{
		ID:            ride.ID,
		RideRequestID: ride.RideRequestID,
		Transaction: schemas.TransactionDetail{
			ID:            transaction.ID,
			Amount:        transaction.Amount,
			Status:        transaction.Status,
			PaymentMethod: transaction.PaymentMethod,
		},
		Status:                 ride.Status,
		StartTime:              ride.StartTime,
		RideOfferID:            ride.RideOfferID,
		DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
		DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
		RiderCurrentLatitude:   rideRequest.RiderCurrentLatitude,
		RiderCurrentLongitude:  rideRequest.RiderCurrentLongitude,
		EndTime:                ride.EndTime,
		StartAddress:           ride.StartAddress,
		EndAddress:             ride.EndAddress,
		Fare:                   ride.Fare,
		EncodedPolyline:        string(ride.EncodedPolyline),
		Distance:               ride.Distance,
		Duration:               ride.Duration,
		StartLatitude:          ride.StartLatitude,
		StartLongitude:         ride.StartLongitude,
		EndLatitude:            ride.EndLatitude,
		EndLongitude:           ride.EndLongitude,
		ReceiverID:             req.ReceiverID,
		UserInfo: schemas.UserInfo{
			ID:            accepter.ID,
			PhoneNumber:   accepter.PhoneNumber,
			FullName:      accepter.FullName,
			AvatarURL:     accepter.AvatarURL,
			Gender:        accepter.Gender,
			IsMomoLinked:  accepter.IsMomoLinked,
			BalanceInApp:  accepter.BalanceInApp,
			AverageRating: accepter.AverageRating,
		},
		Vehicle:   vehicle,
		Waypoints: waypointDetails,
	}

	// Prepare the WebSocket message
	wsMessage := schemas.WebSocketMessage{
		UserID:  req.ReceiverID.String(),
		Type:    "accept-hitch-ride-request",
		Payload: wsRes,
	}

	// Convert res to map[string]string
	resMap, err := helper.ConvertToStringMap(wsRes)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	notificationPayload := schemas.NotificationPayload{
		Type: "accept-hitch-ride-request",
		Data: resMap,
	}

	// Convert notificationPayload to map[string]string
	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the notification message
	notification := schemas.Notification{
		Title: "Lời mời đi nhờ của bạn đã được chấp nhận",
		Body:  "Chuyến đi của bạn đã được chấp nhận, hãy chuẩn bị sẵn sàng để bắt đầu chuyến đi",
		Token: receiver.DeviceToken,
		Data:  notificationPayloadMap,
	}

	// Send the WebSocket message using the async client
	go func() {
		err := ctrl.asyncClient.EnqueueWebsocketMessage(wsMessage)
		if err != nil {
			log.Printf("Failed to enqueue websocket message: %v", err)
		}
	}()

	// Send the notification message using the async client
	go func() {
		err = ctrl.asyncClient.EnqueueFCMNotification(notification)
		if err != nil {
			log.Printf("Failed to enqueue FCM notification: %v", err)
		}
	}()

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
		ReceiverID:    req.ReceiverID,
	}

	// Send the cancel notification to the driver
	// ctrl.hub.SendToUser(req.ReceiverID.String(), "cancel-give-ride-request", res)

	// Get receiver device token to send notification
	receiver, err := ctrl.UserService.GetUserByID(req.ReceiverID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get receiver details",
			"Không thể lấy thông tin người nhận",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the WebSocket message
	wsMessage := schemas.WebSocketMessage{
		UserID:  req.ReceiverID.String(),
		Type:    "cancel-give-ride-request",
		Payload: res,
	}

	// Convert res to map[string]string
	resMap, err := helper.ConvertToStringMap(res)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	notificationPayload := schemas.NotificationPayload{
		Type: "cancel-give-ride-request",
		Data: resMap,
	}

	// Convert notificationPayload to map[string]string
	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the notification message
	notification := schemas.Notification{
		Title: "Lời mời đi nhờ của bạn đã bị hủy",
		Body:  "Lời mời đi nhờ của bạn đã bị hủy, vui lòng thử lại sau",
		Token: receiver.DeviceToken,
		Data:  notificationPayloadMap,
	}

	// Send the WebSocket message using the async client
	go func() {
		err := ctrl.asyncClient.EnqueueWebsocketMessage(wsMessage)
		if err != nil {
			log.Printf("Failed to enqueue websocket message: %v", err)
		}
	}()

	// Send the notification message using the async client
	go func() {
		err = ctrl.asyncClient.EnqueueFCMNotification(notification)
		if err != nil {
			log.Printf("Failed to enqueue FCM notification: %v", err)
		}
	}()

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
		ReceiverID:    req.ReceiverID,
	}
	// Send the cancel notification to the hitcher
	// ctrl.hub.SendToUser(req.ReceiverID.String(), "cancel-hitch-ride-request", res)

	// Get receiver device token to send notification
	receiver, err := ctrl.UserService.GetUserByID(req.ReceiverID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get receiver details",
			"Không thể lấy thông tin người nhận",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the WebSocket message
	wsMessage := schemas.WebSocketMessage{
		UserID:  req.ReceiverID.String(),
		Type:    "cancel-hitch-ride-request",
		Payload: res,
	}

	// Convert res to map[string]string
	resMap, err := helper.ConvertToStringMap(res)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	notificationPayload := schemas.NotificationPayload{
		Type: "cancel-hitch-ride-request",
		Data: resMap,
	}

	// Convert notificationPayload to map[string]string
	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the notification message
	notification := schemas.Notification{
		Title: "Yêu cầu đi nhờ của bạn đã bị hủy",
		Body:  "Yêu cầu đi nhờ của bạn đã bị hủy, vui lòng thử lại sau",
		Token: receiver.DeviceToken,
		Data:  notificationPayloadMap,
	}

	// Send the WebSocket message using the async client
	go func() {
		err := ctrl.asyncClient.EnqueueWebsocketMessage(wsMessage)
		if err != nil {
			log.Printf("Failed to enqueue websocket message: %v", err)
		}
	}()

	// Send the notification message using the async client
	go func() {
		err = ctrl.asyncClient.EnqueueFCMNotification(notification)
		if err != nil {
			log.Printf("Failed to enqueue FCM notification: %v", err)
		}
	}()

	// Return success response
	helper.GinResponse(ctx, 200, helper.SuccessResponse(
		nil,
		"Successfully canceled ride request",
		"Đã hủy yêu cầu chia sẻ chuyến đi thành công",
	))
}

// StartRide starts the ride between the driver and the hitcher (the driver must starts the ride)
// StartRide starts the ride between the driver and the hitcher (the driver must starts the ride)
// @Summary Start a ride
// @Description Starts the ride between the driver and the hitcher
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.StartRideRequest true "Start ride request"
// @Success 200 {object} helper.Response{data=schemas.StartRideResponse} "Successfully started ride"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/start-ride [post]
func (ctrl *RideController) StartRide(ctx *gin.Context) {
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

	var req schemas.StartRideRequest
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

	// Start the ride
	ride, err := ctrl.RideService.StartRide(req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to start ride",
			"Không thể bắt đầu chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get ride offer details from ride_offer_id
	rideOffer, err := ctrl.RideService.GetRideOfferByID(ride.RideOfferID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride offer details",
			"Không thể lấy thông tin chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get waypoints details from ride_offer_id
	waypoints, err := ctrl.MapsService.GetAllWaypoints(rideOffer.ID)
	if err != nil {
		helper.GinResponse(ctx, 500, helper.ErrorResponseWithMessage(
			err,
			"Failed to get waypoints",
			"Không thể lấy thông tin waypoints",
		))
		return
	}

	var waypointDetails []schemas.Waypoint
	if waypoints != nil {
		waypointDetails = make([]schemas.Waypoint, 0, len(waypoints))
		for _, waypoint := range waypoints {
			waypointDetails = append(waypointDetails, schemas.Waypoint{
				Latitude:  waypoint.Latitude,
				Longitude: waypoint.Longitude,
				Address:   waypoint.Address,
				ID:        waypoint.ID,
				Order:     waypoint.WaypointOrder,
			})
		}
	}

	// Get ride request details from ride_request_id
	rideRequest, err := ctrl.RideService.GetRideRequestByID(ride.RideRequestID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride request details",
			"Không thể lấy thông tin yêu cầu chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get vehicle details from vehicle_id
	vehicle, err := ctrl.VehicleService.GetVehicleFromID(ride.VehicleID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get vehicle details",
			"Không thể lấy thông tin phương tiện",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get driver details from user_id
	driver, err := ctrl.UserService.GetUserByID(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get driver details",
			"Không thể lấy thông tin tài xế",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get transaction details from transaction_id
	transaction, err := ctrl.RideService.GetTransactionByRideID(ride.ID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get transaction details",
			"Không thể lấy thông tin giao dịch",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.StartRideResponse{
		ID:            ride.ID,
		RideOfferID:   ride.RideOfferID,
		RideRequestID: ride.RideRequestID,
		Transaction: schemas.TransactionDetail{
			ID:            transaction.ID,
			Amount:        transaction.Amount,
			Status:        transaction.Status,
			PaymentMethod: transaction.PaymentMethod,
		},
		User: schemas.UserInfo{
			ID:            driver.ID,
			FullName:      driver.FullName,
			PhoneNumber:   driver.PhoneNumber,
			AvatarURL:     driver.AvatarURL,
			Gender:        driver.Gender,
			IsMomoLinked:  driver.IsMomoLinked,
			BalanceInApp:  driver.BalanceInApp,
			AverageRating: driver.AverageRating,
		},
		Status:                 ride.Status,
		StartTime:              ride.StartTime,
		DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
		DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
		RiderCurrentLatitude:   rideRequest.RiderCurrentLatitude,
		RiderCurrentLongitude:  rideRequest.RiderCurrentLongitude,
		EndTime:                ride.EndTime,
		StartAddress:           ride.StartAddress,
		EndAddress:             ride.EndAddress,
		Fare:                   ride.Fare,
		EncodedPolyline:        string(ride.EncodedPolyline),
		Distance:               ride.Distance,
		Duration:               ride.Duration,
		StartLatitude:          ride.StartLatitude,
		StartLongitude:         ride.StartLongitude,
		EndLatitude:            ride.EndLatitude,
		EndLongitude:           ride.EndLongitude,
		Vehicle:                vehicle,
		ReceiverID:             rideRequest.UserID, // ReceiverID is the hitcher's user_id
		Waypoints:              waypointDetails,
	}

	// Get receiver device token to send notification
	receiver, err := ctrl.UserService.GetUserByID(rideRequest.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get receiver details",
			"Không thể lấy thông tin người nhận",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the WebSocket message
	wsMessage := schemas.WebSocketMessage{
		UserID:  rideRequest.UserID.String(),
		Type:    "start-ride",
		Payload: res,
	}

	// Convert res to map[string]string
	resMap, err := helper.ConvertToStringMap(res)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	notificationPayload := schemas.NotificationPayload{
		Type: "start-ride",
		Data: resMap,
	}

	// Convert notificationPayload to map[string]string
	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the notification message
	notification := schemas.Notification{
		Title: "Chuyến đi của bạn đã bắt đầu",
		Body:  "Chuyến đi của bạn đã bắt đầu, hãy chuẩn bị sẵn sàng để bắt đầu chuyến đi",
		Token: receiver.DeviceToken,
		Data:  notificationPayloadMap,
	}

	// Send the WebSocket message using the async client
	go func() {
		err := ctrl.asyncClient.EnqueueWebsocketMessage(wsMessage)
		if err != nil {
			log.Printf("Failed to enqueue websocket message: %v", err)
		}
	}()

	// Send the notification message using the async client
	go func() {
		err = ctrl.asyncClient.EnqueueFCMNotification(notification)
		if err != nil {
			log.Printf("Failed to enqueue FCM notification: %v", err)
		}
	}()

	// Return success response
	response := helper.SuccessResponse(
		res,
		"Successfully started ride",
		"Đã bắt đầu chuyến đi thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// EndRide ends the ride between the driver and the hitcher (the driver must ends the ride)
// EndRide godoc
// @Summary End a ride
// @Description Ends the ride between the driver and the hitcher
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.EndRideRequest true "End ride request"
// @Success 200 {object} helper.Response{data=schemas.EndRideResponse} "Successfully ended ride"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/end-ride [post]
func (ctrl *RideController) EndRide(ctx *gin.Context) {
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

	var req schemas.EndRideRequest
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

	// End the ride
	ride, err := ctrl.RideService.EndRide(req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to end ride",
			"Không thể kết thúc chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get ride offer details from ride_offer_id
	rideOffer, err := ctrl.RideService.GetRideOfferByID(ride.RideOfferID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride offer details",
			"Không thể lấy thông tin chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get waypoints details from ride_id
	waypoints, err := ctrl.MapsService.GetAllWaypoints(rideOffer.ID)
	if err != nil {
		helper.GinResponse(ctx, 500, helper.ErrorResponseWithMessage(
			err,
			"Failed to get waypoints",
			"Không thể lấy thông tin waypoints",
		))
		return
	}

	var waypointDetails []schemas.Waypoint
	if waypoints != nil {
		waypointDetails = make([]schemas.Waypoint, 0, len(waypoints))
		for _, waypoint := range waypoints {
			waypointDetails = append(waypointDetails, schemas.Waypoint{
				Latitude:  waypoint.Latitude,
				Longitude: waypoint.Longitude,
				Address:   waypoint.Address,
				ID:        waypoint.ID,
				Order:     waypoint.WaypointOrder,
			})
		}
	}

	// Get ride request details from ride_request_id
	rideRequest, err := ctrl.RideService.GetRideRequestByID(ride.RideRequestID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride request details",
			"Không thể lấy thông tin yêu cầu chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get vehicle details from vehicle_id
	vehicle, err := ctrl.VehicleService.GetVehicleFromID(ride.VehicleID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get vehicle details",
			"Không thể lấy thông tin phương tiện",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get driver details from user_id
	driver, err := ctrl.UserService.GetUserByID(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get driver details",
			"Không thể lấy thông tin tài xế",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get transaction details from transaction_id
	transaction, err := ctrl.RideService.GetTransactionByRideID(ride.ID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get transaction details",
			"Không thể lấy thông tin giao dịch",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get receiver device token to send notification
	receiver, err := ctrl.UserService.GetUserByID(rideRequest.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get receiver details",
			"Không thể lấy thông tin người nhận",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.EndRideResponse{
		ID:            ride.ID,
		RideOfferID:   ride.RideOfferID,
		RideRequestID: ride.RideRequestID,
		Transaction: schemas.TransactionDetail{
			ID:            transaction.ID,
			Amount:        transaction.Amount,
			Status:        transaction.Status,
			PaymentMethod: transaction.PaymentMethod,
		},
		User: schemas.UserInfo{
			ID:            receiver.ID,
			FullName:      receiver.FullName,
			PhoneNumber:   receiver.PhoneNumber,
			AvatarURL:     receiver.AvatarURL,
			Gender:        receiver.Gender,
			IsMomoLinked:  receiver.IsMomoLinked,
			BalanceInApp:  receiver.BalanceInApp,
			AverageRating: receiver.AverageRating,
		},
		Status:                 ride.Status,
		StartTime:              ride.StartTime,
		DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
		DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
		RiderCurrentLatitude:   rideRequest.RiderCurrentLatitude,
		RiderCurrentLongitude:  rideRequest.RiderCurrentLongitude,
		EndTime:                ride.EndTime,
		StartAddress:           ride.StartAddress,
		EndAddress:             ride.EndAddress,
		Fare:                   ride.Fare,
		EncodedPolyline:        string(ride.EncodedPolyline),
		Distance:               ride.Distance,
		Duration:               ride.Duration,
		StartLatitude:          ride.StartLatitude,
		StartLongitude:         ride.StartLongitude,
		EndLatitude:            ride.EndLatitude,
		EndLongitude:           ride.EndLongitude,
		Vehicle:                vehicle,
		ReceiverID:             rideRequest.UserID, // ReceiverID is the hitcher's user_id
		Waypoints:              waypointDetails,
	}

	wsRes := schemas.EndRideResponse{
		ID:            ride.ID,
		RideOfferID:   ride.RideOfferID,
		RideRequestID: ride.RideRequestID,
		Transaction: schemas.TransactionDetail{
			ID:            transaction.ID,
			Amount:        transaction.Amount,
			Status:        transaction.Status,
			PaymentMethod: transaction.PaymentMethod,
		},
		User: schemas.UserInfo{
			ID:            driver.ID,
			FullName:      driver.FullName,
			PhoneNumber:   driver.PhoneNumber,
			AvatarURL:     driver.AvatarURL,
			Gender:        driver.Gender,
			IsMomoLinked:  driver.IsMomoLinked,
			BalanceInApp:  driver.BalanceInApp,
			AverageRating: driver.AverageRating,
		},
		Status:                 ride.Status,
		StartTime:              ride.StartTime,
		DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
		DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
		RiderCurrentLatitude:   rideRequest.RiderCurrentLatitude,
		RiderCurrentLongitude:  rideRequest.RiderCurrentLongitude,
		EndTime:                ride.EndTime,
		StartAddress:           ride.StartAddress,
		EndAddress:             ride.EndAddress,
		Fare:                   ride.Fare,
		EncodedPolyline:        string(ride.EncodedPolyline),
		Distance:               ride.Distance,
		Duration:               ride.Duration,
		StartLatitude:          ride.StartLatitude,
		StartLongitude:         ride.StartLongitude,
		EndLatitude:            ride.EndLatitude,
		EndLongitude:           ride.EndLongitude,
		Vehicle:                vehicle,
		ReceiverID:             rideRequest.UserID, // ReceiverID is the hitcher's user_id
		Waypoints:              waypointDetails,
	}

	// Prepare the WebSocket message
	wsMessage := schemas.WebSocketMessage{
		UserID:  rideRequest.UserID.String(),
		Type:    "end-ride",
		Payload: wsRes,
	}

	// Convert res to map[string]string
	resMap, err := helper.ConvertToStringMap(wsRes)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	notificationPayload := schemas.NotificationPayload{
		Type: "end-ride",
		Data: resMap,
	}

	// Convert notificationPayload to map[string]string
	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the notification message
	notification := schemas.Notification{
		Title: "Chuyến đi của bạn đã kết thúc",
		Body:  "Chuyến đi của bạn đã kết thúc, cảm ơn bạn đã sử dụng dịch vụ của chúng tôi",
		Token: receiver.DeviceToken,
		Data:  notificationPayloadMap,
	}

	// Send the WebSocket message using the async client
	go func() {
		err := ctrl.asyncClient.EnqueueWebsocketMessage(wsMessage)
		if err != nil {
			log.Printf("Failed to enqueue websocket message: %v", err)
		}
	}()

	// Send the notification message using the async client
	go func() {
		err = ctrl.asyncClient.EnqueueFCMNotification(notification)
		if err != nil {
			log.Printf("Failed to enqueue FCM notification: %v", err)
		}
	}()

	// Return success response
	response := helper.SuccessResponse(
		res,
		"Successfully ended ride",
		"Đã kết thúc chuyến đi thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// UpdateRideLocation updates the current location of the driver during the ride (the driver must update the location)
// UpdateRideLocation godoc
// @Summary Update the current location of the driver during the ride
// @Description Updates the current location of the driver during the ride
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.UpdateRideLocationRequest true "Update ride location request"
// @Success 200 {object} helper.Response{data=schemas.UpdateRideLocationResponse} "Successfully updated ride location"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/update-ride-location [post]
func (ctrl *RideController) UpdateRideLocation(ctx *gin.Context) {
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

	var req schemas.UpdateRideLocationRequest
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

	// Update the ride location
	ride, err := ctrl.RideService.UpdateRideLocation(req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to update ride location",
			"Không thể cập nhật vị trí chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get ride offer details from ride_offer_id
	rideOffer, err := ctrl.RideService.GetRideOfferByID(ride.RideOfferID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride offer details",
			"Không thể lấy thông tin chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	waypoints, err := ctrl.MapsService.GetAllWaypoints(rideOffer.ID)
	if err != nil {
		helper.GinResponse(ctx, 500, helper.ErrorResponseWithMessage(
			err,
			"Failed to get waypoints",
			"Không thể lấy thông tin waypoints",
		))
		return
	}
	var waypointDetails []schemas.Waypoint
	if waypoints != nil {
		waypointDetails = make([]schemas.Waypoint, 0, len(waypoints))
		for _, waypoint := range waypoints {
			waypointDetails = append(waypointDetails, schemas.Waypoint{
				Latitude:  waypoint.Latitude,
				Longitude: waypoint.Longitude,
				Address:   waypoint.Address,
				ID:        waypoint.ID,
				Order:     waypoint.WaypointOrder,
			})
		}
	}

	// Get ride request details from ride_request_id
	rideRequest, err := ctrl.RideService.GetRideRequestByID(ride.RideRequestID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride request details",
			"Không thể lấy thông tin yêu cầu chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get vehicle details from vehicle_id
	vehicle, err := ctrl.VehicleService.GetVehicleFromID(ride.VehicleID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get vehicle details",
			"Không thể lấy thông tin phương tiện",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get driver details from user_id
	driver, err := ctrl.UserService.GetUserByID(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get driver details",
			"Không thể lấy thông tin tài xế",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get transaction details from transaction_id
	transaction, err := ctrl.RideService.GetTransactionByRideID(ride.ID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get transaction details",
			"Không thể lấy thông tin giao dịch",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.UpdateRideLocationResponse{
		ID:            ride.ID,
		RideOfferID:   ride.RideOfferID,
		RideRequestID: ride.RideRequestID,
		Transaction: schemas.TransactionDetail{
			ID:            transaction.ID,
			Amount:        transaction.Amount,
			Status:        transaction.Status,
			PaymentMethod: transaction.PaymentMethod,
		},
		User: schemas.UserInfo{
			ID:            driver.ID,
			FullName:      driver.FullName,
			PhoneNumber:   driver.PhoneNumber,
			AvatarURL:     driver.AvatarURL,
			Gender:        driver.Gender,
			IsMomoLinked:  driver.IsMomoLinked,
			BalanceInApp:  driver.BalanceInApp,
			AverageRating: driver.AverageRating,
		},
		Status:                 ride.Status,
		StartTime:              ride.StartTime,
		DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
		DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
		RiderCurrentLatitude:   rideRequest.RiderCurrentLatitude,
		RiderCurrentLongitude:  rideRequest.RiderCurrentLongitude,
		EndTime:                ride.EndTime,
		StartAddress:           ride.StartAddress,
		EndAddress:             ride.EndAddress,
		Fare:                   ride.Fare,
		EncodedPolyline:        string(ride.EncodedPolyline),
		Distance:               ride.Distance,
		Duration:               ride.Duration,
		StartLatitude:          ride.StartLatitude,
		StartLongitude:         ride.StartLongitude,
		EndLatitude:            ride.EndLatitude,
		EndLongitude:           ride.EndLongitude,
		Vehicle:                vehicle,
		ReceiverID:             rideRequest.UserID, // ReceiverID is the hitcher's user_id
		Waypoints:              waypointDetails,
	}

	// // Get receiver device token to send notification
	// receiver, err := ctrl.UserService.GetUserByID(rideRequest.UserID)
	// if err != nil {
	// 	response := helper.ErrorResponseWithMessage(
	// 		err,
	// 		"Failed to get receiver details",
	// 		"Không thể lấy thông tin người nhận",
	// 	)
	// 	helper.GinResponse(ctx, 500, response)
	// 	return
	// }

	// Prepare the WebSocket message
	wsMessage := schemas.WebSocketMessage{
		UserID:  rideRequest.UserID.String(),
		Type:    "update-ride-location",
		Payload: res,
	}

	// // Convert res to map[string]string
	// resMap, err := helper.ConvertToStringMap(res)
	// if err != nil {
	// 	response := helper.ErrorResponseWithMessage(
	// 		err,
	// 		"Failed to convert struct to map",
	// 		"Không thể chuyển đổi struct sang map",
	// 	)
	// 	helper.GinResponse(ctx, 500, response)
	// 	return
	// }

	// notificationPayload := schemas.NotificationPayload{
	// 	Type: "update-ride-location",
	// 	Data: resMap,
	// }

	// // Convert notificationPayload to map[string]string
	// notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	// if err != nil {
	// 	response := helper.ErrorResponseWithMessage(
	// 		err,
	// 		"Failed to convert struct to map",
	// 		"Không thể chuyển đổi struct sang map",
	// 	)
	// 	helper.GinResponse(ctx, 500, response)
	// 	return
	// }

	// // Prepare the notification message
	// notification := schemas.Notification{
	// 	Title: "Vị trí của tài xế đã được cập nhật",
	// 	Body:  "Vị trí của tài xế đã được cập nhật, vui lòng kiểm tra vị trí của tài xế",
	// 	Token: receiver.DeviceToken,
	// 	Data:  notificationPayloadMap,
	// }

	// Send the WebSocket message using the async client
	go func() {
		err := ctrl.asyncClient.EnqueueWebsocketMessage(wsMessage)
		if err != nil {
			log.Printf("Failed to enqueue websocket message: %v", err)
		}
	}()

	// // Send the notification message using the async client
	// go func() {
	// 	err = ctrl.asyncClient.EnqueueFCMNotification(notification)
	// 	if err != nil {
	// 		log.Printf("Failed to enqueue FCM notification: %v", err)
	// 	}
	// }()

	// Return success response
	response := helper.SuccessResponse(
		res,
		"Successfully updated ride location",
		"Đã cập nhật vị trí chuyến đi thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// CancelRideByDriver cancels the ride
// CancelRideByDriver godoc
// @Summary Cancel a ride
// @Description Cancels the ride
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.CancelRideRequest true "Cancel ride request"
// @Success 200 {object} helper.Response{data=schemas.CancelRideResponse} "Successfully canceled ride by driver"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/cancel-ride [post]
func (ctrl *RideController) CancelRide(ctx *gin.Context) {
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

	var req schemas.CancelRideRequest
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

	// Check if payment method is momo to refund the hitcher
	transaction, err := ctrl.RideService.GetTransactionByRideID(req.RideID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get transaction details",
			"Không thể lấy thông tin giao dịch",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get the device token of the hitcher to send notification
	receiver, err := ctrl.UserService.GetUserByID(req.ReceiverID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get receiver details",
			"Không thể lấy thông tin người nhận",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get the ride details
	rideDetail, err := ctrl.RideService.GetRideByID(req.RideID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride details",
			"Không thể lấy thông tin chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Check if the payment method is momo to refund the hitcher
	if transaction.PaymentMethod == "momo" {
		err := ctrl.PaymentService.RefundRide(
			req.ReceiverID, schemas.RefundMomoRequest{
				RideRequestID: rideDetail.RideRequestID,
				RideOfferID:   rideDetail.RideOfferID,
			},
		)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to refund the hitcher",
				"Không thể hoàn tiền cho người nhận",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}

		// Send notification and WebSocket message to the hitcher
		notification := schemas.Notification{
			Title: "Chuyến đi của bạn đã bị hủy",
			Body:  "Chuyến đi của bạn đã bị hủy, bạn đã được hoàn tiền",
			Token: receiver.DeviceToken,
			Data:  nil,
		}

		wsMessage := schemas.WebSocketMessage{
			UserID:  req.ReceiverID.String(),
			Type:    "refund-success",
			Payload: nil,
		}

		// Send the WebSocket message using the async client
		go func() {
			err := ctrl.asyncClient.EnqueueWebsocketMessage(wsMessage)
			if err != nil {
				log.Printf("Failed to enqueue websocket message: %v", err)
			}
		}()

		// Send the notification message using the async client
		go func() {
			err = ctrl.asyncClient.EnqueueFCMNotification(notification)
			if err != nil {
				log.Printf("Failed to enqueue FCM notification: %v", err)
			}
		}()
	}

	// Cancel the ride by the driver
	ride, err := ctrl.RideService.CancelRide(req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to cancel ride by driver",
			"Không thể hủy chuyến đi bởi tài xế",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.CancelRideResponse{
		RideID:        ride.ID,
		RideOfferID:   ride.RideOfferID,
		RideRequestID: ride.RideRequestID,
		ReceiverID:    req.ReceiverID,
	}

	// // Get ride offer details from ride_offer_id
	// rideOffer, err := ctrl.RideService.GetRideOfferByID(ride.RideOfferID)
	// if err != nil {
	// 	response := helper.ErrorResponseWithMessage(
	// 		err,
	// 		"Failed to get ride offer details",
	// 		"Không thể lấy thông tin chuyến đi",
	// 	)
	// 	helper.GinResponse(ctx, 500, response)
	// 	return
	// }

	// // Get ride request details from ride_request_id
	// rideRequest, err := ctrl.RideService.GetRideRequestByID(ride.RideRequestID)
	// if err != nil {
	// 	response := helper.ErrorResponseWithMessage(
	// 		err,
	// 		"Failed to get ride request details",
	// 		"Không thể lấy thông tin yêu cầu chuyến đi",
	// 	)
	// 	helper.GinResponse(ctx, 500, response)
	// 	return
	// }

	// Prepare the WebSocket message
	wsMessage := schemas.WebSocketMessage{
		UserID:  req.ReceiverID.String(),
		Type:    "cancel-ride-by-driver",
		Payload: res,
	}

	// Convert ride to map[string]string
	resMap, err := helper.ConvertToStringMap(res)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the notification payload
	notificationPayload := schemas.NotificationPayload{
		Type: "cancel-ride-by-driver",
		Data: resMap,
	}

	// Convert notificationPayload to map[string]string
	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert struct to map",
			"Không thể chuyển đổi struct sang map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare the notification message
	notification := schemas.Notification{
		Title: "Chuyến đi của bạn đã bị hủy",
		Body:  "Chuyến đi của bạn đã bị hủy, vui lòng thử lại sau",
		Token: receiver.DeviceToken,
		Data:  notificationPayloadMap,
	}

	// Send the WebSocket message using the async client
	go func() {
		err := ctrl.asyncClient.EnqueueWebsocketMessage(wsMessage)
		if err != nil {
			log.Printf("Failed to enqueue websocket message: %v", err)
		}
	}()

	// Send the notification message using the async client
	go func() {
		err = ctrl.asyncClient.EnqueueFCMNotification(notification)
		if err != nil {
			log.Printf("Failed to enqueue FCM notification: %v", err)
		}
	}()

	// Return success response
	response := helper.SuccessResponse(
		res,
		"Successfully canceled ride by driver",
		"Đã hủy chuyến đi bởi tài xế thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// // CancelRideByHitcher cancels the ride by the hitcher
// // CancelRideByHitcher godoc
// // @Summary Cancel a ride by the hitcher
// // @Description Cancels the ride by the hitcher
// // @Tags ride
// // @Accept json
// // @Produce json
// // @Security BearerAuth
// // @Param request body schemas.CancelRideByHitcherRequest true "Cancel ride by hitcher request"
// // @Success 200 {object} helper.Response{data=schemas.CancelRideByHitcherResponse} "Successfully canceled ride by hitcher"
// // @Failure 400 {object} helper.Response "Invalid request"
// // @Failure 500 {object} helper.Response "Internal server error"
// // @Router /ride/cancel-ride-by-hitcher [post]
// func (ctrl *RideController) CancelRideByHitcher(ctx *gin.Context) {
// 	// Get payload from context
// 	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

// 	// Convert payload to map
// 	data, err := helper.ConvertToPayload(payload)

// 	// If error occurs, return error response
// 	if err != nil {
// 		response := helper.ErrorResponseWithMessage(
// 			fmt.Errorf("failed to convert payload"),
// 			"Failed to convert payload",
// 			"Không thể chuyển đổi payload",
// 		)
// 		helper.GinResponse(ctx, 500, response)
// 		return
// 	}

// 	var req schemas.CancelRideByHitcherRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		response := helper.ErrorResponseWithMessage(
// 			err,
// 			"Failed to bind JSON",
// 			"Không thể bind JSON",
// 		)
// 		helper.GinResponse(ctx, 400, response)
// 		return
// 	}

// 	if err := ctrl.validate.Struct(req); err != nil {
// 		response := helper.ErrorResponseWithMessage(
// 			err,
// 			"Failed to validate request",
// 			"Không thể validate request",
// 		)
// 		helper.GinResponse(ctx, 400, response)
// 		return
// 	}

// 	// Cancel the ride by the hitcher
// 	ride, err := ctrl.RideService.CancelRideByHitcher(req, data.UserID)
// 	if err != nil {
// 		response := helper.ErrorResponseWithMessage(
// 			err,
// 			"Failed to cancel ride by hitcher",
// 			"Không thể hủy chuyến đi bởi người đi",
// 		)
// 		helper.GinResponse(ctx, 500, response)
// 		return
// 	}

// 	// // Get ride offer details from ride_offer_id
// 	// rideOffer, err := ctrl.RideService.GetRideOfferByID(ride.RideOfferID)
// 	// if err != nil {
// 	// 	response := helper.ErrorResponseWithMessage(
// 	// 		err,
// 	// 		"Failed to get ride offer details",
// 	// 		"Không thể lấy thông tin chuyến đi",
// 	// 	)
// 	// 	helper.GinResponse(ctx, 500, response)
// 	// 	return
// 	// }

// 	// // Get ride request details from ride_request_id
// 	// rideRequest, err := ctrl.RideService.GetRideRequestByID(ride.RideRequestID)
// 	// if err != nil {
// 	// 	response := helper.ErrorResponseWithMessage(
// 	// 		err,
// 	// 		"Failed to get ride request details",
// 	// 		"Không thể lấy thông tin yêu cầu chuyến đi",
// 	// 	)
// 	// 	helper.GinResponse(ctx, 500, response)
// 	// 	return
// 	// }

// 	res := schemas.CancelRideByHitcherResponse{
// 		RideID:        ride.ID,
// 		RideOfferID:   ride.RideOfferID,
// 		RideRequestID: ride.RideRequestID,
// 		ReceiverID:    req.ReceiverID,
// 	}

// 	// Get the device token of the driver to send notification
// 	receiver, err := ctrl.UserService.GetUserByID(req.ReceiverID)
// 	if err != nil {
// 		response := helper.ErrorResponseWithMessage(
// 			err,
// 			"Failed to get receiver details",
// 			"Không thể lấy thông tin người nhận",
// 		)
// 		helper.GinResponse(ctx, 500, response)
// 		return
// 	}

// 	// Prepare the WebSocket message
// 	wsMessage := schemas.WebSocketMessage{
// 		UserID:  req.ReceiverID.String(),
// 		Type:    "cancel-ride-by-hitcher",
// 		Payload: res,
// 	}

// 	// Convert ride to map[string]string
// 	resMap, err := helper.ConvertToStringMap(res)

// 	// Prepare the notification payload
// 	notificationPayload := schemas.NotificationPayload{
// 		Type: "cancel-ride-by-hitcher",
// 		Data: resMap,
// 	}

// 	// Convert notificationPayload to map[string]string
// 	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
// 	if err != nil {
// 		response := helper.ErrorResponseWithMessage(
// 			err,
// 			"Failed to convert struct to map",
// 			"Không thể chuyển đổi struct sang map",
// 		)
// 		helper.GinResponse(ctx, 500, response)
// 		return
// 	}

// 	// Prepare the notification message
// 	notification := schemas.Notification{
// 		Title: "Chuyến đi của bạn đã bị hủy",
// 		Body:  "Chuyến đi của bạn đã bị hủy, vui lòng thử lại sau",
// 		Token: receiver.DeviceToken,
// 		Data:  notificationPayloadMap,
// 	}

// 	// Send the WebSocket message using the async client
// 	go func() {
// 		err := ctrl.asyncClient.EnqueueWebsocketMessage(wsMessage)
// 		if err != nil {
// 			log.Printf("Failed to enqueue websocket message: %v", err)
// 		}
// 	}()

// 	// Send the notification message using the async client
// 	go func() {
// 		err = ctrl.asyncClient.EnqueueFCMNotification(notification)
// 		if err != nil {
// 			log.Printf("Failed to enqueue FCM notification: %v", err)
// 		}
// 	}()

// 	// Return success response
// 	response := helper.SuccessResponse(
// 		res,
// 		"Successfully canceled ride by hitcher",
// 		"Đã hủy chuyến đi bởi người đi thành công",
// 	)
// 	helper.GinResponse(ctx, 200, response)

// }

// GetAllPendingRide get all ride request and ride offer that are not cancelled of the user
// GetAllPendingRide godoc
// @Summary Get all pending ride
// @Description Get all ride request and ride offer that are not cancelled of the user
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.Response{data=schemas.GetAllPendingRideResponse} "Successfully get all pending ride"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/get-all-pending-ride [get]
func (ctrl *RideController) GetAllPendingRide(ctx *gin.Context) {
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

	// Get all ride request and ride offer that are not cancelled of the user
	pendingRideOffer, pendingRideRequest, err := ctrl.RideService.GetAllPendingRide(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get all pending ride",
			"Không thể lấy tất cả chuyến đi đang chờ",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	pendingRideOfferDetails := make([]schemas.RideOfferDetail, 0, len(pendingRideOffer))
	for _, rideOffer := range pendingRideOffer {
		driver, err := ctrl.UserService.GetUserByID(rideOffer.UserID)
		if err != nil {
			helper.GinResponse(ctx, 500, helper.ErrorResponseWithMessage(
				err,
				"Failed to get driver details",
				"Không thể lấy thông tin tài xế",
			))
			return
		}

		vehicle, err := ctrl.VehicleService.GetVehicleFromID(rideOffer.VehicleID)
		if err != nil {
			helper.GinResponse(ctx, 500, helper.ErrorResponseWithMessage(
				err,
				"Failed to get vehicle details",
				"Không thể lấy thông tin phương tiện",
			))
			return
		}

		waypoints, err := ctrl.MapsService.GetAllWaypoints(rideOffer.ID)
		if err != nil {
			helper.GinResponse(ctx, 500, helper.ErrorResponseWithMessage(
				err,
				"Failed to get waypoints",
				"Không thể lấy thông tin waypoints",
			))
			return
		}

		var waypointDetails []schemas.Waypoint
		if waypoints != nil {
			waypointDetails = make([]schemas.Waypoint, 0, len(waypoints))
			for _, waypoint := range waypoints {
				waypointDetails = append(waypointDetails, schemas.Waypoint{
					Latitude:  waypoint.Latitude,
					Longitude: waypoint.Longitude,
					Address:   waypoint.Address,
					ID:        waypoint.ID,
					Order:     waypoint.WaypointOrder,
				})
			}
		}

		pendingRideOfferDetails = append(pendingRideOfferDetails, schemas.RideOfferDetail{
			ID:      rideOffer.ID,
			Vehicle: vehicle,
			User: schemas.UserInfo{
				ID:            driver.ID,
				FullName:      driver.FullName,
				PhoneNumber:   driver.PhoneNumber,
				Gender:        driver.Gender,
				AvatarURL:     driver.AvatarURL,
				IsMomoLinked:  driver.IsMomoLinked,
				BalanceInApp:  driver.BalanceInApp,
				AverageRating: driver.AverageRating,
			},
			StartTime:              rideOffer.StartTime,
			StartLatitude:          rideOffer.StartLatitude,
			StartLongitude:         rideOffer.StartLongitude,
			EndLatitude:            rideOffer.EndLatitude,
			EndLongitude:           rideOffer.EndLongitude,
			StartAddress:           rideOffer.StartAddress,
			EndAddress:             rideOffer.EndAddress,
			EncodedPolyline:        string(rideOffer.EncodedPolyline),
			Distance:               rideOffer.Distance,
			Duration:               rideOffer.Duration,
			DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
			DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
			Status:                 rideOffer.Status,
			EndTime:                rideOffer.EndTime,
			Fare:                   rideOffer.Fare,
			Waypoints:              waypointDetails,
		})
	}

	pendingRideRequestDetails := make([]schemas.RideRequestDetail, 0, len(pendingRideRequest))
	for _, rideRequest := range pendingRideRequest {
		rider, err := ctrl.UserService.GetUserByID(rideRequest.UserID)
		if err != nil {
			helper.GinResponse(ctx, 500, helper.ErrorResponseWithMessage(
				err,
				"Failed to get rider details",
				"Không thể lấy thông tin người đi",
			))
			return
		}

		pendingRideRequestDetails = append(pendingRideRequestDetails, schemas.RideRequestDetail{
			ID: rideRequest.ID,
			User: schemas.UserInfo{
				ID:            rider.ID,
				FullName:      rider.FullName,
				PhoneNumber:   rider.PhoneNumber,
				AvatarURL:     rider.AvatarURL,
				Gender:        rider.Gender,
				IsMomoLinked:  rider.IsMomoLinked,
				BalanceInApp:  rider.BalanceInApp,
				AverageRating: rider.AverageRating,
			},
			StartTime:             rideRequest.StartTime,
			StartLatitude:         rideRequest.StartLatitude,
			StartLongitude:        rideRequest.StartLongitude,
			EndLatitude:           rideRequest.EndLatitude,
			EndLongitude:          rideRequest.EndLongitude,
			StartAddress:          rideRequest.StartAddress,
			EndAddress:            rideRequest.EndAddress,
			EncodedPolyline:       string(rideRequest.EncodedPolyline),
			Distance:              rideRequest.Distance,
			Duration:              rideRequest.Duration,
			RiderCurrentLatitude:  rideRequest.RiderCurrentLatitude,
			RiderCurrentLongitude: rideRequest.RiderCurrentLongitude,
			Status:                rideRequest.Status,
			EndTime:               rideRequest.EndTime,
		})
	}
	res := schemas.GetAllPendingRideResponse{
		PendingRideOffer:   pendingRideOfferDetails,
		PendingRideRequest: pendingRideRequestDetails,
	}

	// Return success response
	response := helper.SuccessResponse(
		res,
		"Successfully got all pending ride",
		"Đã lấy tất cả chuyến đi đang chờ thành công",
	)
	helper.GinResponse(ctx, 200, response)

}

// RatingRideHitcherRequest is the request to rate the hitcher after the ride by the driver
// RatingRideHitcherRequest godoc
// @Summary Rate the hitcher after the ride by the driver
// @Description Rate the hitcher after the ride by the driver
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.RatingRideHitcherRequest true "Rating ride hitcher request"
// @Success 200 {object} helper.Response "Successfully rated the hitcher"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/rating-ride-hitcher [post]
func (ctrl *RideController) RatingRideHitcher(ctx *gin.Context) {
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

	var req schemas.RatingRideHitcherRequest
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

	// Rate the hitcher after the ride by the driver
	err = ctrl.RideService.RatingRideHitcher(req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to rate the hitcher",
			"Không thể đánh giá người đi nhờ",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Return success response
	response := helper.SuccessResponse(
		nil,
		"Successfully rated the hitcher",
		"Đã đánh giá người đi nhờ thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// RatingRideDriverRequest is the request to rate the driver after the ride by the hitcher
// RatingRideDriverRequest godoc
// @Summary Rate the driver after the ride by the hitcher
// @Description Rate the driver after the ride by the hitcher
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.RatingRideDriverRequest true "Rating ride driver request"
// @Success 200 {object} helper.Response "Successfully rated the driver"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/rating-ride-driver [post]
func (ctrl *RideController) RatingRideDriver(ctx *gin.Context) {
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

	var req schemas.RatingRideDriverRequest
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

	// Rate the driver after the ride by the hitcher
	err = ctrl.RideService.RatingRideDriver(req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to rate the driver",
			"Không thể đánh giá tài xế",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Return success response
	response := helper.SuccessResponse(
		nil,
		"Successfully rated the driver",
		"Đã đánh giá tài xế thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// GetRideHistory gets the ride history of the user (both as driver and hitcher) included cancelled rides and completed rides
// GetRideHistory godoc
// @Summary Get ride history of the user
// @Description Get ride history of the user (both as driver and hitcher) included cancelled rides and completed rides
// @Tags ride
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.Response{data=schemas.GetRideHistoryResponse} "Successfully got ride history"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ride/get-ride-history [get]
func (ctrl *RideController) GetRideHistory(ctx *gin.Context) {
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

	// Get ride history of the user
	rideHistory, err := ctrl.RideService.GetRideHistory(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get ride history",
			"Không thể lấy lịch sử chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	rideHistoryDetails := make([]schemas.RideHistoryDetail, 0, len(rideHistory))
	for _, ride := range rideHistory {
		// Get the driver id from ride_offer_id
		rideOffer, err := ctrl.RideService.GetRideOfferByID(ride.RideOfferID)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to get driver id",
				"Không thể lấy id tài xế",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}

		// Get the driver details from the driver id
		driver, err := ctrl.UserService.GetUserByID(rideOffer.UserID)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to get driver details",
				"Không thể lấy thông tin tài xế",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}

		// Get the hitcher id from ride_request_id
		rideRequest, err := ctrl.RideService.GetRideRequestByID(ride.RideRequestID)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to get hitcher id",
				"Không thể lấy id người đi nhờ",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}

		// Get the hitcher details from the hitcher id
		hitcher, err := ctrl.UserService.GetUserByID(rideRequest.UserID)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to get hitcher details",
				"Không thể lấy thông tin người đi nhờ",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}

		// Get the vehicle details from the vehicle id
		vehicle, err := ctrl.VehicleService.GetVehicleFromID(rideOffer.VehicleID)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to get vehicle details",
				"Không thể lấy thông tin phương tiện",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}

		// Get the transaction details from the ride id
		transaction, err := ctrl.RideService.GetTransactionByRideID(ride.ID)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to get transaction details",
				"Không thể lấy thông tin giao dịch",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}

		// Get the waypoints of the ride
		waypoints, err := ctrl.MapsService.GetAllWaypoints(rideOffer.ID)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to get waypoints",
				"Không thể lấy thông tin waypoints",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}

		var waypointDetails []schemas.Waypoint
		if waypoints != nil {
			waypointDetails = make([]schemas.Waypoint, 0, len(waypoints))
			for _, waypoint := range waypoints {
				waypointDetails = append(waypointDetails, schemas.Waypoint{
					Latitude:  waypoint.Latitude,
					Longitude: waypoint.Longitude,
					Address:   waypoint.Address,
					ID:        waypoint.ID,
					Order:     waypoint.WaypointOrder,
				})
			}
		}

		rideHistoryDetails = append(rideHistoryDetails, schemas.RideHistoryDetail{
			Driver: schemas.UserInfo{
				ID:            driver.ID,
				FullName:      driver.FullName,
				PhoneNumber:   driver.PhoneNumber,
				AvatarURL:     driver.AvatarURL,
				Gender:        driver.Gender,
				IsMomoLinked:  driver.IsMomoLinked,
				BalanceInApp:  driver.BalanceInApp,
				AverageRating: driver.AverageRating,
			},
			Hitcher: schemas.UserInfo{
				ID:            hitcher.ID,
				FullName:      hitcher.FullName,
				PhoneNumber:   hitcher.PhoneNumber,
				AvatarURL:     hitcher.AvatarURL,
				Gender:        hitcher.Gender,
				IsMomoLinked:  hitcher.IsMomoLinked,
				BalanceInApp:  hitcher.BalanceInApp,
				AverageRating: hitcher.AverageRating,
			},
			ID:            ride.ID,
			RideOfferID:   ride.RideOfferID,
			RideRequestID: ride.RideRequestID,
			Transaction: schemas.TransactionDetail{
				ID:            transaction.ID,
				Amount:        transaction.Amount,
				Status:        transaction.Status,
				PaymentMethod: transaction.PaymentMethod,
			},
			Status:                 ride.Status,
			StartTime:              ride.StartTime,
			DriverCurrentLatitude:  rideOffer.DriverCurrentLatitude,
			DriverCurrentLongitude: rideOffer.DriverCurrentLongitude,
			RiderCurrentLatitude:   rideRequest.RiderCurrentLatitude,
			RiderCurrentLongitude:  rideRequest.RiderCurrentLongitude,
			EndTime:                ride.EndTime,
			StartAddress:           ride.StartAddress,
			EndAddress:             ride.EndAddress,
			Fare:                   ride.Fare,
			EncodedPolyline:        string(ride.EncodedPolyline),
			Distance:               ride.Distance,
			Duration:               ride.Duration,
			StartLatitude:          ride.StartLatitude,
			StartLongitude:         ride.StartLongitude,
			EndLatitude:            ride.EndLatitude,
			EndLongitude:           ride.EndLongitude,
			Vehicle:                vehicle,
			Waypoints:              waypointDetails,
		})
	}

	res := schemas.GetRideHistoryResponse{
		RideHistory: rideHistoryDetails,
	}

	// Return success response
	response := helper.SuccessResponse(
		res,
		"Successfully got ride history",
		"Đã lấy lịch sử chuyến đi thành công",
	)
	helper.GinResponse(ctx, 200, response)
}
