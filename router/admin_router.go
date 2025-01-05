package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

func SetupAdminRouter(group *gin.RouterGroup, server *APIServer) {
	adminController := controller.NewAdminController(
		server.Cfg,
		server.Validate,
		server.Service.AdminService,
	)
	group.GET("/get-profile", adminController.GetAdminProfile)
}
