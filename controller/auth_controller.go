package controller

import (
	"errors"
	"net/http"

	"shareway/helper"
	"shareway/schemas"
	"shareway/service"
	"shareway/util"
	"shareway/util/token"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthController handles authentication-related requests
type AuthController struct {
	cfg         util.Config
	db          *gorm.DB
	OTPService  *service.OtpService
	UserService *service.UsersService
	token       *token.PasetoMaker
}

// NewAuthController creates a new AuthController instance
func NewAuthController(cfg util.Config, db *gorm.DB, otpService *service.OtpService, userService *service.UsersService, tokenMaker *token.PasetoMaker) *AuthController {
	return &AuthController{
		cfg:         cfg,
		db:          db,
		OTPService:  otpService,
		UserService: userService,
		token:       tokenMaker,
	}
}

// InitiateRegistration starts the registration process by sending an OTP
// InitiateRegistration godoc
// @Summary Initiate user registration
// @Description Starts the registration process by sending an OTP
// @Tags auth
// @Accept json
// @Produce json
// @Param request body schemas.FirstRegisterUserRequest true "Registration request"
// @Success 200 {object} helper.Response{data=schemas.FirstRegisterUserResponse} "OTP sent successfully"
// @Failure 400 {object} helper.Response "Invalid request body"
// @Failure 409 {object} helper.Response "User already exists"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /auth/init-register [post]
func (ctrl *AuthController) InitiateRegistration(ctx *gin.Context) {
	var req schemas.FirstRegisterUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			errors.New("Invalid request body"),
			"Invalid request body",
			"Số điện thoại không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Check if user already exists
	exists, err := ctrl.UserService.UserExistsByPhone(req.PhoneNumber)
	if err != nil {
		// ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to check user existence",
			"Không thể kiểm tra sự tồn tại của người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}
	if exists {
		response := helper.ErrorResponseWithMessage(
			errors.New("User already exists"),
			"User already exists",
			"Người dùng đã tồn tại",
		)
		helper.GinResponse(ctx, http.StatusConflict, response)
		return
	}

	// Send OTP via Twilio
	_, err = ctrl.OTPService.SendOTP(req.PhoneNumber)
	if err != nil {
		// ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to send OTP",
			"Không thể gửi mã OTP",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Generate OTP and return to the user
	var res schemas.FirstRegisterUserResponse
	// Add phone number to db and return user_id
	userID, err := ctrl.UserService.CreateUserByPhone(req.PhoneNumber)
	if err != nil {
		// ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create user",
			"Không thể tạo người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}
	res.UserID = userID
	res.PhoneNumber = req.PhoneNumber

	response := helper.SuccessResponse(res, "OTP sent successfully", "Mã OTP đã được gửi thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// ResendOTP resends the OTP, optionally via voice call
func (ctrl *AuthController) ResendOTP(ctx *gin.Context) {
	var req struct {
		PhoneNumber string `json:"phone_number" binding:"required,e164"`
		ViaVoice    bool   `json:"via_voice"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Implement logic to check if we should allow resending (e.g., rate limiting)

	// Send OTP via Twilio (you might need to modify OTPService to support voice calls)
	sid, err := ctrl.OTPService.SendOTP(req.PhoneNumber) // Add voice parameter if supported
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend OTP"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OTP resent successfully", "sid": sid})
}

// // VerifyOTPAndRegister verifies the OTP and completes the registration process
// func (ctrl *AuthController) VerifyOTPAndRegister(ctx *gin.Context) {
// 	var req struct {
// 		PhoneNumber string `json:"phone_number" binding:"required,e164"`
// 		OTP         string `json:"otp" binding:"required,len=6"`
// 		Name        string `json:"name" binding:"required"`
// 		Email       string `json:"email" binding:"required,email"`
// 	}

// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Verify OTP with Twilio
// 	err := ctrl.OTPService.VerifyOTP(req.PhoneNumber, req.OTP)
// 	if err != nil {
// 		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
// 		return
// 	}

// 	// Create user in the database
// 	user, err := ctrl.UserService.CreateUser(req.Name, req.Email, req.PhoneNumber)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
// 		return
// 	}

// 	// Generate token for the new user
// 	token, payload, err := ctrl.token.CreateToken(user.ID, ctrl.cfg.AccessTokenDuration)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"user":         user,
// 		"access_token": token,
// 		"expires_at":   payload.ExpiredAt,
// 	})
// }

// func (ctrl *AuthController) Login(ctx *gin.Context) {
// 	// Implement login logic here
// }
