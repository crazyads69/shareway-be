package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

func SetupChatRouter(group *gin.RouterGroup, server *APIServer) {
	chatsController := controller.NewChatController(
		server.Validate,
		server.Service.RideService,
		server.Service.MapService,
		server.Service.UserService,
		server.Service.VehicleService,
		server.Service.ChatService,
		server.AsyncClient,
		server.Hub,
	)
	group.POST("/send-message", chatsController.SendMessage)
	group.POST("/send-image", chatsController.SendImage)
	group.POST("/get-chat-rooms", chatsController.GetAllChatRooms)
	group.POST("/get-chat-messages", chatsController.GetChatMessages)
}
