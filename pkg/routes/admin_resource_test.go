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
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kameshsampath/balloon-popper/pkg/logger"
	"github.com/kameshsampath/balloon-popper/pkg/security"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const (
	testUser         = "testUser"
	testUserName     = "testUserName"
	testUserEmail    = "testuser@example.com"
	testPassword     = "sup3rSecret!"
	testPasswordHash = "$2a$10$kaH32ILCHvagRK.f3girQO29Ccyp7Lw40bMEnA8zhow4ke3alhTXq"
	testUserRole     = "admin"
)

var (
	client         *secretsmanager.Client
	err            error
	jwtSecretName  string
	userSecretName string
	kgc            *security.Config
)

func init() {
	suffix := time.Now().UnixMilli()
	jwtSecretName = fmt.Sprintf("kameshs-bgd-server-admin-test-%d", suffix)
	userSecretName = fmt.Sprintf("bgd-user-test-%d", suffix)
	client, err = security.InitAndGetAWSSecretManagerClient()
	if err != nil {
		panic(err)
	}
	//write the test credentials to AWS Secret Manager
	var u security.UserCredentials
	ustr := fmt.Sprintf(`{
"username": "%s",
"name": "%s",
"password_hash": "%s",
"email": "%s",
"role": "%s"
}`, testUser, testUserName, testPasswordHash, testUserEmail, testUserRole)
	err = json.Unmarshal([]byte(ustr), &u)
	if err != nil {
		panic(err)
	}

	err = u.WriteCredentials(userSecretName)
	if err != nil {
		panic(err)
	}
	//write test jwt-keys to AWS Secret Manager
	kgc, err = security.NewRSAKeyGenerator(0)
	if err != nil {
		panic(err)
	}
	kgc.SecretName = jwtSecretName
	//Generate and save key
	err = kgc.GenerateAndSaveKeyPair()
	err = kgc.VerifyKeyPair()
	if err != nil {
		panic(err)
	}
}
func TestLogin(t *testing.T) {
	log := logger.Get()
	log.Infof("Loading credentials secret %s", userSecretName)
	// Create form data
	form := url.Values{}
	form.Add("username", testUser)
	form.Add("password", testPassword)

	//Load JWT Keys
	sv, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(jwtSecretName),
	})
	assert.NoError(t, err)
	str := *sv.SecretString
	assert.NotNil(t, str)
	var epk security.EncryptedKeyPair
	err = json.Unmarshal([]byte(str), &epk)
	assert.NoError(t, err)
	assert.NotNil(t, epk)

	kgc.KeyInfo.SetPassPhrase(epk.Passphrase)
	privKey, err := kgc.DecodePrivateKey(epk.EncryptedPrivateKey)
	assert.NoError(t, err)
	assert.NotNil(t, privKey)
	pubKey, err := kgc.DecodePublicKey(epk.PublicKey)
	assert.NoError(t, err)
	assert.NotNil(t, pubKey)

	//Create request
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	//Create response recorder
	rec := httptest.NewRecorder()
	//Endpoint Config
	ec := EndpointConfig{
		UserSecretName: userSecretName,
		Manager: &security.JWTManager{
			Config: security.JWTConfig{
				PrivateKey: privKey,
				PublicKey:  pubKey,
			},
		},
	}

	//Fire the request
	e := echo.New()
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
			return pubKey, nil
		})
		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		// Verify the claims
		if claims, ok := parsedToken.Claims.(*security.JWTClaims); assert.True(t, ok) {
			assert.Equal(t, testUser, claims.Username)
			assert.Equal(t, testUserName, claims.Name)
			assert.Equal(t, testUserEmail, claims.Email)
			assert.Equal(t, testUserRole, claims.Role)
			assert.NotZero(t, claims.ExpiresAt)
			// Verify expiration is set correctly (about 1 hour from now)
			assert.WithinDuration(t, time.Now().Add(time.Hour), claims.ExpiresAt.Time, 5*time.Second)
		}
	}
	_, err = client.DeleteSecret(context.Background(), &secretsmanager.DeleteSecretInput{SecretId: aws.String(jwtSecretName)})
	assert.NoError(t, err)
	_, err = client.DeleteSecret(context.Background(), &secretsmanager.DeleteSecretInput{SecretId: aws.String(userSecretName)})
	assert.NoError(t, err)
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
