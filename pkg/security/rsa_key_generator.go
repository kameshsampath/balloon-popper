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
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/kameshsampath/balloon-popper/pkg/logger"
	"github.com/youmark/pkcs8"
	"os"
)

const (
	RsaKeyFormatPkcs8 = "PKCS#8"
)

var (
	_ RSAKeyGenerator = (*Config)(nil)
)

// EncryptedKeyPair represents a structure for storing encrypted key information
type EncryptedKeyPair struct {
	EncryptedPrivateKey string `json:"private_key"`
	PublicKey           string `json:"public_key"`
	Passphrase          string `json:"passphrase"`
	KeyFormat           string `json:"key_format"`         // e.g., "PEM", "PKCS#8"
	KeyBits             int    `json:"key_bits,omitempty"` // e.g., 2048, 4096 for RSA
}

// NewRSAKeyGenerator creates the new instance of generator
func NewRSAKeyGenerator(bits int) (*Config, error) {
	// set the encrypting bits to 4096 by default
	if bits == 0 {
		bits = 4096
	}
	return &Config{
		KeyInfo: &KeyInfo{
			bits: bits,
		},
	}, nil
}

// GenerateAndSaveKeyPair generate the PKCS#8 RSA KeyPair and stores them in the AWS Secret Manager
func (c *Config) GenerateAndSaveKeyPair() error {

	key, err := rsa.GenerateKey(rand.Reader, c.KeyInfo.bits)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key pair: %v", err)
	}

	c.KeyInfo.privateKey = key
	c.KeyInfo.publicKey = &key.PublicKey

	// Save private key
	var privKey, pubKey string
	if privKey, err = c.encryptPrivateKeyAsString(); err != nil {
		fmt.Printf("Error encrypting private key: %v\n", err)
		os.Exit(1)
	}

	if pubKey, err = c.publicKeyAsString(); err != nil {
		fmt.Printf("Error building public key string: %v\n", err)
		os.Exit(1)
	}

	keyPair := EncryptedKeyPair{
		EncryptedPrivateKey: privKey,
		PublicKey:           pubKey,
		Passphrase:          string(c.KeyInfo.passphrase),
		KeyFormat:           RsaKeyFormatPkcs8,
		KeyBits:             c.KeyInfo.bits,
	}

	// Marshal encrypted key pair to JSON
	secretValue, err := json.Marshal(keyPair)
	if err != nil {
		return err
	}

	// Create the secret with KMS encryption
	_, err = InitAndGetAWSSecretManagerClient()
	if err != nil {
		return err
	}
	so, err := client.CreateSecret(context.TODO(), &secretsmanager.CreateSecretInput{
		Name:         aws.String(c.SecretName),
		SecretString: aws.String(string(secretValue)),
		Description:  aws.String("Encrypted SSH key pair with passphrase"),
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
				Value: aws.String("JWT Authentication"),
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
			SecretId:     aws.String(c.SecretName),
			SecretString: aws.String(string(secretValue)),
		})
		if err != nil {
			return err
		}
		log.Infof("Updated existing encrypted key pair secret: %s, with ARN: %s", c.SecretName, *po.ARN)
	} else if err != nil {
		log.Errorf("Error:%v", err)
		return err
	} else {
		log.Infof("Created new encrypted key pair secret: %s, with ARN: %s", c.SecretName, *so.ARN)
	}

	return nil
}

// VerifyKeyPair attempts to retrieve and verify the key pair from AWS Secrets Manager
func (c *Config) VerifyKeyPair() error {
	// Get the secret
	result, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(c.SecretName),
	})
	if err != nil {
		return fmt.Errorf("failed to get secret from AWS Secrets Manager: %v", err)
	}

	if result.SecretString == nil {
		return fmt.Errorf("secret value is empty")
	}

	// Parse the secret JSON
	var keyData struct {
		PrivateKey string `json:"private_key"`
	}
	if err := json.Unmarshal([]byte(*result.SecretString), &keyData); err != nil {
		return fmt.Errorf("failed to unmarshal secret data: %v", err)
	}

	_, err = c.DecodePrivateKey(keyData.PrivateKey)

	if err != nil {
		return err
	}

	return nil
}

// DecodePrivateKey decodes a PEM-encoded PKCS#8 private key string and returns an RSA private key
func (c *Config) DecodePrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	// Decode PEM block
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse private key based on PEM block type
	var privateKey interface{}
	var err error

	switch block.Type {
	case "ENCRYPTED PRIVATE KEY":
		// This is an encrypted PKCS#8 key, use the passphrase
		if c.KeyInfo.passphrase == nil {
			return nil, fmt.Errorf("encrypted private key found but no passphrase provided")
		}
		privateKey, err = pkcs8.ParsePKCS8PrivateKey(block.Bytes, c.KeyInfo.passphrase)
	case "PRIVATE KEY":
		// This is an unencrypted PKCS#8 key, don't use passphrase
		privateKey, err = pkcs8.ParsePKCS8PrivateKey(block.Bytes, nil)
	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS#8 private key: %v", err)
	}

	// Check if it's an RSA key
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaKey, nil
}

// DecodePublicKey decodes a PEM-encoded PKCS#8 public key string and returns an RSA public key
func (c *Config) DecodePublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	// Decode PEM block
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Check PEM block type - PKCS#8 public keys use "PUBLIC KEY"
	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	// Parse PKCS#8 public key
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS#8 public key: %v", err)
	}

	// Check if it's an RSA key
	rsaKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaKey, nil
}

// encryptPrivateKeyAsString encrypts the Private Key to string
func (c *Config) encryptPrivateKeyAsString() (string, error) {
	var privateKeyBytes []byte
	var err error

	// If passphrase is nil, marshal without encryption
	if c.KeyInfo.passphrase == nil {
		privateKeyBytes, err = pkcs8.MarshalPrivateKey(c.KeyInfo.privateKey, nil, nil)
	} else {
		privateKeyBytes, err = pkcs8.MarshalPrivateKey(c.KeyInfo.privateKey, c.KeyInfo.passphrase, nil)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal private key to PKCS#8: %v", err)
	}

	// Create PEM block
	privatePEM := &pem.Block{
		Type: func() string {
			if c.KeyInfo.passphrase != nil {
				return "ENCRYPTED PRIVATE KEY"
			}
			return "PRIVATE KEY" // standard type for PKCS#8
		}(),
		Bytes: privateKeyBytes,
	}

	var privateKeyBuffer bytes.Buffer
	if err := pem.Encode(&privateKeyBuffer, privatePEM); err != nil {
		return "", fmt.Errorf("failed to encode private key: %v", err)
	}
	privateKeyString := privateKeyBuffer.String()

	return privateKeyString, nil
}

// publicKeyAsString gets the string representation of the public key
func (c *Config) publicKeyAsString() (string, error) {
	// Marshal public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(c.KeyInfo.publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %v", err)
	}

	// Create PEM block
	publicPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	var publicKeyBuffer bytes.Buffer
	if err := pem.Encode(&publicKeyBuffer, publicPEM); err != nil {
		return "", fmt.Errorf("failed to encode private key: %v", err)
	}
	privateKeyString := publicKeyBuffer.String()

	return privateKeyString, nil
}
