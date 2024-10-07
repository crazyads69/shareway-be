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
	)
	// RegisterUserRequest
	group.POST("/register", authController.Register)
	// LoginWithPhoneNumberRequest
	group.POST("/login-phone", authController.LoginWithPhoneNumber)
	// LoginWithOAuthRequest
	group.POST("/login-oauth", authController.LoginWithOAuth)
	// ResendOTPRequest
	group.POST("/resend-otp", authController.ResendOTP)
	// VerifyRegisterOTPRquest
	group.POST("/verify-register-otp", authController.VerifyRegisterOTP)
	// VerifyCCCDRequest
	group.POST("/verify-cccd", authController.VerifyCCCD)
	// RegisterOAuthRequest
	group.POST("/register-oauth", authController.RegisterOAuth)
	// VerifyLoginOTPRequest
	group.POST("/verify-login-otp", authController.VerifyLoginOTP)
	// group.POST("/login", auth_controller.Login)
}
