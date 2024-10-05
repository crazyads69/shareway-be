package router

import (
	"shareway/middleware"
	"shareway/util"
	"shareway/util/token"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	docs "shareway/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// APIServer represents the API server structure
type APIServer struct {
	router *gin.Engine
	Maker  *token.PasetoMaker
	Cfg    util.Config
	DB     *gorm.DB
}

// NewAPIServer creates and initializes a new APIServer instance
func NewAPIServer(maker *token.PasetoMaker, cfg util.Config, db *gorm.DB) (*APIServer, error) {
	r := gin.Default()

	// Set up basic routes
	setupBasicRoutes(r)

	return &APIServer{
		router: r,
		Maker:  maker,
		Cfg:    cfg,
		DB:     db,
	}, nil
}

// setupBasicRoutes adds the root and health check routes to the router
func setupBasicRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"error":   false,
			"message": "Application is running",
		})
	})

	r.GET("/health_check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"error":   false,
			"message": "Server is running (Healthy)",
		})
	})
}

// Start begins the API server on the specified address
func (server *APIServer) Start(address string) error {
	return server.router.Run(address)
}

// SetupRouter configures the main routes for the API server
func (server *APIServer) SetupRouter() {
	SetupAuthRouter(server.router.Group("/auth"))
	SetupProtectedRouter(server.router.Group("/protected", middleware.AuthMiddleware(server.Maker)), server.Maker, server.Cfg, server.DB)
}

// SetupSwagger configures the Swagger documentation route
func (server *APIServer) SetupSwagger(swaggerURL string) {
	docs.SwaggerInfo.BasePath = "/"
	server.router.GET(swaggerURL+"/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
