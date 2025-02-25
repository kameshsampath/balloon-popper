package routes

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kameshsampath/balloon-popper-server/pkg/models"
	"github.com/kameshsampath/balloon-popper-server/pkg/producer"
	"github.com/kameshsampath/balloon-popper-server/pkg/security"
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
		upgrader: upgrader,
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
