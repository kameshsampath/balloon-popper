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
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateKeyPairWithPassphrase(t *testing.T) {
	_, err := InitAndGetAWSSecretManagerClient()
	assert.Nil(t, err)
	kgc, err := NewRSAKeyGenerator(0)
	assert.Nil(t, err)
	assert.NotNil(t, kgc)
	kgc.SecretName = "kameshs-bgd-server-jwt-test-1"
	kgc.KeyInfo.passphrase = []byte("password123")
	err = kgc.GenerateAndSaveKeyPair()
	assert.Nil(t, err)
	err = kgc.VerifyKeyPair()
	assert.Nil(t, err)
	//Cleanup
	err = kgc.deleteTestSecret()
	assert.Nil(t, err)
}

func TestGenerateKeyPairWithoutPassphrase(t *testing.T) {
	_, err := InitAndGetAWSSecretManagerClient()
	assert.Nil(t, err)

	kgc, err := NewRSAKeyGenerator(0)
	assert.Nil(t, err)
	assert.NotNil(t, kgc)
	kgc.SecretName = "kameshs-bgd-server-jwt-test-2"
	//Generate and save key
	err = kgc.GenerateAndSaveKeyPair()
	err = kgc.VerifyKeyPair()
	assert.Nil(t, err)
	//Get secret and ensure password is nil
	sv, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(kgc.SecretName),
	})
	assert.Nil(t, err)
	str := *sv.SecretString
	assert.NotNil(t, str)
	var epk EncryptedKeyPair
	err = json.Unmarshal([]byte(str), &epk)
	assert.Nil(t, err)
	assert.NotNil(t, epk)
	assert.Equalf(t, "", epk.Passphrase, "Passphrase should be empty")
	//Cleanup
	err = kgc.deleteTestSecret()
	assert.Nil(t, err)
}

func (c *Config) deleteTestSecret() error {
	_, err := client.DeleteSecret(context.TODO(), &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(c.SecretName),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	return err
}
