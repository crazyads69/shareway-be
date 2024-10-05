package router

import (
	controller "shareway/controlller"

	"github.com/gin-gonic/gin"
)

// SetupProtectedRouter ...
//	@Summary	test protected endpoint
//	@Schemes
//	@Description				test protected endpoint desc
//	@Tags						Protected branch
//	@Accept						json
//	@Produce					json
//	@BasePath					/auth/login [post]
//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func SetupAuthRouter(group *gin.RouterGroup) {
	auth_controller := controller.AuthController{}
	group.POST("/login", auth_controller.Login)
	group.POST("/register", auth_controller.Login)
	group.POST("/refresh", auth_controller.Login)
}
