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
	"github.com/gorilla/websocket"
	"github.com/kameshsampath/balloon-popper/pkg/models"
	"github.com/kameshsampath/balloon-popper/pkg/producer"
	"github.com/kameshsampath/balloon-popper/pkg/security"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

// EndpointConfig is the marker interface for defining routes
type EndpointConfig struct {
	Manager       *security.JWTManager
	mu            sync.RWMutex // For thread-safe gameState access
	gameState     *models.GameState
	config        *models.GameConfig
	KafkaProducer *producer.KafkaScoreProducer
	upgrader      websocket.Upgrader
	Users         []models.UserCredentials
	Logger        *zap.SugaredLogger
}

// NewEndpoints gives handle to REST EndpointConfig
func NewEndpoints(privateKeyFile string, passphrase string) (*EndpointConfig, error) {
	kdc, err := security.NewRSAKeyDecryptor(privateKeyFile)
	if err != nil {
		return nil, err
	}
	if kdc.IsEncrypted() && passphrase == "" {
		return nil, fmt.Errorf("error: passphrase is required to encrypt the private key")
	}
	kdc.KeyInfo.SetPassPhrase(passphrase)
	err = kdc.Decrypt()
	if err != nil {
		return nil, err
	}
	// Initialize WebSocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins
		},
	}
	//build the JWT Config
	jwtConfig := security.JWTConfig{
		PrivateKey: kdc.KeyInfo.PrivateKey(),
		PublicKey:  kdc.KeyInfo.PublicKey(),
		ExpiryTime: 1 * time.Hour,
		Issuer:     "BalloonPopper",
	}

	return &EndpointConfig{
		Manager: &security.JWTManager{
			Config: jwtConfig,
		},
		config:    models.NewGameConfig(),
		gameState: models.NewGameState(),
		upgrader:  upgrader,
	}, nil
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
