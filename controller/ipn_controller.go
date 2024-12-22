package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"shareway/helper"
	"shareway/infra/task"
	"shareway/infra/ws"
	"shareway/schemas"
	"shareway/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type IPNController struct {
	validate       *validator.Validate
	hub            *ws.Hub
	RideService    service.IRideService
	MapsService    service.IMapService
	UserService    service.IUsersService
	VehicleService service.IVehicleService
	PaymentService service.IPaymentService
	IPNService     service.IIPNService
	asyncClient    *task.AsyncClient
}

func NewIPNController(validate *validator.Validate, hub *ws.Hub, rideService service.IRideService, mapService service.IMapService, userService service.IUsersService, vehicleService service.IVehicleService, paymentService service.IPaymentService, ipnService service.IIPNService, asyncClient *task.AsyncClient) *IPNController {
	return &IPNController{
		validate:       validate,
		hub:            hub,
		RideService:    rideService,
		MapsService:    mapService,
		UserService:    userService,
		VehicleService: vehicleService,
		PaymentService: paymentService,
		IPNService:     ipnService,
		asyncClient:    asyncClient,
	}
}

// HandleIPN receives IPN from payment gateway and processes it
// HandleIPN godoc
// @Summary Handle IPN from payment gateway
// @Description Handle IPN from payment gateway
// @Tags ipn
// @Accept json
// @Produce json
// @Param body body schemas.MoMoIPN true "MoMo IPN"
// @Success 204 {string} string "No content"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /ipn/handle-ipn [post]
func (i *IPNController) HandleIPN(ctx *gin.Context) {
	logger := log.With().Str("handler", "HandleIPN").Logger()

	var req schemas.MoMoIPN
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error().Err(err).Msg("Failed to bind JSON")
		helper.GinResponse(ctx, http.StatusBadRequest, helper.ErrorResponseWithMessage(err, "Failed to get admin profile", "Không thể lấy thông tin admin"))
		return
	}

	logger.Info().Interface("ipn", req).Msg("Received IPN")

	if !i.IPNService.VerifyIPN(req) {
		logger.Error().Msg("Failed to verify IPN")
		helper.GinResponse(ctx, http.StatusUnauthorized, helper.ErrorResponseWithMessage(nil, "Failed to verify IPN", "Không thể xác minh IPN"))
		return
	}

	extraDataJSON, err := base64.StdEncoding.DecodeString(req.ExtraData)
	if err != nil {
		logger.Error().Err(err).Str("extraData", req.ExtraData).Msg("Failed to decode extra data")
		helper.GinResponse(ctx, http.StatusBadRequest, helper.ErrorResponseWithMessage(err, "Failed to decode extra data", "Không thể giải mã dữ liệu thêm"))
		return
	}

	var extraData schemas.ExtraData
	if err := json.Unmarshal(extraDataJSON, &extraData); err != nil {
		logger.Error().Err(err).RawJSON("extraDataJSON", extraDataJSON).Msg("Failed to unmarshal extra data")
		helper.GinResponse(ctx, http.StatusBadRequest, helper.ErrorResponseWithMessage(err, "Failed to unmarshal extra data", "Không thể unmarshal dữ liệu thêm"))
		return
	}

	logger.Info().Interface("extraData", extraData).Msg("Decoded extra data")

	newPartnerClientID, err := uuid.Parse(req.PartnerClientID)
	if err != nil {
		logger.Error().Err(err).Str("partnerClientID", req.PartnerClientID).Msg("Failed to parse partner client ID")
		helper.GinResponse(ctx, http.StatusBadRequest, helper.ErrorResponseWithMessage(err, "Failed to parse partner client ID", "Không thể phân tích ID khách hàng đối tác"))
		return
	}

	receiver, err := i.UserService.GetUserByID(newPartnerClientID)
	if err != nil {
		logger.Error().Err(err).Str("userID", newPartnerClientID.String()).Msg("Failed to get receiver details")
		helper.GinResponse(ctx, http.StatusInternalServerError, helper.ErrorResponseWithMessage(err, "Failed to get receiver details", "Không thể lấy thông tin người nhận"))
		return
	}

	var wsMessage schemas.WebSocketMessage
	var notification schemas.Notification

	switch extraData.Type {
	case "linkWallet":
		if req.ResultCode != 0 {
			wsMessage = schemas.WebSocketMessage{UserID: newPartnerClientID.String(), Type: "link-wallet-failed"}
			notification = schemas.Notification{Title: "Liên kết ví thất bại", Body: "Liên kết ví thất bại, vui lòng thử lại", Token: receiver.DeviceToken}
		} else {
			if err := i.IPNService.HandleLinkWalletCallback(req); err != nil {
				logger.Error().Err(err).Msg("Failed to handle linking wallet callback")
				helper.GinResponse(ctx, http.StatusInternalServerError, helper.ErrorResponseWithMessage(err, "Failed to handle linking wallet callback", "Không thể xử lý callback liên kết ví"))
				return
			}
			wsMessage = schemas.WebSocketMessage{UserID: newPartnerClientID.String(), Type: "link-wallet-success"}
			notification = schemas.Notification{Title: "Liên kết ví thành công", Body: "Liên kết ví thành công", Token: receiver.DeviceToken}
		}

	case "payment":
		if req.ResultCode != 0 {
			wsMessage = schemas.WebSocketMessage{UserID: newPartnerClientID.String(), Type: "payment-failed"}
			notification = schemas.Notification{Title: "Thanh toán thất bại", Body: "Thanh toán thất bại, vui lòng thử lại", Token: receiver.DeviceToken}
		} else {
			if err := i.IPNService.HandleIPN(req); err != nil {
				logger.Error().Err(err).Msg("Failed to handle IPN")
				helper.GinResponse(ctx, http.StatusInternalServerError, helper.ErrorResponseWithMessage(err, "Failed to handle IPN", "Không thể xử lý IPN"))
				return
			}
			wsMessage = schemas.WebSocketMessage{UserID: newPartnerClientID.String(), Type: "payment-success"}
			notification = schemas.Notification{Title: "Thanh toán thành công", Body: "Thanh toán thành công", Token: receiver.DeviceToken}
		}
	case "withdraw":
		if req.ResultCode != 0 {
			wsMessage = schemas.WebSocketMessage{UserID: extraData.UserID.String(), Type: "withdraw-failed"}
			notification = schemas.Notification{Title: "Rút tiền thất bại", Body: "Rút tiền thất bại, vui lòng thử lại", Token: receiver.DeviceToken}
		} else {
			if err := i.IPNService.HandleWithdrawIPN(req); err != nil {
				logger.Error().Err(err).Msg("Failed to handle withdraw IPN")
				helper.GinResponse(ctx, http.StatusInternalServerError, helper.ErrorResponseWithMessage(err, "Failed to handle withdraw IPN", "Không thể xử lý IPN rút tiền"))
				return
			}
			wsMessage = schemas.WebSocketMessage{UserID: extraData.UserID.String(), Type: "withdraw-success"}
			notification = schemas.Notification{Title: "Rút tiền thành công", Body: "Rút tiền thành công", Token: receiver.DeviceToken}
		}
	default:
		logger.Error().Str("type", extraData.Type).Msg("Unknown extra data type")
		helper.GinResponse(ctx, http.StatusBadRequest, helper.ErrorResponseWithMessage(fmt.Errorf("unknown extra data type"), "Unknown extra data type", "Loại dữ liệu thêm không xác định"))
		return
	}

	notificationPayload := schemas.NotificationPayload{Type: wsMessage.Type}
	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	if err != nil {
		logger.Error().Err(err).Interface("payload", notificationPayload).Msg("Failed to convert struct to map")
		helper.GinResponse(ctx, http.StatusInternalServerError, helper.ErrorResponseWithMessage(err, "Failed to convert struct to map", "Không thể chuyển đổi struct sang map"))
		return
	}
	notification.Data = notificationPayloadMap

	go func() {
		if err := i.asyncClient.EnqueueWebsocketMessage(wsMessage); err != nil {
			logger.Error().Err(err).Interface("message", wsMessage).Msg("Failed to enqueue websocket message")
		}
	}()

	go func() {
		if err := i.asyncClient.EnqueueFCMNotification(notification); err != nil {
			logger.Error().Err(err).Interface("notification", notification).Msg("Failed to enqueue FCM notification")
		}
	}()

	logger.Info().Msg("IPN handled successfully")
	ctx.Status(http.StatusNoContent)
}
