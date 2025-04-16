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
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kameshsampath/balloon-popper/pkg/logger"
	"github.com/kameshsampath/balloon-popper/pkg/models"
	"github.com/kameshsampath/balloon-popper/pkg/security"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

// Login handles user login for admin operations
func (e *EndpointConfig) Login(c echo.Context) error {
	var creds *security.UserCredentials
	var err error
	log := logger.Get()

	username := c.FormValue("username")
	password := c.FormValue("password")

	// if the secret is not set on the endpoint configure use the default pattern
	// this also help testing this method with arbitrary secret name
	if e.UserSecretName == "" {
		e.UserSecretName = fmt.Sprintf("kameshs-bgd-user-%s", username)
	}

	// Load credentials form secret manager
	creds, err = security.LoadCredentials(e.UserSecretName)
	if err != nil {
		log.Errorf("Failed to load user credentials: %v", err)
		return echo.ErrUnauthorized
	}
	// Verify Login
	if err = creds.VerifyLogin(username, password); err != nil {
		return echo.ErrUnauthorized
	}
	//build the claims
	claims := &security.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
		},
		Name:     creds.Name,
		Username: creds.Username,
		Email:    creds.Email,
		Role:     creds.Role,
	}
	//generate the token
	t, err := e.Manager.GenerateToken(claims)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}

func (e *EndpointConfig) StartGame(c echo.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.gameState.IsActive {
		return echo.NewHTTPError(http.StatusBadRequest, "Game is already in progress")
	}

	now := time.Now().UTC()
	e.gameState.IsActive = true
	e.gameState.StartedAt = now
	e.gameState.EndedAt = time.Time{}
	e.gameState.CurrentPlayers = make([]string, 0)

	gameStatus := models.GameStatus{
		Message: "Game started",
		SessionStats: models.SessionStats{
			StartedAt: e.gameState.StartedAt,
		},
	}

	return c.JSON(http.StatusOK, gameStatus)
}

func (e *EndpointConfig) StopGame(c echo.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.gameState.IsActive {
		return echo.NewHTTPError(http.StatusBadRequest, "No game in progress")
	}

	now := time.Now().UTC()
	e.gameState.IsActive = false
	e.gameState.EndedAt = now

	e.gameState.CurrentPlayers = make([]string, 0)

	gameStatus := models.GameStatus{
		Message: "Game stopped",
		SessionStats: models.SessionStats{
			StartedAt:       e.gameState.StartedAt,
			EndedAt:         e.gameState.EndedAt,
			DurationSeconds: e.gameState.EndedAt.Sub(e.gameState.StartedAt).Seconds(),
			TotalPlayers:    len(e.gameState.CurrentPlayers),
			PlayerList:      e.gameState.CurrentPlayers,
		},
	}

	return c.JSON(http.StatusOK, gameStatus)
}
