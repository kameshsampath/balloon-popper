package web

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/kameshsampath/balloon-popper-server/pkg/producer"
	"github.com/kameshsampath/balloon-popper-server/pkg/routes"
	"github.com/kameshsampath/balloon-popper-server/pkg/security"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Server allows configuring the echo server
type Server struct {
	endPointsConfig       *routes.EndpointConfig
	echo                  *echo.Echo
	Logger                *zap.SugaredLogger
	Port                  int
	JTWPrivateKeyFile     string
	PrivateKeyPassphrase  string
	KafkaBootstrapServers string
	KafkaTopic            string
	UserCredentialsFile   string
}

func NewServer(logger *zap.SugaredLogger) *Server {
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

	return &Server{}
}

func (s *Server) Start() error {
	logger := s.Logger

	//check to see if the private key file exists
	if _, err := os.Stat(s.JTWPrivateKeyFile); err != nil {
		logger.Fatalf("Error: error loading %s  private key file: %v", s.JTWPrivateKeyFile, err)
	}
	// create endpoint with JWT config
	ec, err := routes.NewEndpoints(s.JTWPrivateKeyFile, s.PrivateKeyPassphrase)
	if err != nil {
		logger.Fatalf("Failed to configure endpoints: %v", err)
	}

	//Load Users
	if c, err := security.LoadCredentials(s.UserCredentialsFile); err != nil {
		return err
	} else {
		ec.Users = c
	}
	// Initialize Kafka kafkaScoreProducer
	kp, err := producer.NewKafkaScoreProducer(s.KafkaBootstrapServers, s.KafkaTopic)
	if err != nil {
		logger.Fatalf("Failed to create Kafka kafkaScoreProducer: %v", err)
	}
	ec.KafkaProducer = kp
	// Start Kafka producer
	if err := ec.KafkaProducer.Start(); err != nil {
		return fmt.Errorf("failed to start Kafka producer: %v", err)
	}

	// Set ec on to the serv
	s.endPointsConfig = ec

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
				return s.endPointsConfig.Manager.Config.PublicKey, nil
			},
		}
		admin.Use(echojwt.WithConfig(config))
		admin.POST("/start", ec.StartGame)
		admin.POST("/stop", ec.StopGame)
	}
	// Start server
	port := strconv.Itoa(s.Port)
	if port == "" {
		port = "8080"
	}
	return s.echo.Start(":" + port)
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
