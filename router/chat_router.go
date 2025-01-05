package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

func SetupChatRouter(group *gin.RouterGroup, server *APIServer) {
	chatController := controller.NewChatController(
		server.Validate,
		server.Service.RideService,
		server.Service.MapService,
		server.Service.UserService,
		server.Service.VehicleService,
		server.Service.ChatService,
		server.AsyncClient,
		server.Hub,
		server.Agora,
	)
	group.POST("/send-message", chatController.SendMessage)
	group.POST("/send-image", chatController.SendImage)
	group.POST("/get-chat-rooms", chatController.GetAllChatRooms)
	group.POST("/get-chat-messages", chatController.GetChatMessages)
	group.GET("/initiate-call", chatController.InitiateCall)
	group.POST("/update-call-status", chatController.UpdateCallStatus)
	group.POST("/search-users", chatController.SearchUsers)
}
