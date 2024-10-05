package router

import (
	"shareway/controller"
	"shareway/util"
	"shareway/util/token"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupProtectedRouter configures the protected routes
func SetupProtectedRouter(group *gin.RouterGroup, maker *token.PasetoMaker, cfg util.Config, db *gorm.DB) {
	protectedController := controller.NewProtectedController(maker, cfg, db)
	group.GET("/test", protectedController.ProtectedEndpoint)
}
