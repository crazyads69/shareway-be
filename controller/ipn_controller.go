package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"shareway/helper"
	"shareway/infra/task"
	"shareway/infra/ws"
	"shareway/schemas"
	"shareway/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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
	var req schemas.MoMoIPN
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get admin profile",
			"Không thể lấy thông tin admin",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Verify IPN signature
	if !i.IPNService.VerifyIPN(req) {
		response := helper.ErrorResponseWithMessage(
			nil,
			"Failed to verify IPN",
			"Không thể xác minh IPN",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Decode extra data and check for type
	extraDataJSON, err := base64.StdEncoding.DecodeString(req.ExtraData)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to decode extra data",
			"Không thể giải mã dữ liệu thêm",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	var extraData schemas.ExtraData
	err = json.Unmarshal(extraDataJSON, &extraData)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to unmarshal extra data",
			"Không thể unmarshal dữ liệu thêm",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Handle IPN case
	// Use extraData.Type to determine the type of IPN request to handle
	switch extraData.Type {
	case "linkWallet":
		// Check if result code is 0
		if req.ResultCode != 0 {
			// Send websocket message to user
			// Get the device token of the hitcher to send notification
			newPartnerClientID, err := uuid.Parse(req.PartnerClientID)
			if err != nil {
				response := helper.ErrorResponseWithMessage(
					err,
					"Failed to parse partner client ID",
					"Không thể phân tích ID khách hàng đối tác",
				)
				helper.GinResponse(ctx, 500, response)
				return
			}
			receiver, err := i.UserService.GetUserByID(newPartnerClientID)
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
				UserID:  newPartnerClientID.String(),
				Type:    "link-wallet-failed",
				Payload: nil,
			}

			// Prepare the notification payload
			notificationPayload := schemas.NotificationPayload{
				Type: "cancel-ride-by-driver",
				Data: nil,
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
				Title: "Liên kết ví thất bại",
				Body:  "Liên kết ví thất bại, vui lòng thử lại",
				Token: receiver.DeviceToken,
				Data:  notificationPayloadMap,
			}

			// Send the WebSocket message using the async client
			go func() {
				err := i.asyncClient.EnqueueWebsocketMessage(wsMessage)
				if err != nil {
					log.Printf("Failed to enqueue websocket message: %v", err)
				}
			}()

			// Send the notification message using the async client
			go func() {
				err = i.asyncClient.EnqueueFCMNotification(notification)
				if err != nil {
					log.Printf("Failed to enqueue FCM notification: %v", err)
				}
			}()

		}
		err := i.IPNService.HandleLinkWalletCallback(req)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to handle linking wallet callback",
				"Không thể xử lý callback liên kết ví",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}

		// If run to this point, it means the linking wallet is successful
		// Send websocket message to user
		// Get the device token of the hitcher to send notification
		newPartnerClientID, err := uuid.Parse(req.PartnerClientID)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to parse partner client ID",
				"Không thể phân tích ID khách hàng đối tác",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}
		receiver, err := i.UserService.GetUserByID(newPartnerClientID)
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
			UserID:  newPartnerClientID.String(),
			Type:    "link-wallet-success",
			Payload: nil,
		}

		// Prepare the notification payload
		notificationPayload := schemas.NotificationPayload{
			Type: "link-wallet-success",
			Data: nil,
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
			Title: "Liên kết ví thành công",
			Body:  "Liên kết ví thành công",
			Token: receiver.DeviceToken,
			Data:  notificationPayloadMap,
		}

		// Send the WebSocket message using the async client
		go func() {
			err := i.asyncClient.EnqueueWebsocketMessage(wsMessage)
			if err != nil {
				log.Printf("Failed to enqueue websocket message: %v", err)
			}
		}()

		// Send the notification message using the async client
		go func() {
			err = i.asyncClient.EnqueueFCMNotification(notification)
			if err != nil {
				log.Printf("Failed to enqueue FCM notification: %v", err)
			}
		}()

	case "payment":
		err := i.IPNService.HandleIPN(req)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to handle IPN",
				"Không thể xử lý IPN",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}
	default:
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("unknown extra data type"),
			"Unknown extra data type",
			"Loại dữ liệu thêm không xác định",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}
	// If callbackToken is not empty, it means this is a callback from the user after linking the wallet successfully
	if req.CallbackToken != "" {
		// Handle linking wallet callback token to get aesToken
		err := i.IPNService.HandleLinkWalletCallback(req)
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to handle linking wallet callback",
				"Không thể xử lý callback liên kết ví",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}
	}

	// Answer to MOMO payment gateway status 204
	ctx.Status(http.StatusNoContent)
}
