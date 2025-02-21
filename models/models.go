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
