package router

import (
	controller "shareway/controller"

	"github.com/gin-gonic/gin"
)

// SetupProtectedRouter ...
//	@Summary	test protected endpoint
//	@Schemes
//	@Description				test protected endpoint desc
//	@Tags						Protected branch
//	@Accept						json
//	@Produce					json
//	@BasePath					/auth/login [post]
//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func SetupAuthRouter(group *gin.RouterGroup, server *APIServer) {
	authController := controller.NewAuthController(
		server.Cfg,
		server.Service.OTPService,
		server.Service.UserService,
		server.Maker,
	)
	// RegisterUserRequest
	group.POST("/register", authController.Register)
	// ResendOTPRequest
	group.POST("/resend-otp", authController.ResendOTP)
	// VerifyRegisterOTPRquest
	group.POST("/verify-register-otp", authController.VerifyRegisterOTP)
	// group.POST("/login", auth_controller.Login)
	// group.POST("/register", auth_controller.Login)
	// group.POST("/refresh", auth_controller.Login)
}
