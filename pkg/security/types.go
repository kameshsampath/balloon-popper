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
	"crypto/rsa"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// RSAKeyGenerator handles RSA key pair generation and verification
type RSAKeyGenerator interface {
	// GenerateAndSaveKeyPair creates a new RSA key pair and saves it to AWS Secret Manager
	GenerateAndSaveKeyPair() error // Fixed typo from "GeneratorKeyPair"
	//VerifyKeyPair verifies the Generated Key pair by retrieving from AWS Secret Manager
	VerifyKeyPair() error
}

// KeyInfo holds RSA key pair data and configuration
type KeyInfo struct { // Removed RSA prefix as it's redundant given the package name
	bits       int // Unexported as they should be accessed via methods
	passphrase []byte
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey // Fixed comment typo
}

// Config configures the RSA keypair generation
type Config struct {
	//SecretName is the AWS Secret Manager Key Name for the RSA KeyPair
	SecretName string
	KeyInfo    *KeyInfo
}

// PrivateKeyDecryptorConfig defines the Private Key decryptor Config
type PrivateKeyDecryptorConfig struct {
	rawPEM   []byte
	isLocked bool
	KeyInfo  *KeyInfo
}

// PrivateKeyDecryptor defines command to decrypt the Private Keys
type PrivateKeyDecryptor interface {
	IsEncrypted() bool
	Decrypt() error
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Issuer     string
	ExpiryTime time.Duration
}

// JWTClaims extends standard claims
type JWTClaims struct {
	jwt.RegisteredClaims
	Name     string `json:"name"`
	Username string `json:"user_name,omitempty"`
	Role     string `json:"role,omitempty"`
	Email    string `json:"email,omitempty"`
}

// JWTManager handles JWT operations
type JWTManager struct {
	Config JWTConfig
}

// UserCredentials defines the structure for storing credentials
type UserCredentials struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Password string `json:"password_hash"` // Stores bcrypt hash
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// JWTTokenCommand handles JWT commands
type JWTTokenCommand interface {
	//GenerateToken generates the JWT
	GenerateToken(claims *JWTClaims) (string, error)
	//ValidateToken validates if the Token is a valid JWT
	ValidateToken(tokenString string) (*JWTClaims, error)
}

func (k *KeyInfo) Bits() int                       { return k.bits }
func (k *KeyInfo) PrivateKey() *rsa.PrivateKey     { return k.privateKey }
func (k *KeyInfo) PublicKey() *rsa.PublicKey       { return k.publicKey }
func (k *KeyInfo) SetPassPhrase(passphrase string) { k.passphrase = []byte(passphrase) }
