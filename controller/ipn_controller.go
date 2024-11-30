package controller

import (
	"net/http"
	"shareway/helper"
	"shareway/infra/task"
	"shareway/infra/ws"
	"shareway/schemas"
	"shareway/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

	// Handle IPN case
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

	// Handle IPN case (there is no callbackToken)
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

	// Answer to MOMO payment gateway status 204
	ctx.Status(http.StatusNoContent)
}
