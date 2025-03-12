/*
 * Copyright (c) 2025.  Kamesh Sampath <kamesh.sampath@hotmail.com>
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 *
 */

package routes

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kameshsampath/balloon-popper/pkg/models"
	"github.com/labstack/echo/v4"
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
	log := e.Logger
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
		log.Infof("Player %s disconnected", playerName)
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
				log.Infof("WebSocket error: %v", err)
			}
			return nil
		}

		log.Infof("Recevied message %s", msg)

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
			log.Infof("Failed to send score to Kafka: %v", err)
		}
		cancel()

		// Send score update
		update := models.ScoreUpdate{
			Type:  "score_update",
			Event: event,
		}
		if err := ws.WriteJSON(update); err != nil {
			log.Infof("Failed to send score update: %v", err)
			return nil
		}
	}
}
