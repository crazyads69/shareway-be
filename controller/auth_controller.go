package controller

import (
	"fmt"
	"net/http"
	"strings"

	"shareway/helper"
	"shareway/schemas"
	"shareway/service"
	"shareway/util"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// AuthController handles authentication-related requests
type AuthController struct {
	cfg         util.Config
	validate    *validator.Validate
	OTPService  service.IOTPService
	UserService service.IUsersService
}

// NewAuthController creates a new AuthController instance
func NewAuthController(cfg util.Config, validate *validator.Validate, otpService service.IOTPService, userService service.IUsersService) *AuthController {
	return &AuthController{
		cfg:         cfg,
		validate:    validate,
		OTPService:  otpService,
		UserService: userService,
	}
}

// InitRegister initializes the registration process by sending an OTP
// InitRegister godoc
// @Summary Initialize user registration
// @Description Start the registration process by sending an OTP to the provided phone number
// @Tags auth
// @Accept json
// @Produce json
// @Param request body schemas.InitRegisterRequest true "Registration initialization request"
// @Success 200 {object} helper.Response{data=schemas.InitRegisterResponse} "OTP sent successfully"
// @Failure 400 {object} helper.Response "Invalid request body or input"
// @Failure 409 {object} helper.Response "User already exists"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /auth/init-register [post]
func (ctrl *AuthController) InitRegister(ctx *gin.Context) {
	var req schemas.InitRegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Số điện thoại không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Validate user input
	if err := ctrl.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response := helper.ErrorResponseWithMessage(
			validationErrors,
			"Invalid input",
			"Dữ liệu đầu vào không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Check if user already exists
	exists, err := ctrl.UserService.UserExistsByPhone(req.PhoneNumber)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to check user existence",
			"Không thể kiểm tra sự tồn tại của người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// If user already exists, return error
	if exists {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("User already exists"),
			"User already exists",
			"Người dùng đã tồn tại",
		)
		helper.GinResponse(ctx, http.StatusConflict, response)
		return
	}

	// Send OTP via Twilio
	_, err = ctrl.OTPService.SendOTP(ctx.Request.Context(), req.PhoneNumber)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to send OTP",
			"Không thể gửi mã OTP",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// When register complete, mean user is not activated and not verified
	res := schemas.InitRegisterResponse{
		PhoneNumber: req.PhoneNumber,
		IsActivated: false,
		IsVerified:  false,
	}

	response := helper.SuccessResponse(res, "OTP sent successfully", "Mã OTP đã được gửi thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// Register starts the registration process and creating a user account
// Register godoc
// @Summary Register a new user
// @Description Starts the registration process and creates a user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body schemas.RegisterUserRequest true "Registration request containing phone number, full name, and optional email"
// @Success 200 {object} helper.Response{data=schemas.RegisterUserResponse} "User created successfully"
// @Failure 400 {object} helper.Response "Invalid request body or input"
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

	// Validate user input
	if err := ctrl.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response := helper.ErrorResponseWithMessage(
			validationErrors,
			"Invalid input",
			"Dữ liệu đầu vào không hợp lệ",
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
			fmt.Errorf("User already exists"),
			"User already exists",
			"Người dùng đã tồn tại",
		)
		helper.GinResponse(ctx, http.StatusConflict, response)
		return
	}

	// Add phone number to db and return user_id
	userID, err := ctrl.UserService.CreateUser(req.PhoneNumber, req.FullName, req.Email)
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

	// Activate user
	err = ctrl.UserService.ActivateUser(req.PhoneNumber)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to activate user",
			"Không thể kích hoạt người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// When register complete, mean user is not activated and not verified
	res := schemas.RegisterUserResponse{
		UserID:      userID,
		PhoneNumber: req.PhoneNumber,
		FullName:    req.FullName,
		IsActivated: true,
		IsVerified:  false,
	}

	response := helper.SuccessResponse(res, "User created successfully", "Người dùng đã được tạo thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// ResendOTP resends the OTP, optionally via voice call
// ResendOTP godoc
// @Summary Resend OTP
// @Description Resends the OTP to the provided phone number for user verification
// @Tags auth
// @Accept json
// @Produce json
// @Param request body schemas.ResendOTPRequest true "OTP resend request containing phone number"
// @Success 200 {object} helper.Response{data=object} "OTP sent successfully"
// @Failure 400 {object} helper.Response "Invalid request body or input"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /auth/resend-otp [post]
func (ctrl *AuthController) ResendOTP(ctx *gin.Context) {

	var req schemas.ResendOTPRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Yêu cầu không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Validate user input
	if err := ctrl.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response := helper.ErrorResponseWithMessage(
			validationErrors,
			"Invalid input",
			"Dữ liệu đầu vào không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Send OTP via Twilio (you might need to modify OTPService to support voice calls)
	_, err := ctrl.OTPService.SendOTP(ctx.Request.Context(), req.PhoneNumber) // Add voice parameter if supported
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to send OTP",
			"Không thể gửi mã OTP",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// RETURN EMPTY DATA
	res := map[string]interface{}{}

	response := helper.SuccessResponse(res, "OTP sent successfully", "Mã OTP đã được gửi thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// VerifyRegisterOTP verifies the OTP and activates the user account
// VerifyRegisterOTP godoc
// @Summary Verify registration OTP
// @Description Verifies the OTP sent during registration
// @Tags auth
// @Accept json
// @Produce json
// @Param request body schemas.VerifyRegisterOTPRequest true "OTP verification request containing phone number and OTP"
// @Success 200 {object} helper.Response{data=object} "OTP verified successfully"
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

	// Validate user input
	if err := ctrl.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response := helper.ErrorResponseWithMessage(
			validationErrors,
			"Invalid input",
			"Dữ liệu đầu vào không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Verify the OTP
	err := ctrl.OTPService.VerifyOTP(ctx.Request.Context(), req.PhoneNumber, req.OTP)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"OTP verification failed",
			"Xác minh OTP thất bại",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// After OTP verified return empty data
	res := map[string]interface{}{}

	response := helper.SuccessResponse(res, "OTP verified successfully", "OTP đã được xác minh thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// VerifyCCCD verifies the CCCD (Citizen Identity Card) of the user
// VerifyCCCD godoc
// @Summary Verify user's CCCD (Citizen Identity Card)
// @Description Verifies the front and back images of a user's CCCD, saves the information, and updates user status
// @Tags auth
// @Accept multipart/form-data
// @Produce json
// @Param front_image formData file true "Front image of CCCD"
// @Param back_image formData file true "Back image of CCCD"
// @Param user_id formData string true "User ID (UUID format)"
// @Param phone_number formData string true "User's phone number (9-11 digits)"
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

	// Validate user input
	if err := ctrl.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response := helper.ErrorResponseWithMessage(
			validationErrors,
			"Invalid input",
			"Dữ liệu đầu vào không hợp lệ",
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

	frontChan := make(chan schemas.VerifyResult)
	backChan := make(chan schemas.VerifyResult)

	// Perform front and back image verifications concurrently
	go func() {
		info, err := ctrl.UserService.VerifyCCCD(req.FrontImage)
		frontChan <- schemas.VerifyResult{Info: info, Err: err}
	}()

	go func() {
		info, err := ctrl.UserService.VerifyCCCD(req.BackImage)
		backChan <- schemas.VerifyResult{Info: info, Err: err}
	}()

	// Wait for both verifications to complete
	frontResult := <-frontChan
	backResult := <-backChan

	// Check for errors in either verification
	if frontResult.Err != nil {
		response := helper.ErrorResponseWithMessage(
			frontResult.Err,
			"Failed to verify front CCCD",
			"Không thể xác minh mặt trước CCCD",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	if backResult.Err != nil {
		response := helper.ErrorResponseWithMessage(
			backResult.Err,
			"Failed to verify back CCCD",
			"Không thể xác minh mặt sau CCCD",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Check if the CCCD issued date and expiry date are valid
	if err := helper.ValidateCCCDInfo(frontResult.Info, backResult.Info); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid CCCD info",
			"Thông tin CCCD không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Encrypt and save the CCCD info
	err = ctrl.UserService.EncryptAndSaveCCCDInfo(frontResult.Info, userID)
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

// Login with phone number
// LoginWithPhoneNumber godoc
// @Summary Login with phone number
// @Description Initiates login process by sending OTP to the provided phone number
// @Tags auth
// @Accept json
// @Produce json
// @Param request body schemas.LoginRequest true "Phone number for login"
// @Success 200 {object} helper.Response{data=schemas.LoginResponse} "OTP sent successfully"
// @Failure 400 {object} helper.Response "Invalid request body"
// @Failure 404 {object} helper.Response "User does not exist"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /auth/login-phone [post]
func (ctrl *AuthController) LoginWithPhoneNumber(ctx *gin.Context) {
	var req schemas.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Yêu cầu không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Validate user input
	if err := ctrl.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response := helper.ErrorResponseWithMessage(
			validationErrors,
			"Invalid input",
			"Dữ liệu đầu vào không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Check if user exists
	exists, err := ctrl.UserService.UserExistsByPhone(req.PhoneNumber)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to check user existence",
			"Không thể kiểm tra sự tồn tại của người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}
	if !exists {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("User does not exist"),
			"User does not exist",
			"Người dùng không tồn tại",
		)
		helper.GinResponse(ctx, http.StatusNotFound, response)
		return
	}

	// Send OTP via Twilio
	_, err = ctrl.OTPService.SendOTP(ctx.Request.Context(), req.PhoneNumber)
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
	user, err := ctrl.UserService.GetUserByPhone(req.PhoneNumber)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get user info",
			"Không thể lấy thông tin người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	res := schemas.LoginResponse{
		PhoneNumber: req.PhoneNumber,
		UserID:      user.ID,
		IsActivated: user.IsActivated,
		IsVerified:  user.IsVerified,
	}

	response := helper.SuccessResponse(res, "OTP sent successfully", "Mã OTP đã được gửi thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// VerifyLoginOTP verifies the OTP and get user info, access token, and refresh token
// VerifyLoginOTP godoc
// @Summary Verify login OTP and create user session
// @Description Verifies the OTP for login, creates a user session, and returns user info with access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body schemas.VerifyLoginOTPRequest true "OTP verification details"
// @Success 200 {object} helper.Response{data=schemas.VerifyLoginOTPResponse} "OTP verified successfully"
// @Failure 400 {object} helper.Response "Invalid request body or OTP verification failed"
// @Failure 500 {object} helper.Response "Failed to create session"
// @Router /auth/verify-login-otp [post]
func (ctrl *AuthController) VerifyLoginOTP(ctx *gin.Context) {
	var req schemas.VerifyLoginOTPRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Yêu cầu không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Validate user input
	if err := ctrl.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response := helper.ErrorResponseWithMessage(
			validationErrors,
			"Invalid input",
			"Dữ liệu đầu vào không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Verify the OTP
	err := ctrl.OTPService.VerifyOTP(ctx.Request.Context(), req.PhoneNumber, req.OTP)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"OTP verification failed",
			"Xác minh OTP thất bại",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Create a new session
	user, accessToken, refreshToken, err := ctrl.UserService.CreateSession(req.PhoneNumber, req.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create session",
			"Không thể tạo phiên xác thực người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	res := schemas.VerifyLoginOTPResponse{
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

	response := helper.SuccessResponse(res, "OTP verified successfully", "OTP đã được xác minh thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// Login with OAuth2
// LoginWithOAuth godoc
// @Summary Login with OAuth2
// @Description Authenticates a user using OAuth2 and sends an OTP to their phone number
// @Tags auth
// @Accept json
// @Produce json
// @Param request body schemas.LoginWithOAuthRequest true "OAuth2 login details"
// @Success 200 {object} helper.Response{data=schemas.LoginWithOAuthResponse} "OTP sent successfully"
// @Failure 400 {object} helper.Response "Invalid request body"
// @Failure 404 {object} helper.Response "User does not exist"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /auth/login-oauth [post]
func (ctrl *AuthController) LoginWithOAuth(ctx *gin.Context) {
	var req schemas.LoginWithOAuthRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Yêu cầu không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Validate user input
	if err := ctrl.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response := helper.ErrorResponseWithMessage(
			validationErrors,
			"Invalid input",
			"Dữ liệu đầu vào không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Check if user exists
	exists, err := ctrl.UserService.UserExistsByEmail(req.Email)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to check user existence",
			"Không thể kiểm tra sự tồn tại của người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}
	if !exists {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("User does not exist"),
			"User does not exist",
			"Người dùng không tồn tại",
		)
		helper.GinResponse(ctx, http.StatusNotFound, response)
		return
	}

	// Get user info
	user, err := ctrl.UserService.GetUserByEmail(req.Email)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get user info",
			"Không thể lấy thông tin người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Send OTP via Twilio
	_, err = ctrl.OTPService.SendOTP(ctx.Request.Context(), user.PhoneNumber)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to send OTP",
			"Không thể gửi mã OTP",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Return user info
	res := schemas.LoginWithOAuthResponse{
		PhoneNumber: user.PhoneNumber,
		UserID:      user.ID,
		FullName:    user.FullName,
		IsActivated: user.IsActivated,
		IsVerified:  user.IsVerified,
	}

	response := helper.SuccessResponse(res, "OTP sent successfully", "Mã OTP đã được gửi thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// Refresh token and return new access token for the user
// RefreshToken godoc
// @Summary Refresh token and return new access token for the user
// @Description Validates the refresh token and issues a new access token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.Response{data=schemas.RefreshTokenResponse} "Access token refreshed successfully"
// @Failure 400 {object} helper.Response "Invalid refresh token or authorization header"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /auth/refresh-token [post]
func (ctrl *AuthController) RefreshToken(ctx *gin.Context) {
	// Get the refresh token from the request header
	authHeader := ctx.GetHeader("Authorization")

	if authHeader == "" {
		errorToken := fmt.Errorf("authorization header is missing")
		response := helper.ErrorResponseWithMessage(
			errorToken,
			"Authorization header is required",
			"Không thể tìm thấy header xác thực",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Split the token string to get the actual token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		errorToken := fmt.Errorf("invalid authorization header format")
		response := helper.ErrorResponseWithMessage(
			errorToken,
			"Invalid Authorization header format",
			"Định dạng header xác thực không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	refreshToken := parts[1]
	if refreshToken == "" {
		errorToken := fmt.Errorf("refresh token is empty")
		response := helper.ErrorResponseWithMessage(
			errorToken,
			"Refresh token is empty",
			"Refresh token trống",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Validate the refresh token
	claims, err := ctrl.UserService.ValidateRefreshToken(refreshToken)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid refresh token",
			"Refresh token không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}
	// Get user id from db
	// If the refresh token is valid, create a new access token
	accessToken, err := ctrl.UserService.RefreshNewToken(claims.PhoneNumber, claims.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create new access token",
			"Không thể tạo phiên xác thực mới",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Update session
	err = ctrl.UserService.UpdateSession(accessToken, claims.UserID, refreshToken)
	// If error mean token has been revoked
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to update session",
			"Không thể cập nhật phiên xác thực",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	res := schemas.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       claims.UserID}

	response := helper.SuccessResponse(res, "Access token refreshed successfully", "Phiên xác thực đã được cập nhật")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// Logut user and revoke the token from the database
// @Summary Logout user and revoke the token
// @Description Logs out the user by revoking their refresh token from the database
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.Response "Logout successful"
// @Failure 400 {object} helper.Response "Bad request"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /auth/logout [post]
func (ctrl *AuthController) Logout(ctx *gin.Context) {
	// Get the refresh token from the request header
	authHeader := ctx.GetHeader("Authorization")

	if authHeader == "" {
		errorToken := fmt.Errorf("authorization header is missing")
		response := helper.ErrorResponseWithMessage(
			errorToken,
			"Authorization header is required",
			"Không thể tìm thấy header xác thực",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Split the token string to get the actual token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		errorToken := fmt.Errorf("invalid authorization header format")
		response := helper.ErrorResponseWithMessage(
			errorToken,
			"Invalid Authorization header format",
			"Định dạng header xác thực không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	refreshToken := parts[1]
	if refreshToken == "" {
		errorToken := fmt.Errorf("refresh token is empty")
		response := helper.ErrorResponseWithMessage(
			errorToken,
			"Refresh token is empty",
			"Refresh token trống",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Validate the refresh token
	claims, err := ctrl.UserService.ValidateRefreshToken(refreshToken)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid refresh token",
			"Refresh token không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Revoke the token
	err = ctrl.UserService.RevokeToken(claims.UserID, refreshToken)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to revoke token",
			"Không thể thu hồi token",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	response := helper.SuccessResponse(nil, "Logout successful", "Đăng xuất thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}

// DeleteUser delete the user from provide phone number in db (only available in dev environment)
// DeleteUser godoc
// @Summary Delete a user
// @Description Delete the user from the provided phone number in the database (only available in dev environment)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body schemas.DeleteUserRequest true "Delete user request"
// @Success 200 {object} helper.Response "User deleted successfully"
// @Failure 400 {object} helper.Response "Invalid request body or input"
// @Failure 500 {object} helper.Response "Failed to delete user"
// @Router /auth/delete-user [post]
func (ctrl *AuthController) DeleteUser(ctx *gin.Context) {
	var req schemas.DeleteUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request body",
			"Yêu cầu không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	if err := ctrl.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response := helper.ErrorResponseWithMessage(
			validationErrors,
			"Invalid input",
			"Dữ liệu đầu vào không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Delete user
	err := ctrl.UserService.DeleteUser(req.PhoneNumber)

	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to delete user",
			"Không thể xóa người dùng",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	response := helper.SuccessResponse(nil, "User deleted successfully", "Người dùng đã được xóa thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}
