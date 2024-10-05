package router

import (
	controller "shareway/controlller"

	"github.com/gin-gonic/gin"
)

func SetupProtectedRouter(group *gin.RouterGroup) {
	protected_router := controller.ProtectedController{}
	group.GET("/protected", protected_router.Protected_endpoint)
}
