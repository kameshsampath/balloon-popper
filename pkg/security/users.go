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
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/kameshsampath/balloon-popper/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// WriteCredentials writes the user credentials to AWS Secret Manager
func (u *UserCredentials) WriteCredentials(secretName string) error {
	_, err := InitAndGetAWSSecretManagerClient()
	if err != nil {
		return err
	}
	// Marshal  to JSON
	secretValue, err := json.Marshal(u)
	if err != nil {
		return err
	}
	// Create the secret with KMS encryption
	so, err := client.CreateSecret(context.TODO(), &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(string(secretValue)),
		Description:  aws.String("Balloon Popper Demo Admin User Credentials"),
		Tags: []types.Tag{
			{
				Key:   aws.String("Owner"),
				Value: aws.String("Kamesh-DevRel"),
			},
			{
				Key:   aws.String("Type"),
				Value: aws.String("SSH"),
			},
			{
				Key:   aws.String("Service"),
				Value: aws.String("User Authentication"),
			},

			{
				Key:   aws.String("Demo"),
				Value: aws.String("BalloonPopper"),
			},
		},
	})
	log := logger.Get()
	var resourceExistsErr *types.ResourceExistsException
	if err != nil && errors.As(err, &resourceExistsErr) {
		po, err := client.PutSecretValue(context.TODO(), &secretsmanager.PutSecretValueInput{
			SecretId:     aws.String(secretName),
			SecretString: aws.String(string(secretValue)),
		})
		if err != nil {
			return err
		}
		log.Infof("Updated existing encrypted key pair secret: %s, with ARN: %s", secretName, *po.ARN)
	} else if err != nil {
		log.Errorf("Error:%v", err)
		return err
	} else {
		log.Infof("Created new encrypted key pair secret: %s, with ARN: %s", secretName, *so.ARN)
	}

	return nil
}

// LoadCredentials loads user credentials from a file
func LoadCredentials(secretName string) (*UserCredentials, error) {
	log := logger.Get()
	_, err := InitAndGetAWSSecretManagerClient()
	if err != nil {
		return nil, err
	}
	//Load the credentials
	log.Infof("Loading credentials for %s", secretName)
	svo, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	})

	if err != nil {
		return nil, err
	}
	data := svo.SecretString
	var c UserCredentials
	err = json.Unmarshal([]byte(*data), &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// VerifyLogin checks if username/password match stored credentials
func (u *UserCredentials) VerifyLogin(username, password string) error {
	// Compare password with bcrypt hash
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err == nil {
		return nil // Password matches
	}
	return fmt.Errorf("no matching credentials for user %s", username) // No matching credentials
}
