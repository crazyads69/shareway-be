package router

import (
	controller "golang_template/controlller"

	"github.com/gin-gonic/gin"
)

func SetupProtectedRouter(group *gin.RouterGroup) {
	protected_router := controller.ProtectedController{}
	group.GET("/protected", protected_router.Protected_endpoint)
}
