package controller

import (
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
	OTPService  service.IOTPService
	UserService service.IUsersService
	token       *token.PasetoMaker
}

// NewAuthController creates a new AuthController instance
func NewAuthController(cfg util.Config, db *gorm.DB, otpService service.IOTPService, userService service.IUsersService, tokenMaker *token.PasetoMaker) *AuthController {
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
// @Description Starts the registration process by sending an OTP and creating a user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body schemas.RegisterUserRequest true "Registration request containing phone number and full name"
// @Success 200 {object} helper.Response{data=schemas.RegisterUserResponse} "User created and OTP sent successfully"
// @Failure 400 {object} helper.Response "Invalid request body"
// @Failure 409 {object} helper.Response "User already exists"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /auth/register [post]
func (ctrl *AuthController) Register(ctx *gin.Context) {
	var req schemas.RegisterUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
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
			err,
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
	var res schemas.RegisterUserResponse
	// Add phone number to db and return user_id
	userID, fullName, err := ctrl.UserService.CreateUserByPhone(req.PhoneNumber, req.FullName)
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
	res.FullName = fullName

	response := helper.SuccessResponse(res, "OTP sent successfully", "Mã OTP đã được gửi thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// ResendOTP resends the OTP, optionally via voice call
// ResendOTP godoc
// @Summary Resend OTP
// @Description Resends the OTP to the provided phone number
// @Tags auth
// @Accept json
// @Produce json
// @Param request body schemas.GenerateOTPRequest true "OTP resend request"
// @Success 200 {object} helper.Response{data=schemas.GenerateOTPResponse} "OTP sent successfully"
// @Failure 400 {object} helper.Response "Invalid request body"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /auth/resend-otp [post]
func (ctrl *AuthController) ResendOTP(ctx *gin.Context) {

	var req schemas.GenerateOTPRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Yêu cầu không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Send OTP via Twilio (you might need to modify OTPService to support voice calls)
	_, err := ctrl.OTPService.SendOTP(req.PhoneNumber) // Add voice parameter if supported
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to send OTP",
			"Không thể gửi mã OTP",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Get user_id
	userID, err := ctrl.UserService.GetUserIDByPhone(req.PhoneNumber)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get user_id",
			"Không thể lấy user_id",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	var res schemas.GenerateOTPResponse
	res.PhoneNumber = req.PhoneNumber
	res.UserID = userID

	response := helper.SuccessResponse(res, "OTP sent successfully", "Mã OTP đã được gửi thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// VerifyRegisterOTP verifies the OTP and activates the user account
// VerifyRegisterOTP godoc
// @Summary Verify registration OTP
// @Description Verifies the OTP sent during registration and activates the user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body schemas.VerifyRegisterOTPRequest true "OTP verification request"
// @Success 200 {object} helper.Response{data=schemas.VerifyRegisterOTPResponse} "OTP verified and user activated successfully"
// @Failure 400 {object} helper.Response "Invalid request body or OTP verification failed"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /auth/verify-register-otp [post]
func (ctrl *AuthController) VerifyRegisterOTP(ctx *gin.Context) {
	var req schemas.VerifyRegisterOTPRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Yêu cầu không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Verify the OTP
	err := ctrl.OTPService.VerifyOTP(req.PhoneNumber, req.OTP)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"OTP verification failed",
			"Xác minh OTP thất bại",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Update user status to activated (is_activated = true)
	err = ctrl.UserService.ActivateUser(req.PhoneNumber)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to activate user account",
			"Không thể kích hoạt người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Update user status
	err = ctrl.UserService.ActivateUser(req.PhoneNumber)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to update user status",
			"Không thể cập nhật trạng thái người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Get user info
	user, err := ctrl.UserService.GetUserByPhone(req.PhoneNumber)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get user_id",
			"Không thể lấy user_id",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	var res schemas.VerifyRegisterOTPResponse
	res.UserID = user.ID
	res.FullName = user.FullName
	res.PhoneNumber = req.PhoneNumber
	res.IsActivated = true

	response := helper.SuccessResponse(res, "OTP verified successfully", "OTP đã được xác minh thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// // VerifyCCCD verifies the CCCD of the user
// func (ctrl *AuthController) VerifyCCCD(ctx *gin.Context) {
// 	var req schemas.VerifyCCCDRequest

// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		response := helper.ErrorResponseWithMessage(
// 			err,
// 			"Invalid request body",
// 			"Yêu cầu không hợp lệ",
// 		)
// 		helper.GinResponse(ctx, http.StatusBadRequest, response)
// 		return
// 	}

// 	// Verify the CCCD
// 	err := ctrl.UserService.VerifyCCCD(req.UserID, req.CCCD)
// 	if err != nil {
// 		response := helper.ErrorResponseWithMessage(
// 			err,
// 			"CCCD verification failed",
// 			"Xác minh CCCD thất bại",
// 		)
// 		helper.GinResponse(ctx, http.StatusBadRequest, response)
// 		return
// 	}

// 	response := helper.SuccessResponse(nil, "CCCD verified successfully", "CCCD đã được xác minh thành công")
// 	helper.GinResponse(ctx, http.StatusOK, response)
// }
