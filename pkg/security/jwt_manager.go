package security

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

var _ JWTTokenCommand = (*JWTManager)(nil)

// NewJWTManager creates a new instance of JWTManager
func NewJWTManager(config JWTConfig) *JWTManager {
	return &JWTManager{
		Config: config,
	}
}

func (m *JWTManager) GenerateToken(claims *JWTClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.Config.PrivateKey)
}

func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.Config.PublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
