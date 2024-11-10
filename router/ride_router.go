package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

func SetupRideRouter(group *gin.RouterGroup, server *APIServer) {
	rideController := controller.NewRideController(
		server.Validate,
		server.Hub,
		server.Service.RideService,
		server.Service.MapService,
		server.Service.UserService,
		server.Service.VehicleService,
		server.AsyncClient,
	)
	group.POST("/give-ride-request", rideController.SendGiveRideRequest)
	group.POST("/hitch-ride-request", rideController.SendHitchRideRequest)
	group.POST("/accept-give-ride-request", rideController.AcceptGiveRideRequest)
	group.POST("/accept-hitch-ride-request", rideController.AcceptHitchRideRequest)
	group.POST("/cancel-give-ride-request", rideController.CancelGiveRideRequest)
	group.POST("/cancel-hitch-ride-request", rideController.CancelHitchRideRequest)
	group.POST("/start-ride", rideController.StartRide)
	group.POST("/end-ride", rideController.EndRide)
	group.POST("/update-ride-location", rideController.UpdateRideLocation)
}
