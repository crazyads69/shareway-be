package controller

import (
	"fmt"
	"net/http"

	"shareway/helper"
	"shareway/schemas"
	"shareway/service"
	"shareway/util"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AuthController handles authentication-related requests
type AuthAdminController struct {
	cfg          util.Config
	validate     *validator.Validate
	AdminService service.IAdminService
}

// NewAuthController creates a new AuthController instance
func NewAuthAdminController(cfg util.Config, validate *validator.Validate, adminService service.IAdminService) *AuthAdminController {
	return &AuthAdminController{
		cfg:          cfg,
		validate:     validate,
		AdminService: adminService,
	}
}

// Login handles the login request for admin users
// @Summary Login as an admin
// @Description Authenticates an admin user and returns a token along with user information
// @Tags admin/auth
// @Accept json
// @Produce json
// @Param request body schemas.LoginAdminRequest true "The login request"
// @Success 200 {object} helper.Response{data=schemas.LoginAdminResponse} "Login successful"
// @Failure 400 {object} helper.Response "Invalid request body or input validation failed"
// @Failure 401 {object} helper.Response "Invalid credentials"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /admin/auth/login [post]
func (a *AuthAdminController) Login(ctx *gin.Context) {
	var req schemas.LoginAdminRequest
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
	if err := a.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response := helper.ErrorResponseWithMessage(
			validationErrors,
			"Invalid input",
			"Dữ liệu đầu vào không hợp lệ",
		)
		helper.GinResponse(ctx, http.StatusBadRequest, response)
		return
	}

	// Check if the user exists
	admin, err := a.AdminService.CheckAdminExists(req)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid credentials",
			"Tài khoản hoặc mật khẩu không đúng",
		)
		helper.GinResponse(ctx, http.StatusUnauthorized, response)
		return
	}

	// Verify the password
	if !a.AdminService.VerifyPassword(req.Password, admin.Password) {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("Invalid credentials"),
			"Invalid credentials",
			"Tài khoản hoặc mật khẩu không đúng",
		)
		helper.GinResponse(ctx, http.StatusUnauthorized, response)
		return
	}

	// Create a new token
	token, err := a.AdminService.CreateToken(admin)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to create token",
			"Không thể tạo token",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Return the response
	res := schemas.LoginAdminResponse{
		Token: token,
		AdminInfo: schemas.AdminInfo{
			ID:        admin.ID,
			CreatedAt: admin.CreatedAt,
			UpdatedAt: admin.UpdatedAt,
			Username:  admin.Username,
			FullName:  admin.FullName,
			Role:      admin.Role,
		},
	}

	response := helper.SuccessResponse(res, "Login successful", "Đăng nhập thành công")
	helper.GinResponse(ctx, http.StatusOK, response)
}
