package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

func SetupVehicleRouter(group *gin.RouterGroup, server *APIServer) {
	vehicleController := controller.NewVehicleController(
		server.Service.VehicleService,
		server.Validate,
	)
	group.GET("/vehicles", vehicleController.GetVehicles)
	group.POST("/register-vehicle", vehicleController.RegisterVehicle)
	// group.GET("/vehicles/:id", vehicleController.GetVehicle)
	// group.POST("/vehicles", vehicleController.CreateVehicle)
	// group.PUT("/vehicles/:id", vehicleController.UpdateVehicle)
	// group.DELETE("/vehicles/:id", vehicleController.DeleteVehicle)
}
