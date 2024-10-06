package controller

import (
	"fmt"
	"shareway/helper"
	"shareway/middleware"
	"shareway/util"
	"shareway/util/token"

	"github.com/gin-gonic/gin"
)

// ProtectedController handles protected route operations
type ProtectedController struct {
	maker *token.PasetoMaker
	cfg   util.Config
}

// NewProtectedController creates a new instance of ProtectedController
func NewProtectedController(maker *token.PasetoMaker, cfg util.Config) *ProtectedController {
	return &ProtectedController{
		maker: maker,
		cfg:   cfg,
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
	if !err {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
	}

	response := helper.SuccessResponse(data, "Successfully authenticated", "Xác thực thành công")
	helper.GinResponse(ctx, 200, response)
}
