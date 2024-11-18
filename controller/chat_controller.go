package controller

import (
	"fmt"
	"log"
	"shareway/helper"
	"shareway/infra/agora"
	"shareway/infra/task"
	"shareway/infra/ws"
	"shareway/middleware"
	"shareway/schemas"
	"shareway/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ChatController struct {
	validate       *validator.Validate
	RideService    service.IRideService
	asyncClient    *task.AsyncClient
	hub            *ws.Hub
	MapsService    service.IMapService
	UserService    service.IUsersService
	VehicleService service.IVehicleService
	ChatService    service.IChatService
	agora          *agora.Agora
}

func NewChatController(validate *validator.Validate, rideService service.IRideService, mapService service.IMapService, userService service.IUsersService, vehicleService service.IVehicleService, chatService service.IChatService, asyncClient *task.AsyncClient, hub *ws.Hub, agora *agora.Agora) *ChatController {
	return &ChatController{
		validate:       validate,
		RideService:    rideService,
		MapsService:    mapService,
		UserService:    userService,
		VehicleService: vehicleService,
		asyncClient:    asyncClient,
		hub:            hub,
		ChatService:    chatService,
		agora:          agora,
	}
}

// SendMessage sends a message to the chat
// SendMessage godoc
// @Summary Send a message to the chat room
// @Description Send a message to the chat room
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.SendMessageRequest true "Send message request"
// @Success 200 {object} helper.Response{data=schemas.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /chat/send-message [post]
func (cc *ChatController) SendMessage(ctx *gin.Context) {
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

	var req schemas.SendMessageRequest
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
	if err := cc.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Send message
	message, err := cc.ChatService.SendMessage(req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to send message",
			"Không thể gửi tin nhắn",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.SendMessageResponse{
		MessageID:  message.ID,
		Message:    message.Message,
		ReceiverID: message.ReceiverID,
		CreatedAt:  message.CreatedAt,
	}

	// Get receiver device token to send notification
	receiver, err := cc.UserService.GetUserByID(req.ReceiverID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get receiver details",
			"Không thể lấy thông tin người nhận",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare websocket message
	wsMessage := schemas.WebSocketMessage{
		Type:    "new-text-message",
		UserID:  req.ReceiverID.String(),
		Payload: res,
	}

	// Prepare notification message
	resMap, err := helper.ConvertToStringMap(res)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert response to map",
			"Không thể chuyển đổi response thành map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Send notification
	notificationPayload := schemas.NotificationPayload{
		Type: "new-text-message",
		Data: resMap,
	}

	// Convert notification payload to map
	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert notification payload to map",
			"Không thể chuyển đổi notification payload thành map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	notification := schemas.Notification{
		Title: "Tin nhắn mới",
		Body:  fmt.Sprintf("Bạn có tin nhắn mới từ %s", receiver.FullName),
		Data:  notificationPayloadMap,
		Token: receiver.DeviceToken,
	}

	go func() {
		err := cc.asyncClient.EnqueueFCMNotification(notification)
		if err != nil {
			log.Printf("failed to send notification: %v", err)
		}
	}()

	go func() {
		err := cc.asyncClient.EnqueueWebsocketMessage(wsMessage)
		if err != nil {
			log.Printf("failed to send websocket message: %v", err)
		}
	}()

	response := helper.SuccessResponse(
		res,
		"Message sent successfully",
		"Tin nhắn đã được gửi thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// SendImage sends an image to the chat room
// SendImage godoc
// @Summary Send an image to the chat room
// @Description Send an image to the chat room
// @Tags chat
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param chatRoomID formData string true "Chat room ID"
// @Param image formData file true "Image file"
// @Param receiverID formData string true "Receiver ID"
// @Success 200 {object} helper.Response{data=schemas.SendImageResponse} "Image sent successfully"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /chat/send-image [post]
func (cc *ChatController) SendImage(ctx *gin.Context) {
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

	var req schemas.SendImageRequest
	// use shouldBind because the request is multipart/form-data
	if err := ctx.ShouldBind(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind form",
			"Không thể bind form",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := cc.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Get image and upload to cloud storage
	chat, err := cc.ChatService.UploadImage(ctx.Request.Context(), req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to upload image",
			"Không thể upload ảnh",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.SendImageResponse{
		ImageURL:   chat.Message,
		ReceiverID: chat.ReceiverID.String(),
		CreatedAt:  chat.CreatedAt,
		MessageID:  chat.ID.String(),
	}

	// Get receiver device token to send notification
	receiver, err := cc.UserService.GetUserByID(chat.ReceiverID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get receiver details",
			"Không thể lấy thông tin người nhận",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Prepare websocket message
	wsMessage := schemas.WebSocketMessage{
		Type:    "new-image-message",
		UserID:  req.ReceiverID,
		Payload: res,
	}

	// Prepare notification message
	resMap, err := helper.ConvertToStringMap(res)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert response to map",
			"Không thể chuyển đổi response thành map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Send notification
	notificationPayload := schemas.NotificationPayload{
		Type: "new-image-message",
		Data: resMap,
	}

	// Convert notification payload to map
	notificationPayloadMap, err := helper.ConvertToStringMap(notificationPayload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to convert notification payload to map",
			"Không thể chuyển đổi notification payload thành map",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	notification := schemas.Notification{
		Title: "Tin nhắn mới",
		Body:  fmt.Sprintf("Bạn có tin nhắn mới từ %s", receiver.FullName),
		Data:  notificationPayloadMap,
		Token: receiver.DeviceToken,
	}

	go func() {
		err := cc.asyncClient.EnqueueFCMNotification(notification)
		if err != nil {
			log.Printf("failed to send notification: %v", err)
		}
	}()

	go func() {
		err := cc.asyncClient.EnqueueWebsocketMessage(wsMessage)
		if err != nil {
			log.Printf("failed to send websocket message: %v", err)
		}
	}()

	response := helper.SuccessResponse(
		res,
		"Image sent successfully",
		"Ảnh đã được gửi thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// GetAllChatRooms gets all chat rooms of a user
// GetAllChatRooms godoc
// @Summary Get all chat rooms of a user
// @Description Get all chat rooms of a user
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.GetAllChatRoomsRequest true "Get all chat rooms request"
// @Success 200 {object} helper.Response{data=schemas.GetAllChatRoomsResponse} "Chat rooms fetched successfully"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /chat/get-chat-rooms [post]
func (cc *ChatController) GetAllChatRooms(ctx *gin.Context) {
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

	// Get all chat rooms
	chatRooms, err := cc.ChatService.GetAllChatRooms(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get chat rooms",
			"Không thể lấy danh sách chat rooms",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	userInfos := make([]schemas.UserInfo, len(chatRooms))
	for i, room := range chatRooms {
		// Get receiver info
		// Check who is the receiver
		var receiverID uuid.UUID
		if room.User1ID == data.UserID {
			receiverID = room.User2ID
		} else {
			receiverID = room.User1ID
		}
		receiver, err := cc.UserService.GetUserByID(receiverID) // Make sure to get the receiver info
		if err != nil {
			response := helper.ErrorResponseWithMessage(
				err,
				"Failed to get receiver info",
				"Không thể lấy thông tin người nhận",
			)
			helper.GinResponse(ctx, 500, response)
			return
		}

		userInfos[i] = schemas.UserInfo{
			ID:          receiver.ID,
			FullName:    receiver.FullName,
			PhoneNumber: receiver.PhoneNumber,
		}
	}

	res := schemas.GetAllChatRoomsResponse{
		ChatRooms: make([]schemas.ChatRoomResponse, len(chatRooms)),
	}

	for i, room := range chatRooms {
		res.ChatRooms[i] = schemas.ChatRoomResponse{
			ID:            room.ID,
			ReceiverInfo:  userInfos[i],
			LastMessage:   room.LastMessageText,
			LastMessageAt: room.LastMessageAt,
			LastMessageID: room.LastMessageID,
		}
	}

	response := helper.SuccessResponse(
		res,
		"Chat rooms fetched successfully",
		"Danh sách chat rooms đã được lấy thành công",
	)
	helper.GinResponse(ctx, 200, response)
}

// GetChatMessages gets all messages of a chat room
// GetChatMessages godoc
// @Summary Get all messages of a chat room
// @Description Get all messages of a chat room
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.GetChatMessagesRequest true "Get chat messages request"
// @Success 200 {object} helper.Response{data=schemas.GetChatMessagesResponse} "Chat messages fetched successfully"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /chat/get-chat-messages [post]
func (cc *ChatController) GetChatMessages(ctx *gin.Context) {
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

	var req schemas.GetChatMessagesRequest
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
	if err := cc.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Get all messages of a chat room
	messages, err := cc.ChatService.GetChatMessages(req, data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get chat messages",
			"Không thể lấy tin nhắn",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.GetChatMessagesResponse{
		Messages: make([]schemas.MessageResponse, len(messages)),
	}

	for i, message := range messages {
		res.Messages[i] = schemas.MessageResponse{
			ID:          message.ID,
			Message:     message.Message,
			SenderID:    message.SenderID,
			ReceiverID:  message.ReceiverID,
			CreatedAt:   message.CreatedAt,
			MessageType: message.MessageType,
		}
	}

	response := helper.SuccessResponse(
		res,
		"Chat messages fetched successfully",
		"Tin nhắn đã được lấy thành công",
	)
	helper.GinResponse(ctx, 200, response)
}
