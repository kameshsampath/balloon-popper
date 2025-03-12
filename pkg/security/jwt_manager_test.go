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

package security

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	kgc := NewRSAKeyGenerator(0)
	assert.NotNil(t, kgc)
	kgc.PrivateKeyFile = DefaultPrivateKeyFileName
	kgc.PublicKeyFile = DefaultPublicKeyFileName
	kgc.KeyInfo.passphrase = []byte("password123")
	err := kgc.GenerateKeyPair()
	assert.Nil(t, err)
	jwtConfig := JWTConfig{
		PrivateKey: kgc.KeyInfo.privateKey,
		PublicKey:  kgc.KeyInfo.publicKey,
		Issuer:     "TestGenerateToken",
		ExpiryTime: time.Minute * 3,
	}
	jwtManager := NewJWTManager(jwtConfig)
	want := &JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
		},
		Name:  "Balloon Game Admin",
		Email: "balloon-game-admin@example.com",
		Role:  "admin",
	}
	token, err := jwtManager.GenerateToken(want)
	assert.Nil(t, err, "Error generating token")
	assert.NotNil(t, token, "Expecting token to be not nil but it is")
	passFile := filepath.Join(path.Dir(kgc.PrivateKeyFile), ".pass")

	got, err := jwtManager.ValidateToken(token)
	assert.Nil(t, err, "Error validating token")

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	_ = os.Remove(kgc.PrivateKeyFile)
	_ = os.Remove(kgc.PublicKeyFile)
	_ = os.Remove(passFile)
}
