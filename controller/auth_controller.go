package controller

import (
	"net/http"

	"shareway/helper"
	"shareway/schemas"
	"shareway/service"
	"shareway/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthController handles authentication-related requests
type AuthController struct {
	cfg         util.Config
	OTPService  service.IOTPService
	UserService service.IUsersService
}

// NewAuthController creates a new AuthController instance
func NewAuthController(cfg util.Config, otpService service.IOTPService, userService service.IUsersService) *AuthController {
	return &AuthController{
		cfg:         cfg,
		OTPService:  otpService,
		UserService: userService,
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

	res := schemas.RegisterUserResponse{
		UserID:      userID,
		PhoneNumber: req.PhoneNumber,
		FullName:    fullName,
	}

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

	res := schemas.GenerateOTPResponse{
		PhoneNumber: req.PhoneNumber,
		UserID:      userID,
	}

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

	res := schemas.VerifyRegisterOTPResponse{
		UserID:      user.ID,
		PhoneNumber: user.PhoneNumber,
		FullName:    user.FullName,
		IsActivated: true,
	}

	response := helper.SuccessResponse(res, "OTP verified successfully", "OTP đã được xác minh thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// VerifyCCCD verifies the CCCD (Citizen Identity Card) of the user
// @Summary Verify user's CCCD
// @Description Verifies the front and back images of a user's CCCD, saves the information, and updates user status
// @Tags auth
// @Accept multipart/form-data
// @Produce json
// @Param front_image formData file true "Front image of CCCD"
// @Param back_image formData file true "Back image of CCCD"
// @Param user_id formData string true "User ID (UUID format)"
// @Param phone_number formData string true "User's phone number (9-11 digits)" minlength(9) maxlength(11)
// @Success 200 {object} helper.Response{data=schemas.VerifyCCCDResponse} "CCCD verified successfully"
// @Failure 400 {object} helper.Response "Invalid request or CCCD info"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /auth/verify-cccd [post]
func (ctrl *AuthController) VerifyCCCD(ctx *gin.Context) {
	var req schemas.VerifyCCCDRequest

	// We use ShouldBind instead of ShouldBindJSON because the request is multipart/form-data
	if err := ctx.ShouldBind(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Yêu cầu không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Convert UserID string to UUID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid user ID",
			"ID người dùng không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Call FPT AI to verify the CCCD
	frontCCCDInfo, err := ctrl.UserService.VerifyCCCD(req.FrontImage)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to verify CCCD",
			"Không thể xác minh CCCD",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	backCCCDInfo, err := ctrl.UserService.VerifyCCCD(req.BackImage)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to verify CCCD",
			"Không thể xác minh CCCD",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Check if the CCCD issued date and expiry date are valid
	if err := helper.ValidateCCCDInfo(frontCCCDInfo, backCCCDInfo); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid CCCD info",
			"Thông tin CCCD không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}
	// Encrypt and save the CCCD info
	err = ctrl.UserService.EncryptAndSaveCCCDInfo(frontCCCDInfo, userID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to save CCCD info",
			"Không thể lưu thông tin CCCD",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Update user status to verified (is_verified = true)
	err = ctrl.UserService.VerifyUser(req.PhoneNumber)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to verify user",
			"Không thể xác minh người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Create a new session
	user, accessToken, refreshToken, err := ctrl.UserService.CreateSession(req.PhoneNumber, userID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create session",
			"Không thể tạo phiên xác thực người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	res := schemas.VerifyCCCDResponse{
		User: schemas.UserResponse{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			PhoneNumber: user.PhoneNumber,
			Email:       user.Email,
			FullName:    user.FullName,
			IsVerified:  user.IsVerified,
			IsActivated: user.IsActivated,
			Role:        user.Role,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	response := helper.SuccessResponse(res, "CCCD verified successfully", "CCCD đã được xác minh thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}
