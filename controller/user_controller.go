package controller

import (
	"fmt"
	"shareway/helper"
	"shareway/middleware"
	"shareway/schemas"
	"shareway/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserService service.IUsersService
}

func NewUserController(userService service.IUsersService) *UserController {
	return &UserController{
		UserService: userService,
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
// @Param Authorization header string true "Bearer <access_token>"
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
	}

	response := helper.SuccessResponse(res, "Successfully authenticated", "Xác thực thành công")
	helper.GinResponse(ctx, 200, response)
}
