package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/kameshsampath/balloon-popper-server/models"
	"github.com/kameshsampath/balloon-popper-server/producer"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	echo          *echo.Echo
	gameState     *models.GameState
	config        *models.GameConfig
	kafkaProducer *producer.KafkaScoreProducer
	upgrader      websocket.Upgrader
	mu            sync.RWMutex // For thread-safe gameState access
}

type GameMessage struct {
	Player       string `json:"player"`
	Character    string `json:"character"`
	BalloonColor string `json:"balloon_color"`
}

type ScoreUpdate struct {
	Type  string            `json:"type"`
	Event *models.GameEvent `json:"event"`
}

func NewServer() *Server {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize Echo
	e := echo.New()

	// Configure middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Create static directory if it doesn't exist
	staticDir := filepath.Join(".", "static")

	// Serve static files
	e.Static("/static", staticDir)

	// Initialize WebSocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins
		},
	}

	// Initialize Kafka kafkaScoreProducer
	kafkaBootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	if kafkaBootstrapServers == "" {
		kafkaBootstrapServers = "localhost:9092"
	}
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "game-scores"
	}

	kafkaScoreProducer, err := producer.NewKafkaScoreProducer(kafkaBootstrapServers, kafkaTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka kafkaScoreProducer: %v", err)
	}

	return &Server{
		echo:          e,
		gameState:     models.NewGameState(),
		config:        models.NewGameConfig(),
		kafkaProducer: kafkaScoreProducer,
		upgrader:      upgrader,
	}
}

func (s *Server) Start() error {
	// Start Kafka producer
	if err := s.kafkaProducer.Start(); err != nil {
		return fmt.Errorf("failed to start Kafka producer: %v", err)
	}

	// Define routes
	s.echo.GET("/", s.handleRoot)
	s.echo.GET("/index.html", s.handleRoot)
	s.echo.GET("/config", s.handleGetConfig)
	s.echo.POST("/game/start", s.handleStartGame)
	s.echo.POST("/game/stop", s.handleStopGame)
	s.echo.GET("/game/status", s.handleGameStatus)
	s.echo.GET("/ws/:player", s.handleWebSocket)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return s.echo.Start(":" + port)
}

func (s *Server) Stop() error {
	if err := s.kafkaProducer.Stop(); err != nil {
		return fmt.Errorf("failed to stop Kafka producer: %v", err)
	}
	return s.echo.Close()
}

func (s *Server) handleRoot(c echo.Context) error {
	return c.File(filepath.Join("static", "index.html"))
}

func (s *Server) handleGetConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, s.config)
}

// TODO: add auth and protect
func (s *Server) handleStartGame(c echo.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.gameState.IsActive {
		return echo.NewHTTPError(http.StatusBadRequest, "Game is already in progress")
	}

	now := time.Now().UTC()
	s.gameState.IsActive = true
	s.gameState.StartedAt = now
	s.gameState.EndedAt = time.Time{}
	s.gameState.CurrentPlayers = make([]string, 0)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "Game started",
		"started_at": s.gameState.StartedAt,
	})
}

// TODO: add auth and protect
func (s *Server) handleStopGame(c echo.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.gameState.IsActive {
		return echo.NewHTTPError(http.StatusBadRequest, "No game in progress")
	}

	now := time.Now().UTC()
	s.gameState.IsActive = false
	s.gameState.EndedAt = now

	sessionStats := map[string]interface{}{
		"started_at":       s.gameState.StartedAt,
		"ended_at":         s.gameState.EndedAt,
		"duration_seconds": s.gameState.EndedAt.Sub(s.gameState.StartedAt).Seconds(),
		"total_players":    len(s.gameState.CurrentPlayers),
		"player_list":      s.gameState.CurrentPlayers,
	}

	s.gameState.CurrentPlayers = make([]string, 0)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Game stopped",
		"session_stats": sessionStats,
	})
}

func (s *Server) handleGameStatus(c echo.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"is_active":       s.gameState.IsActive,
		"started_at":      s.gameState.StartedAt,
		"ended_at":        s.gameState.EndedAt,
		"current_players": s.gameState.CurrentPlayers,
		"player_count":    len(s.gameState.CurrentPlayers),
	})
}

func (s *Server) handleWebSocket(c echo.Context) error {
	s.mu.Lock()
	if !s.gameState.IsActive {
		s.mu.Unlock()
		return echo.NewHTTPError(http.StatusForbidden, "No active game session")
	}
	s.mu.Unlock()

	playerName := c.Param("player")
	ws, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %v", err)
	}
	defer ws.Close()

	// Add player to current session
	s.mu.Lock()
	if !contains(s.gameState.CurrentPlayers, playerName) {
		s.gameState.CurrentPlayers = append(s.gameState.CurrentPlayers, playerName)
	}
	s.mu.Unlock()

	// Remove player when done
	defer func() {
		s.mu.Lock()
		s.gameState.CurrentPlayers = removeString(s.gameState.CurrentPlayers, playerName)
		s.mu.Unlock()
		log.Printf("Player %s disconnected", playerName)
	}()

	for {
		// Check if game is still active
		s.mu.RLock()
		isActive := s.gameState.IsActive
		s.mu.RUnlock()

		if !isActive {
			return nil
		}

		// Read message
		var msg GameMessage
		if err := ws.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return nil
		}

		c.Logger().Printf("Recevied message %s", msg)

		// Process game event
		isFavoriteHit := contains(s.config.CharacterFavorites[msg.Character], msg.BalloonColor)
		score := s.config.Colors[msg.BalloonColor]
		//double the score for bonus hits
		if isFavoriteHit {
			score = score * 2
		}
		event := models.NewGameEvent(
			playerName,
			msg.BalloonColor,
			score,
			isFavoriteHit,
		)

		// Send to Kafka with context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := s.kafkaProducer.SendScore(ctx, event); err != nil {
			log.Printf("Failed to send score to Kafka: %v", err)
		}
		cancel()

		// Send score update
		update := ScoreUpdate{
			Type:  "score_update",
			Event: event,
		}
		if err := ws.WriteJSON(update); err != nil {
			log.Printf("Failed to send score update: %v", err)
			return nil
		}
	}
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeString(slice []string, item string) []string {
	for i, s := range slice {
		if s == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
