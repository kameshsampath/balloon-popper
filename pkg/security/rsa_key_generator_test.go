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
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestGenerateKeyPairWithPassphrase(t *testing.T) {
	kgc := NewRSAKeyGenerator(0)
	assert.NotNil(t, kgc)
	kgc.PrivateKeyFile = DefaultPrivateKeyFileName
	kgc.PublicKeyFile = DefaultPublicKeyFileName
	kgc.KeyInfo.passphrase = []byte("password123")
	err := kgc.GenerateKeyPair()
	assert.Nil(t, err)
	_, err = os.Stat(kgc.PrivateKeyFile)
	assert.Nil(t, err)
	_, err = os.Stat(kgc.PublicKeyFile)
	assert.Nil(t, err)
	passFile := filepath.Join(path.Dir(kgc.PrivateKeyFile), ".pass")
	_, err = os.Stat(passFile)
	assert.Nil(t, err)
	err = kgc.VerifyKeyPair()
	assert.Nil(t, err)
	_ = os.Remove(kgc.PrivateKeyFile)
	_ = os.Remove(kgc.PublicKeyFile)
	_ = os.Remove(passFile)
}

func TestGenerateKeyPairWithoutPassphrase(t *testing.T) {
	kgc := NewRSAKeyGenerator(0)
	assert.NotNil(t, kgc)
	kgc.PrivateKeyFile = DefaultPrivateKeyFileName
	kgc.PublicKeyFile = DefaultPublicKeyFileName
	err := kgc.GenerateKeyPair()
	assert.Nil(t, err)
	_, err = os.Stat(kgc.PrivateKeyFile)
	assert.Nil(t, err)
	_, err = os.Stat(kgc.PublicKeyFile)
	assert.Nil(t, err)
	passFile := filepath.Join(path.Dir(kgc.PrivateKeyFile), ".pass")
	_, err = os.Stat(passFile)
	assert.ErrorIs(t, err, os.ErrNotExist, fmt.Sprintf("Expected file %s not to exist", passFile))
	err = kgc.VerifyKeyPair()
	assert.Nil(t, err)
	_ = os.Remove(kgc.PrivateKeyFile)
	_ = os.Remove(kgc.PublicKeyFile)
}
