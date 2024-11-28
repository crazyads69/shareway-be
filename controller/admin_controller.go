package controller

import (
	"fmt"
	"shareway/helper"
	"shareway/middleware"
	"shareway/schemas"
	"shareway/service"
	"shareway/util"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AdminController handles authentication-related requests
type AdminController struct {
	cfg          util.Config
	validate     *validator.Validate
	AdminService service.IAdminService
}

// NewAdminController creates a new AdminController instance
func NewAdminController(cfg util.Config, validate *validator.Validate, adminService service.IAdminService) *AdminController {
	return &AdminController{
		cfg:          cfg,
		validate:     validate,
		AdminService: adminService,
	}
}

// GetAdminProfile returns the profile of the admin
// @Summary Get the profile of the admin
// @Description Get the profile of the admin
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} schemas.GetAdminProfileResponse
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /admin/get-profile [get]
func (ac *AdminController) GetAdminProfile(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert the payload to a map of string and interface
	// Convert payload to map
	data, err := helper.ConvertToAdminPayload(payload)

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

	// Get the admin ID from the payload and use it to get the admin profile
	admin, err := ac.AdminService.GetAdminProfile(data.AdminID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get admin profile",
			"Không thể lấy thông tin admin",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// Return the admin profile
	res := schemas.GetAdminProfileResponse{
		AdminInfo: schemas.AdminInfo{
			ID:        admin.ID,
			FullName:  admin.FullName,
			CreatedAt: admin.CreatedAt,
			UpdatedAt: admin.UpdatedAt,
			Username:  admin.Username,
			Role:      admin.Role,
		},
	}

	response := helper.SuccessResponse(res, "Admin profile retrieved successfully", "Lấy thông tin admin thành công")
	helper.GinResponse(ctx, 200, response)
}
