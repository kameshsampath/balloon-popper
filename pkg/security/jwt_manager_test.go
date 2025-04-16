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
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	var awsRegion string
	var err error

	kgc, err := NewRSAKeyGenerator(0)
	assert.Nil(t, err)
	assert.NotNil(t, kgc)
	kgc.SecretName = "kameshs-bgd-jwt-secret-1"
	kgc.KeyInfo.passphrase = []byte("password123")
	err = kgc.GenerateAndSaveKeyPair()
	assert.Nil(t, err)

	//Load PK via AWS Secret and update KGC
	if v, ok := os.LookupEnv("AWS_REGION"); ok {
		awsRegion = v
	} else {
		awsRegion = "us-west-2"
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
	)
	client = secretsmanager.NewFromConfig(cfg)
	assert.Nil(t, err)

	//Get secret
	sv, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(kgc.SecretName),
	})
	assert.Nil(t, err)
	str := *sv.SecretString
	assert.NotNil(t, str)
	var epk encryptedKeyPair
	err = json.Unmarshal([]byte(str), &epk)
	assert.Nil(t, err)
	assert.NotNil(t, epk)

	kgc.KeyInfo.passphrase = []byte(epk.Passphrase)
	kgc.KeyInfo.SetPassPhrase(epk.Passphrase)
	privKey, err := kgc.decodePrivateKey(epk.EncryptedPrivateKey)
	assert.Nil(t, err)
	assert.NotNil(t, privKey)
	pubKey, err := kgc.decodePublicKey(epk.PublicKey)
	assert.Nil(t, err)
	assert.NotNil(t, pubKey)

	jwtConfig := JWTConfig{
		PrivateKey: privKey,
		PublicKey:  pubKey,
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

	got, err := jwtManager.ValidateToken(token)
	assert.Nil(t, err, "Error validating token")

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}

	//Cleanup
	err = kgc.deleteTestSecret()
	assert.Nil(t, err)
}
