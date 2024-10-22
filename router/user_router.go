package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

// SetupUserRouter ...
func SetupUserRouter(group *gin.RouterGroup, server *APIServer) {
	userController := controller.NewUserController(
		server.Service.UserService,
		server.Validate,
	)
	// GetUserProfile Request
	group.GET("/get-profile", userController.GetUserProfile)
	// RegisterDeviceToken Request
	group.POST("/register-device-token", userController.RegisterDeviceToken)
	// UpdateUserProfile Request
	group.POST("/update-profile", userController.UpdateUserProfile)
}
