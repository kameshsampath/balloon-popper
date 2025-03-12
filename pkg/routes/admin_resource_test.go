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
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kameshsampath/balloon-popper/pkg/security"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var (
	cwd, _                 = os.Getwd()
	testCredentialsFile    = filepath.Clean(filepath.Join(cwd, "test_users.json"))
	testPrivateKeyFile     = filepath.Clean(filepath.Join(cwd, "jwt-private-key"))
	testPrivateKeyPassFile = filepath.Join(cwd, ".pass")
)

func TestLogin(t *testing.T) {
	users, err := security.LoadCredentials(testCredentialsFile)
	assert.NoError(t, err)
	d, err := security.NewRSAKeyDecryptor(testPrivateKeyFile)
	assert.NoError(t, err)
	data, err := os.ReadFile(filepath.Clean(testPrivateKeyPassFile))
	assert.NoError(t, err)
	d.KeyInfo.SetPassPhrase(string(data))
	assert.NoError(t, err)
	assert.NotNil(t, d)
	err = d.Decrypt()
	assert.NoError(t, err)
	e := echo.New()
	// Create form data
	form := url.Values{}
	form.Add("username", "admin")
	form.Add("password", "sup3rSecret!")
	//Create request
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	//Create response recorder
	rec := httptest.NewRecorder()
	//Endpoint Config
	ec := EndpointConfig{
		Users: users,
		Manager: &security.JWTManager{
			Config: security.JWTConfig{
				PrivateKey: d.KeyInfo.PrivateKey(),
				PublicKey:  d.KeyInfo.PublicKey(),
			},
		},
	}
	//Fire the request
	if c := e.NewContext(req, rec); assert.NoError(t, ec.Login(c)) {
		assert.Equal(t, 200, rec.Code)

		// Verify the token with public key
		var response map[string]string
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "token")

		token := response["token"]
		assert.NotEmpty(t, token)

		// Parse the token
		parsedToken, err := jwt.ParseWithClaims(token, &security.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method is correct
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return d.KeyInfo.PublicKey(), nil
		})
		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		// Verify the claims
		if claims, ok := parsedToken.Claims.(*security.JWTClaims); assert.True(t, ok) {
			assert.Equal(t, "admin", claims.Username)
			assert.Equal(t, "admin", claims.Name)
			assert.Equal(t, "balloon-game-admin@example.com", claims.Email)
			assert.Equal(t, "admin", claims.Role)
			assert.NotZero(t, claims.ExpiresAt)
			// Verify expiration is set correctly (about 1 hour from now)
			assert.WithinDuration(t, time.Now().Add(time.Hour), claims.ExpiresAt.Time, 5*time.Second)
		}
	}
}

func TestProtectedEndpoints(t *testing.T) {
	// Setup once
	e := echo.New()
	h := &EndpointConfig{}

	config := echojwt.Config{
		SigningKey:  []byte("your-secret-key"),
		TokenLookup: "header:Authorization",
	}
	jwtMiddleware := echojwt.WithConfig(config)

	// Register all protected routes
	e.POST("/admin/start", h.StartGame, jwtMiddleware)
	e.POST("/admin/stop", h.StopGame, jwtMiddleware)

	// Test cases
	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{"StartGame", http.MethodPost, "/admin/start"},
		{"StopGame", http.MethodPost, "/admin/stop"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request without token
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Authorization", "Bearer ")
			rec := httptest.NewRecorder()

			// Serve the request
			e.ServeHTTP(rec, req)

			// Assert 401 Unauthorized
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

//TODO start and stop game tests
