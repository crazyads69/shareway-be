package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

func SetupMapRouter(group *gin.RouterGroup, server *APIServer) {
	mapController := controller.NewMapController(
		server.Service.MapService,
		server.Validate,
		server.Service.VehicleService,
	)
	// GetAutoComplete request
	group.GET("/autocomplete", mapController.GetAutoComplete)

	// CreateGiveRide request
	group.POST("/give-ride", mapController.CreateGiveRide)

	// CreateHitchRide request
	group.POST("/hitch-ride", mapController.CreateHitchRide)

	// GetGeoCode request
	group.POST("/geocode", mapController.GetGeoCode)

}
