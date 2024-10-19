package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

func SetupMapsRouter(group *gin.RouterGroup, server *APIServer) {
	mapController := controller.NewMapsController(
		server.Service.MapsService,
		server.Validate,
	)
	group.GET("/auto-complete", mapController.GetAutoComplete)
}
