package router

import (
	"shareway/controller"

	"github.com/gin-gonic/gin"
)

// SetupProtectedRouter configures the protected routes
func SetupProtectedRouter(group *gin.RouterGroup, server *APIServer) {
	protectedController := controller.NewProtectedController(server.Cfg)
	group.GET("/test", protectedController.ProtectedEndpoint)
}
