package router

import (
	"shareway/infra/agora"
	"shareway/infra/task"
	"shareway/infra/ws"
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
	router      *gin.Engine
	Maker       *token.PasetoMaker
	Cfg         util.Config
	Service     *service.ServiceContainer
	Validate    *validator.Validate
	Hub         *ws.Hub
	AsyncClient *task.AsyncClient
	Agora       *agora.Agora
}

// NewAPIServer creates and initializes a new APIServer instance
func NewAPIServer(maker *token.PasetoMaker, cfg util.Config, service *service.ServiceContainer, Validate *validator.Validate, Hub *ws.Hub, AsyncClient *task.AsyncClient, Agora *agora.Agora) (*APIServer, error) {
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
		router:      r,
		Maker:       maker,
		Cfg:         cfg,
		Service:     service,
		Validate:    Validate,
		Hub:         Hub,
		AsyncClient: AsyncClient,
		Agora:       Agora,
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
	// Map routes for map management
	SetupMapRouter(server.router.Group("/map", middleware.AuthMiddleware(server.Maker)), server)
	// Vehicle routes for vehicle management
	SetupVehicleRouter(server.router.Group("/vehicle", middleware.AuthMiddleware(server.Maker)), server)
	// WebSocket route
	server.router.GET("/ws", server.HandleWebSocket)
	// Ride routes for ride matching and engagement
	SetupRideRouter(server.router.Group("/ride", middleware.AuthMiddleware(server.Maker)), server)
	// Notification routes for sending notifications
	SetupNotificationRouter(server.router.Group("/notification", middleware.AuthMiddleware(server.Maker)), server)
	// Chat routes for sending messages
	SetupChatRouter(server.router.Group("/chat", middleware.AuthMiddleware(server.Maker)), server)

}

// SetupSwagger configures the Swagger documentation route
func (server *APIServer) SetupSwagger(swaggerURL string) {
	docs.SwaggerInfo.BasePath = "/"
	server.router.GET(swaggerURL+"/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

// handleWebSocket handles WebSocket connections
func (server *APIServer) HandleWebSocket(ctx *gin.Context) {
	ws.WebSocketHandler(ctx, server.Hub)
}
