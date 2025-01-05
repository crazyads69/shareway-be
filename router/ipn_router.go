package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

func SetupIPNRouter(group *gin.RouterGroup, server *APIServer) {
	ipnController := controller.NewIPNController(
		server.Validate,
		server.Hub,
		server.Service.RideService,
		server.Service.MapService,
		server.Service.UserService,
		server.Service.VehicleService,
		server.Service.PaymentService,
		server.Service.IPNService,
		server.AsyncClient,
	)
	// Link momo wallet to user account
	group.POST("/handle-ipn", ipnController.HandleIPN)
}
