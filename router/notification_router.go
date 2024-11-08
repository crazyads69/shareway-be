package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

func SetupNotificationRouter(group *gin.RouterGroup, server *APIServer) {
	notificationController := controller.NewNotificationController(
		server.Validate,
		server.Service.NotificationService,
	)
	// CreateNotification request
	group.POST("/create-notification", notificationController.CreateNotification)
	// CreateTestWebsocket request
	group.POST("/create-test-websocket", notificationController.CreateTestWebsocket)

}
