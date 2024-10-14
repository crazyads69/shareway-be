package controller

import (
	"fmt"
	"shareway/helper"
	"shareway/middleware"
	"shareway/util"

	"github.com/gin-gonic/gin"
)

// ProtectedController handles protected route operations
type ProtectedController struct {
	cfg util.Config
}

// NewProtectedController creates a new instance of ProtectedController
func NewProtectedController(cfg util.Config) *ProtectedController {
	return &ProtectedController{
		cfg: cfg,
	}
}

// ProtectedEndpoint godoc
// @Summary Test protected endpoint
// @Description This endpoint tests the authentication middleware
// @Tags Protected
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Successfully authenticated"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /protected/test [get]
func (ctrl *ProtectedController) ProtectedEndpoint(ctx *gin.Context) {
	payload := ctx.MustGet(middleware.AuthorizationPayloadKey)

	data, err := helper.ConvertToPayload(payload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	response := helper.SuccessResponse(data, "Successfully authenticated", "Xác thực thành công")
	helper.GinResponse(ctx, 200, response)
}
