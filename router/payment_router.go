package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

func SetupPaymentRouter(group *gin.RouterGroup, server *APIServer) {
	paymentController := controller.NewPaymentController(
		server.Validate,
		server.Hub,
		server.Service.RideService,
		server.Service.MapService,
		server.Service.UserService,
		server.Service.VehicleService,
		server.Service.PaymentService,
		server.AsyncClient,
	)
	// Link momo wallet to user account
	group.POST("/link-momo-wallet", paymentController.LinkMomoWallet)
}
