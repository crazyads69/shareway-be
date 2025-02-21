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
		server.Service.RideService,
		server.Service.MapService,
		server.Service.VehicleService,
		server.Service.UserService,
	)
	group.GET("/get-profile", adminController.GetAdminProfile)
	group.GET("/get-dashboard-general-data", adminController.GetDashboardGeneralData)
	group.GET("/get-user-dashboard-data", adminController.GetUserDashboardData)
	group.GET("/get-ride-dashboard-data", adminController.GetRideDashboardData)
	group.GET("/get-transaction-dashboard-data", adminController.GetTransactionDashboardData)
	group.GET("/get-vehicle-dashboard-data", adminController.GetVehicleDashboardData)
	group.GET("/get-user-list", adminController.GetUserList)
	group.GET("/get-ride-list", adminController.GetRideList)
	group.GET("/get-vehicle-list", adminController.GetVehicleList)
	group.GET("/get-transaction-list", adminController.GetTransactionList)
	group.GET("/get-report-details", adminController.GetReportDetails)
	group.POST("/logout", adminController.AdminLogout)
}
