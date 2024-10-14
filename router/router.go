package router

import (
	"shareway/middleware"
	"shareway/service"
	"shareway/util"
	"shareway/util/token"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	docs "shareway/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// APIServer represents the API server structure
type APIServer struct {
	router   *gin.Engine
	Maker    *token.PasetoMaker
	Cfg      util.Config
	Service  *service.ServiceContainer
	Validate *validator.Validate
}

// NewAPIServer creates and initializes a new APIServer instance
func NewAPIServer(maker *token.PasetoMaker, cfg util.Config, service *service.ServiceContainer, Validate *validator.Validate) (*APIServer, error) {
	r := gin.Default()

	if cfg.GinMode != "release" {
		r.Use(cors.New(cors.Config{
			AllowAllOrigins:  true,
			AllowCredentials: true,
			AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Request-Id", "X-Requested-With"},
			MaxAge:           12 * time.Hour,
		}))
	}
	// Set up basic routes
	setupBasicRoutes(r)

	return &APIServer{
		router:   r,
		Maker:    maker,
		Cfg:      cfg,
		Service:  service,
		Validate: Validate,
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
	// Auth routes for user authentication
	SetupAuthRouter(server.router.Group("/auth"), server)
	SetupProtectedRouter(server.router.Group("/protected", middleware.AuthMiddleware(server.Maker)), server)
	// User routes for user management
	SetupUserRouter(server.router.Group("/user", middleware.AuthMiddleware(server.Maker)), server)
}

// SetupSwagger configures the Swagger documentation route
func (server *APIServer) SetupSwagger(swaggerURL string) {
	docs.SwaggerInfo.BasePath = "/"
	server.router.GET(swaggerURL+"/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
