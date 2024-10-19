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
		server.Validate,
		server.Service.OTPService,
		server.Service.UserService,
	)
	// InitRegisterRequest
	group.POST("/init-register", authController.InitRegister)
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
	// VerifyLoginOTPRequest
	group.POST("/verify-login-otp", authController.VerifyLoginOTP)
	// RefreshTokenRequest
	group.POST("/refresh-token", authController.RefreshToken)
	// LogoutRequest
	group.POST("/logout", authController.Logout)
	// DeleteUser
	group.POST("/delete-user", authController.DeleteUser)
}
