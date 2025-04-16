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

package models

import (
	"os"
	"strconv"
	"time"
)

// GameEvent represents a single game event
type GameEvent struct {
	Player             string    `json:"player"`
	BalloonColor       string    `json:"balloon_color"`
	Score              int       `json:"score"`
	FavoriteColorBonus bool      `json:"favorite_color_bonus"`
	EventTS            time.Time `json:"event_ts"`
}

// PlayerScore tracks the cumulative score for a player
type PlayerScore struct {
	Player      string    `json:"player"`
	TotalScore  int       `json:"total_score"`
	BonusHits   int       `json:"bonus_hits"`
	RegularHits int       `json:"regular_hits"`
	LastUpdated time.Time `json:"last_updated"`
}

// GameConfig holds the game configuration settings
type GameConfig struct {
	Colors             map[string]int      `json:"colors"`
	CharacterFavorites map[string][]string `json:"character_favorites"`
	BonusProbability   float64             `json:"bonus_probability"`
}

// GameState represents the current state of the game
type GameState struct {
	IsActive       bool      `json:"is_active"`
	StartedAt      time.Time `json:"started_at"`
	EndedAt        time.Time `json:"ended_at"`
	CurrentPlayers []string  `json:"current_players"`
	PlayerCount    int       `json:"player_count,omitempty"`
}

// GameMessage represents the message send with each balloon pops
type GameMessage struct {
	Player       string `json:"player"`
	Character    string `json:"character"`
	BalloonColor string `json:"balloon_color"`
}

// ScoreUpdate represents the score state
type ScoreUpdate struct {
	Type  string     `json:"type"`
	Event *GameEvent `json:"event"`
}

// GameStatus represents the game status started, stopped and active
type GameStatus struct {
	Message      string       `json:"message,omitempty"`
	SessionStats SessionStats `json:"session_stats,omitempty"`
}

// SessionStats provides the session stats
type SessionStats struct {
	StartedAt       time.Time `json:"started_at,omitempty"`
	EndedAt         time.Time `json:"ended_at,omitempty"`
	DurationSeconds float64   `json:"duration_seconds,omitempty"`
	TotalPlayers    int       `json:"total_players,omitempty"`
	PlayerList      []string  `json:"player_list,omitempty"`
}

// NewGameConfig creates a new GameConfig with default values
func NewGameConfig() *GameConfig {
	bonusProb, err := strconv.ParseFloat(os.Getenv("BONUS_PROBABILITY"), 64)
	if err != nil {
		bonusProb = 0.15 // default value
	}

	return &GameConfig{
		Colors: map[string]int{
			"red":    100,
			"blue":   75,
			"green":  60,
			"yellow": 50,
			"purple": 40,
			"brown":  35,
			"orange": 45,
			"pink":   55,
			"gold":   90,
			"grey":   70,
			"maroon": 80,
		},
		CharacterFavorites: map[string][]string{
			"Jerry":        {"brown", "yellow"},
			"Tom":          {"grey", "blue"},
			"Mickey":       {"red", "black"},
			"Donald":       {"blue", "maroon"},
			"Bugs_Bunny":   {"grey", "orange"},
			"Daffy_Duck":   {"black", "orange"},
			"SpongeBob":    {"yellow", "brown"},
			"Patrick":      {"pink", "green"},
			"Pikachu":      {"yellow", "red"},
			"Mario":        {"red", "blue"},
			"Sonic":        {"blue", "gold"},
			"Woody":        {"brown", "yellow"},
			"Buzz":         {"green", "purple"},
			"Scooby":       {"brown", "green"},
			"Popeye":       {"blue", "red"},
			"Pink_Panther": {"pink", "purple"},
			"Road_Runner":  {"blue", "purple"},
			"Tweety":       {"yellow", "orange"},
		},
		BonusProbability: bonusProb,
	}
}

// NewGameState creates a new GameState with initialized values
func NewGameState() *GameState {
	now := time.Now().UTC()
	return &GameState{
		IsActive:       false,
		StartedAt:      now,
		EndedAt:        now,
		CurrentPlayers: make([]string, 0),
	}
}

// NewGameEvent creates a new GameEvent with the current timestamp
func NewGameEvent(player, balloonColor string, score int, favoriteColorBonus bool) *GameEvent {
	return &GameEvent{
		Player:             player,
		BalloonColor:       balloonColor,
		Score:              score,
		FavoriteColorBonus: favoriteColorBonus,
		EventTS:            time.Now().UTC(),
	}
}
