package routes

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kameshsampath/balloon-popper-server/pkg/models"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

func (e *EndpointConfig) Root(c echo.Context) error {
	return c.File(filepath.Join("static", "index.html"))
}

func (e *EndpointConfig) GetConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, e.config)
}

func (e *EndpointConfig) GameStatus(c echo.Context) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	gameStatus := models.NewGameState()
	gameStatus.IsActive = e.gameState.IsActive
	gameStatus.StartedAt = e.gameState.StartedAt
	gameStatus.EndedAt = e.gameState.EndedAt
	gameStatus.CurrentPlayers = e.gameState.CurrentPlayers
	gameStatus.PlayerCount = len(e.gameState.CurrentPlayers)

	return c.JSON(http.StatusOK, gameStatus)
}

func (e *EndpointConfig) WebSocket(c echo.Context) error {
	e.mu.Lock()
	if !e.gameState.IsActive {
		e.mu.Unlock()
		return echo.NewHTTPError(http.StatusForbidden, "No active game session")
	}
	e.mu.Unlock()

	playerName := c.Param("player")
	ws, err := e.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %v", err)
	}
	defer ws.Close() //nolint:errcheck

	// Add player to current session
	e.mu.Lock()
	if !contains(e.gameState.CurrentPlayers, playerName) {
		e.gameState.CurrentPlayers = append(e.gameState.CurrentPlayers, playerName)
	}
	e.mu.Unlock()

	// Remove player when done
	defer func() {
		e.mu.Lock()
		e.gameState.CurrentPlayers = removeString(e.gameState.CurrentPlayers, playerName)
		e.mu.Unlock()
		log.Printf("Player %s disconnected", playerName)
	}()

	for {
		// Check if game is still active
		e.mu.RLock()
		isActive := e.gameState.IsActive
		e.mu.RUnlock()

		if !isActive {
			return nil
		}

		// Read message
		var msg models.GameMessage
		if err := ws.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return nil
		}

		c.Logger().Printf("Recevied message %s", msg)

		// Process game event
		isFavoriteHit := contains(e.config.CharacterFavorites[msg.Character], msg.BalloonColor)
		score := e.config.Colors[msg.BalloonColor]
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
		if err := e.KafkaProducer.SendScore(ctx, event); err != nil {
			log.Printf("Failed to send score to Kafka: %v", err)
		}
		cancel()

		// Send score update
		update := models.ScoreUpdate{
			Type:  "score_update",
			Event: event,
		}
		if err := ws.WriteJSON(update); err != nil {
			log.Printf("Failed to send score update: %v", err)
			return nil
		}
	}
}
