package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

func SetupMapRouter(group *gin.RouterGroup, server *APIServer) {
	mapController := controller.NewMapController(
		server.Service.MapService,
		server.Validate,
	)
	group.GET("/autocomplete", mapController.GetAutoComplete)
}
