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
	"github.com/kameshsampath/balloon-popper/pkg/models"
	"golang.org/x/crypto/bcrypt"
	"os"
	"path/filepath"
)

// LoadCredentials loads user credentials from a file
func LoadCredentials(credentialFile string) ([]models.UserCredentials, error) {
	data, err := os.ReadFile(filepath.Clean(credentialFile))
	if err != nil {
		return nil, err
	}

	var c []models.UserCredentials
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// VerifyLogin checks if username/password match stored credentials
func VerifyLogin(username, password string, creds []models.UserCredentials) *models.UserCredentials {
	for _, cred := range creds {
		if cred.Username == username {
			// Compare password with bcrypt hash
			err := bcrypt.CompareHashAndPassword([]byte(cred.Password), []byte(password))
			if err == nil {
				return &cred // Password matches
			}
			break // Username found but password doesn't match
		}
	}
	return nil // No matching credentials
}
