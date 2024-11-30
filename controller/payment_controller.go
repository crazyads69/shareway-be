package controller

import (
	"fmt"
	"shareway/helper"
	"shareway/infra/task"
	"shareway/infra/ws"
	"shareway/middleware"
	"shareway/schemas"
	"shareway/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PaymentController struct {
	validate       *validator.Validate
	hub            *ws.Hub
	RideService    service.IRideService
	MapsService    service.IMapService
	UserService    service.IUsersService
	VehicleService service.IVehicleService
	PaymentService service.IPaymentService
	asyncClient    *task.AsyncClient
}

func NewPaymentController(validate *validator.Validate, hub *ws.Hub, rideService service.IRideService,
	mapService service.IMapService, userService service.IUsersService, vehicleService service.IVehicleService,
	paymentService service.IPaymentService,
	asyncClient *task.AsyncClient) *PaymentController {
	return &PaymentController{
		validate:       validate,
		hub:            hub,
		RideService:    rideService,
		MapsService:    mapService,
		UserService:    userService,
		VehicleService: vehicleService,
		asyncClient:    asyncClient,
	}
}

// LinkMomoWallet godoc
// @Summary Link momo wallet to user account
// @Description Link momo wallet to user account
// @Tags payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} schemas.LinkMomoWalletResponse "Link momo wallet response"
// @Failure 400 {object} helper.Response "Bad request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /payment/link-momo-wallet [post]
func (p *PaymentController) LinkMomoWallet(ctx *gin.Context) {
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

	// Get user info check if user already linked momo wallet
	user, err := p.UserService.GetUserByID(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get user info",
			"Không thể lấy thông tin người dùng",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// If user already linked momo wallet, return error response
	if user.IsMomoLinked {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("user already linked momo wallet"),
			"User already linked momo wallet",
			"Người dùng đã liên kết ví momo",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Link momo wallet to user account
	momo, err := p.PaymentService.LinkMomoWallet(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to link momo wallet",
			"Không thể liên kết ví momo",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.LinkMomoWalletResponse{
		Deeplink: momo.Deeplink,
	}

	response := helper.SuccessResponse(res, "Link momo wallet successfully", "Liên kết ví momo thành công")
	helper.GinResponse(ctx, 200, response)
}
