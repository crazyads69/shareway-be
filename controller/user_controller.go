package controller

import (
	"fmt"
	"shareway/helper"
	"shareway/middleware"
	"shareway/schemas"
	"shareway/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserController struct {
	UserService service.IUsersService
	validate    *validator.Validate
}

func NewUserController(userService service.IUsersService, validate *validator.Validate) *UserController {
	return &UserController{
		UserService: userService,
		validate:    validate,
	}
}

// GetUserProfile receives access token and returns user profile information
// GetUserProfile retrieves and returns the user profile information based on the access token.
// @Summary Get user profile
// @Description Retrieves the profile information of the authenticated user
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} schemas.GetUserProfileResponse
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /user/get-profile [get]
func (ctrl *UserController) GetUserProfile(ctx *gin.Context) {

	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert payload to map
	data, err := helper.ConvertToPayload(payload)

	// If error occurs, return error response
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Get user information from payload (user_id) and return it
	user, err := ctrl.UserService.GetUserByID(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to get user information"),
			"Failed to get user information",
			"Không thể lấy thông tin người dùng",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.GetUserProfileResponse{
		User: schemas.UserResponse{
			ID:            user.ID,
			AvatarURL:     user.AvatarURL,
			Gender:        user.Gender,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
			PhoneNumber:   user.PhoneNumber,
			Email:         user.Email,
			FullName:      user.FullName,
			IsVerified:    user.IsVerified,
			IsMomoLinked:  user.IsMomoLinked,
			IsActivated:   user.IsActivated,
			Role:          user.Role,
			BalanceInApp:  user.BalanceInApp,
			AverageRating: user.AverageRating,
		},
	}

	response := helper.SuccessResponse(res, "Successfully authenticated", "Xác thực thành công")
	helper.GinResponse(ctx, 200, response)
}

// RegisterDeviceToken receives device token and saves it to the database for push notification service through Firebase Cloud Messaging
// RegisterDeviceToken godoc
// @Summary Register device token for push notifications
// @Description Registers the device token for the authenticated user to enable push notifications via Firebase Cloud Messaging
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.RegisterDeviceTokenRequest true "Device token registration request"
// @Success 200 {object} helper.Response "Successfully registered device token"
// @Failure 400 {object} helper.Response "Invalid request or validation error"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /user/register-device-token [post]
func (ctrl *UserController) RegisterDeviceToken(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert payload to map
	data, err := helper.ConvertToPayload(payload)

	// If error occurs, return error response
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Bind request to schema
	var req schemas.RegisterDeviceTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to bind request"),
			"Failed to bind request",
			"Không thể bind request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := ctrl.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Register device token
	err = ctrl.UserService.RegisterDeviceToken(data.UserID, req.DeviceToken)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to register device token"),
			"Failed to register device token",
			"Không thể đăng ký device token",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	response := helper.SuccessResponse(nil, "Successfully registered device token", "Đăng ký device token thành công")
	helper.GinResponse(ctx, 200, response)
}

// UpdateUserProfile receives user profile information and updates it in the database
// UpdateUserProfile godoc
// @Summary Update user profile
// @Description Update the profile information of the authenticated user
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body schemas.UpdateUserProfileRequest true "User profile update information"
// @Success 200 {object} helper.Response{data=schemas.UpdateUserProfileResponse} "Successfully updated user profile"
// @Failure 400 {object} helper.Response "Invalid input"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /user/update-profile [post]
func (ctrl *UserController) UpdateUserProfile(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert payload to map
	data, err := helper.ConvertToPayload(payload)

	// If error occurs, return error response
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Bind request to schema
	var req schemas.UpdateUserProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to bind request"),
			"Failed to bind request",
			"Không thể bind request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := ctrl.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Update user profile
	err = ctrl.UserService.UpdateUserProfile(data.UserID, req.FullName, req.Email, req.Gender)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to update user profile"),
			"Failed to update user profile",
			"Không thể cập nhật thông tin người dùng",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Retrieve updated user information and return it
	user, err := ctrl.UserService.GetUserByID(data.UserID)

	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get user information",
			"Không thể lấy thông tin người dùng")
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.UpdateUserProfileResponse{
		User: schemas.UserResponse{
			ID:            user.ID,
			Gender:        user.Gender,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
			AvatarURL:     user.AvatarURL,
			IsMomoLinked:  user.IsMomoLinked,
			PhoneNumber:   user.PhoneNumber,
			Email:         user.Email,
			FullName:      user.FullName,
			IsVerified:    user.IsVerified,
			IsActivated:   user.IsActivated,
			Role:          user.Role,
			BalanceInApp:  user.BalanceInApp,
			AverageRating: user.AverageRating,
		},
	}

	response := helper.SuccessResponse(res, "Successfully updated user profile", "Cập nhật thông tin người dùng thành công")
	helper.GinResponse(ctx, 200, response)
}

// UpdateAvatar receives avatar image and updates it in the database
// UpdateAvatar godoc
// @Summary Update user avatar
// @Description Update the avatar image of the authenticated user
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param avatar_image formData file true "Avatar image file"
// @Success 200 {object} helper.Response{data=schemas.UpdateAvatarResponse} "Successfully updated avatar"
// @Failure 400 {object} helper.Response "Invalid input"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /user/update-avatar [post]
func (ctrl *UserController) UpdateAvatar(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert payload to map
	data, err := helper.ConvertToPayload(payload)

	// If error occurs, return error response
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Bind request to schema
	var req schemas.UpdateAvatarRequest
	if err := ctx.ShouldBind(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to bind request"),
			"Failed to bind request",
			"Không thể bind request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := ctrl.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Update avatar
	avatarURL, err := ctrl.UserService.UpdateAvatar(ctx.Request.Context(), data.UserID, req.AvatarImage)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to update avatar"),
			"Failed to update avatar",
			"Không thể cập nhật ảnh đại diện",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Retrieve updated user information and return it
	user, err := ctrl.UserService.GetUserByID(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get user information",
			"Không thể lấy thông tin người dùng")
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.UpdateAvatarResponse{
		User: schemas.UserResponse{
			ID:            user.ID,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
			AvatarURL:     avatarURL,
			PhoneNumber:   user.PhoneNumber,
			Email:         user.Email,
			IsMomoLinked:  user.IsMomoLinked,
			FullName:      user.FullName,
			IsVerified:    user.IsVerified,
			IsActivated:   user.IsActivated,
			Role:          user.Role,
			Gender:        user.Gender,
			BalanceInApp:  user.BalanceInApp,
			AverageRating: user.AverageRating,
		}}

	response := helper.SuccessResponse(res, "Successfully updated avatar", "Cập nhật ảnh đại diện thành công")
	helper.GinResponse(ctx, 200, response)
}
