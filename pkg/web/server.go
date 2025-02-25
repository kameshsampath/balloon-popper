package web

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/kameshsampath/balloon-popper-server/pkg/routes"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"path/filepath"
	"strconv"
	"time"
)

// Server represents the HTTP server configuration and dependencies
type Server struct {
	endPointsConfig *routes.EndpointConfig
	echo            *echo.Echo
	port            int
}

// ServerBuilder is a builder for Server
type ServerBuilder struct {
	server *Server
}

// NewServer creates a new ServerBuilder with required dependencies
func NewServer(logger *zap.SugaredLogger, port int, ec *routes.EndpointConfig) *Server {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Warnf("Warning: .env file not found: %v", err)
	}

	// Initialize Echo
	e := echo.New()

	// Configure middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(zapLoggerMiddleware(logger))

	// Create static directory if it doesn't exist
	staticDir := filepath.Join(".", "static")

	// Serve static files
	e.Static("/static", staticDir)

	return &Server{
		echo:            e,
		port:            port,
		endPointsConfig: ec,
	}
}

// Build builds and returns the final Server
func (b *ServerBuilder) Build() *Server {
	return b.server
}

func (s *Server) Start() error {
	ec := s.endPointsConfig
	// Configure Routes
	router := s.echo
	// Login endpoint
	router.POST("/login", ec.Login)
	//Health Endpoints accessible via /health
	health := router.Group("/health")
	{
		health.GET("/live", ec.Live)
		health.GET("/ready", ec.Ready)
	}
	//Game API endpoints /game
	game := router.Group("/game")
	{
		game.GET("/", ec.Root)
		game.GET("/index.html", ec.Root)
		game.GET("/config", ec.GetConfig)
		game.GET("/status", ec.GameStatus)
	}
	//WebSockets
	ws := router.Group("/ws")
	{
		ws.GET("/:player", ec.WebSocket)
	}
	//Protected Game Admin endpoints /admin
	admin := router.Group("/admin")
	{
		// Configure middleware with the custom claims type
		config := echojwt.Config{
			KeyFunc: func(token *jwt.Token) (interface{}, error) {
				return ec.Manager.Config.PublicKey, nil
			},
		}
		admin.Use(echojwt.WithConfig(config))
		admin.POST("/start", ec.StartGame)
		admin.POST("/stop", ec.StopGame)
	}
	// Start server
	port := strconv.Itoa(s.port)
	if port == "" {
		port = "8080"
	}
	return router.Start(":" + port)
}

func (s *Server) Stop() error {
	if err := s.endPointsConfig.KafkaProducer.Stop(); err != nil {
		return fmt.Errorf("failed to stop Kafka producer: %v", err)
	}
	return s.echo.Close()
}

func zapLoggerMiddleware(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Add request ID to the logger context if available
			reqID := c.Request().Header.Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = c.Response().Header().Get(echo.HeaderXRequestID)
			}

			// Process request
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			// Request completion time
			stop := time.Now()

			// Get response status
			status := c.Response().Status

			// Log request details
			logger.Infow("Request completed",
				"id", reqID,
				"method", c.Request().Method,
				"uri", c.Request().RequestURI,
				"status", status,
				"duration", stop.Sub(start).String(),
				"ip", c.RealIP(),
				"user_agent", c.Request().UserAgent(),
			)

			return err
		}
	}
}
