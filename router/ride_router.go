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
		server.Service.PaymentService,
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
	group.POST("/cancel-ride", rideController.CancelRide)
	group.GET("/get-all-pending-ride", rideController.GetAllPendingRide)
	group.POST("/rating-ride-hitcher", rideController.RatingRideHitcher)
	group.POST("/rating-ride-driver", rideController.RatingRideDriver)
	group.GET("/get-ride-history", rideController.GetRideHistory)
}
