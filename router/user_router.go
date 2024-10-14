package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

// SetupUserRouter ...
func SetupUserRouter(
	group *gin.RouterGroup,
	server *APIServer,
) {
	userController := controller.NewUserController(
		server.Service.UserService,
	)
	// GetUserProfile Request
	group.GET("/get-profile", userController.GetUserProfile)
}
