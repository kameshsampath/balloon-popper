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
	"encoding/json"
	"github.com/kameshsampath/balloon-popper-server/pkg/models"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCredentials(t *testing.T) {
	cwd, _ := os.Getwd()
	credentialsFile := filepath.Clean(filepath.Join(cwd, "test_users.json"))
	var want models.UserCredentials
	err := json.Unmarshal([]byte(`{
"username": "admin",
"name": "Balloon Popper Admin",
"password_hash": "$2a$10$x5NssQrS0n1QfyR4ZTB58OOHVrm6F/dFBiYkxaO2ekspZt0bwgZM6",
"email": "balloon-game-admin@example.com",
"role": "admin"
}`), &want)
	assert.Nil(t, err)
	creds, err := LoadCredentials(credentialsFile)
	assert.Nil(t, err)
	assert.NotNil(t, creds)
	assert.Lenf(t, creds, 1, "Expected 1 credentials, got %d", len(creds))
	got := creds[0]
	assert.Equalf(t, want, got, "Expected credentials to match")
}

func TestVerifyLogin(t *testing.T) {
	cwd, _ := os.Getwd()
	credentialsFile := filepath.Clean(filepath.Join(cwd, "test_users.json"))
	var want *models.UserCredentials
	err := json.Unmarshal([]byte(`{
"username": "admin",
"name": "Balloon Popper Admin",
"password_hash": "$2a$10$x5NssQrS0n1QfyR4ZTB58OOHVrm6F/dFBiYkxaO2ekspZt0bwgZM6",
"email": "balloon-game-admin@example.com",
"role": "admin"
}`), &want)
	assert.Nil(t, err)
	creds, err := LoadCredentials(credentialsFile)
	assert.Nil(t, err)
	assert.NotNil(t, creds)
	//check with the test user password
	got := VerifyLogin("admin", "sup3rSecret!", creds)
	assert.NotNil(t, got)
	assert.Equalf(t, want, got, "Expected credentials to match")
}
