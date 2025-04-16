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
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
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
	Manager        *security.JWTManager
	mu             sync.RWMutex // For thread-safe gameState access
	gameState      *models.GameState
	config         *models.GameConfig
	KafkaProducer  *producer.KafkaScoreProducer
	upgrader       websocket.Upgrader
	UserSecretName string
	Logger         *zap.SugaredLogger
}

// NewEndpoints gives handle to REST EndpointConfig
func NewEndpoints(jwtKeysSecretName string) (*EndpointConfig, error) {

	client, err := security.InitAndGetAWSSecretManagerClient()
	if err != nil {
		return nil, err
	}

	//Get secret
	sv, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(jwtKeysSecretName),
	})

	if err != nil {
		return nil, err
	}
	str := *sv.SecretString
	var epk security.EncryptedKeyPair
	err = json.Unmarshal([]byte(str), &epk)

	kgc := security.Config{}
	kgc.KeyInfo.SetPassPhrase(epk.Passphrase)

	privKey, err := kgc.DecodePrivateKey(epk.EncryptedPrivateKey)
	if err != nil {
		return nil, err
	}
	pubKey, err := kgc.DecodePublicKey(epk.PublicKey)
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
		PrivateKey: privKey,
		PublicKey:  pubKey,
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
