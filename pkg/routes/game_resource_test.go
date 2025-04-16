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
	"encoding/json"
	"github.com/kameshsampath/balloon-popper/pkg/logger"
	"github.com/kameshsampath/balloon-popper/pkg/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const defaultConfigJSON = `
{
  "colors": {
    "blue": 75,
    "brown": 35,
    "gold": 90,
    "green": 60,
    "grey": 70,
    "maroon": 80,
    "orange": 45,
    "pink": 55,
    "purple": 40,
    "red": 100,
    "yellow": 50
  },
  "character_favorites": {
    "Bugs_Bunny": ["grey", "orange"],
    "Buzz": ["green", "purple"],
    "Daffy_Duck": ["black", "orange"],
    "Donald": ["blue", "maroon"],
    "Jerry": ["brown", "yellow"],
    "Mario": ["red", "blue"],
    "Mickey": ["red", "black"],
    "Patrick": ["pink", "green"],
    "Pikachu": ["yellow", "red"],
    "Pink_Panther": ["pink", "purple"],
    "Popeye": ["blue", "red"],
    "Road_Runner": ["blue", "purple"],
    "Scooby": ["brown", "green"],
    "Sonic": ["blue", "gold"],
    "SpongeBob": ["yellow", "brown"],
    "Tom": ["grey", "blue"],
    "Tweety": ["yellow", "orange"],
    "Woody": ["brown", "yellow"]
  },
  "bonus_probability": 0.15
}
`

func TestGetConfig(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	rec := httptest.NewRecorder()

	_, _ = logger.NewLogger(logger.Config{
		Development: true,
		WithCaller:  true,
		Level:       "debug",
		Output:      os.Stdout,
	})
	ec := EndpointConfig{
		config:    models.NewGameConfig(),
		gameState: models.NewGameState(),
	}
	if c := e.NewContext(req, rec); assert.NoError(t, ec.GetConfig(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotNil(t, rec.Body)

		var want models.GameConfig
		err := json.Unmarshal([]byte(defaultConfigJSON), &want)
		assert.NoError(t, err, "Failed to parse expected JSON")

		var got models.GameConfig
		err = json.Unmarshal(rec.Body.Bytes(), &got)
		assert.NoError(t, err, "Failed to parse response JSON")

		assert.Equal(t, want, got, "Expecting %s, got %s", want, got)
	}
}

func TestGetGameStatus(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()

	tl, _ := logger.NewLogger(logger.Config{
		Development: true,
		WithCaller:  true,
		Level:       "debug",
		Output:      os.Stdout,
	})
	ec := EndpointConfig{
		config:    models.NewGameConfig(),
		gameState: models.NewGameState(),
	}
	if c := e.NewContext(req, rec); assert.NoError(t, ec.GameStatus(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotNil(t, rec.Body)

		tl.Infof(rec.Body.String())

		var want models.GameState
		err := json.Unmarshal([]byte(`{"is_active":false,"started_at":"2025-02-25T11:07:49.401055Z","ended_at":"2025-02-25T11:07:49.401055Z","current_players":[]}`), &want)
		assert.NoError(t, err, "Failed to parse expected JSON")

		var got models.GameState
		err = json.Unmarshal(rec.Body.Bytes(), &got)
		assert.NoError(t, err, "Failed to parse response JSON")

		assert.Equal(t, want.IsActive, got.IsActive, "IsActive field doesn't match")
		assert.Equal(t, want.CurrentPlayers, got.CurrentPlayers, "CurrentPlayers field doesn't match")
	}
}
