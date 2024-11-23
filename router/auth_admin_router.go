package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

func SetupAuthAdminRouter(group *gin.RouterGroup, server *APIServer) {
	authAdminController := controller.NewAuthAdminController(
		server.Cfg,
		server.Validate,
		server.Service.AdminService,
	)
	group.POST("/login", authAdminController.Login)
}
