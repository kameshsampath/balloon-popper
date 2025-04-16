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
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLoadAndVerifyCredentials(t *testing.T) {
	_, err := InitAndGetAWSSecretManagerClient()
	assert.Nil(t, err)

	//write the credentials to AWS Secret Manager
	var u UserCredentials
	err = json.Unmarshal([]byte(`{
"username": "test",
"name": "Balloon Popper Test Admin",
"password_hash": "$2a$10$x5NssQrS0n1QfyR4ZTB58OOHVrm6F/dFBiYkxaO2ekspZt0bwgZM6",
"email": "balloon-game-test-admin@example.com",
"role": "admin"
}`), &u)
	assert.NoError(t, err)
	suffix := time.Now().UnixMilli()
	err = u.WriteCredentials(fmt.Sprintf("bgd-user-%s-%d", u.Username, suffix))
	assert.NoError(t, err)

	//Verify credentials
	err = u.VerifyLogin("admin", "sup3rSecret!")
	assert.Nil(t, err)

	//clean up
	err = u.deleteTestSecret()
	assert.Nil(t, err)
}

func (u *UserCredentials) deleteTestSecret() error {
	sid := fmt.Sprintf("bgd-user-%s", u.Username)
	_, err := client.DeleteSecret(context.TODO(), &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(sid),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	return err
}
