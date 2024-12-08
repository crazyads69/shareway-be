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
	group.GET("/get-dashboard-general-data", adminController.GetDashboardGeneralData)
	group.GET("/get-user-dashboard-data", adminController.GetUserDashboardData)
	group.GET("/get-ride-dashboard-data", adminController.GetRideDashboardData)
	group.GET("/get-transaction-dashboard-data", adminController.GetTransactionDashboardData)
	group.GET("/get-vehicle-dashboard-data", adminController.GetVehicleDashboardData)
}
