package controller

import (
	"fmt"
	"net/http"
	"shareride/middleware"

	"github.com/gin-gonic/gin"
)

type ProtectedController struct{}

//	@BasePath	/protected

// Protected Need header godoc
//
//	@Summary	test protected endpoint
//	@Schemes
//	@Description	test protected endpoint desc
//	@Tags			Protected branch
//	@Accept			json
//	@Produce		json
//	@Router			/protected/ [get]
func (ctrl *ProtectedController) Protected_endpoint(ctx *gin.Context) {
	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey)

	fmt.Println(authPayload)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}
