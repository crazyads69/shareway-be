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

type NotificationController struct {
	validate *validator.Validate

	NotificationService service.INotificationService
}

func NewNotificationController(validate *validator.Validate, notificationService service.INotificationService) *NotificationController {
	return &NotificationController{
		validate:            validate,
		NotificationService: notificationService,
	}
}

// CreateNotification lets you create a new notification to be sent to the user device using the FCM service
// CreateNotification godoc
// @Summary Create a new notification
// @Description Create a new notification to be sent to the user device using the FCM service
// @Tags notification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.CreateNotificationRequest true "Create notification request"
// @Success 200 {object} helper.Response "Successfully created notification"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /notification/create-notification [post]
func (nc *NotificationController) CreateNotification(ctx *gin.Context) {
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

	var req schemas.CreateNotificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind JSON",
			"Không thể bind JSON",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	if err := nc.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Validation failed",
			"Validation thất bại",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	notificationID, err := nc.NotificationService.CreateNotification(req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create notification",
			"Không thể tạo thông báo",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.CreateNotificationResponse{
		NotificationID: notificationID,
	}

	response := helper.SuccessResponse(res, "Notification created successfully", "Tạo thông báo thành công")
	helper.GinResponse(ctx, 200, response)

}

// CreateTestWebsocket create a test websocket for testing connection
// CreateTestWebsocket godoc
// @Summary Create a test websocket
// @Description Create a test websocket for testing connection
// @Tags notification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.CreateTestWebsocketRequest true "Create test websocket request"
// @Success 200 {object} helper.Response "Successfully created test websocket"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /notification/create-test-websocket [post]
func (nc *NotificationController) CreateTestWebsocket(ctx *gin.Context) {
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

	var req schemas.CreateTestWebsocketRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind JSON",
			"Không thể bind JSON",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	if err := nc.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Validation failed",
			"Validation thất bại",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	err = nc.NotificationService.CreateTestWebsocket(req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create test websocket",
			"Không thể tạo websocket test",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	response := helper.SuccessResponse(nil, "Test websocket created successfully", "Tạo websocket test thành công")
	helper.GinResponse(ctx, 200, response)
}
