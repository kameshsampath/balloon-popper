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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/kameshsampath/balloon-popper-server/pkg/logger"
	"github.com/youmark/pkcs8"
	"os"
	"path"
	"path/filepath"
)

var (
	_   RSAKeyGenerator = (*Config)(nil)
	log                 = logger.Get()
)

// NewRSAKeyGenerator creates the new instance of generator
func NewRSAKeyGenerator(bits int) *Config {
	if bits == 0 {
		bits = 2048
	}

	return &Config{
		KeyInfo: &KeyInfo{
			bits: bits,
		},
	}
}

func (r *Config) GenerateKeyPair() error {
	key, err := rsa.GenerateKey(rand.Reader, r.KeyInfo.bits)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key pair: %v", err)
	}

	//if no error set the key
	r.KeyInfo.publicKey = &key.PublicKey
	r.KeyInfo.privateKey = key

	// Save private key
	if err := r.savePrivateKeyPKCS8(); err != nil {
		fmt.Printf("Error saving private key: %v\n", err)
		os.Exit(1)
	}

	// Save public key
	if err := r.savePublicKey(); err != nil {
		fmt.Printf("Error saving public key: %v\n", err)
		os.Exit(1)
	}

	log.Infof("using passphrase for RSA key pair,%s", r.KeyInfo.passphrase)

	if r.KeyInfo.passphrase != nil {
		err = r.savePassFile()
		if err != nil {
			fmt.Printf("Error saving private key pass file: %v\n", err)
			return err
		}
	}

	return nil
}

// VerifyKeyPair attempts to read and verify the generated key pair
func (r *Config) VerifyKeyPair() error {
	// Read private key file
	privateKeyData, err := os.ReadFile(r.PrivateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %v", err)
	}

	// Decode PEM block
	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	// Parse private key
	var privateKey interface{}
	privateKey, err = pkcs8.ParsePKCS8PrivateKey(block.Bytes, r.KeyInfo.passphrase)

	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	// Check if it's an RSA key
	if _, ok := privateKey.(*rsa.PrivateKey); !ok {
		return fmt.Errorf("not an RSA private key")
	}

	return nil
}

// savePrivateKeyPKCS8 saves the PrivateKey to file
func (r *Config) savePrivateKeyPKCS8() error {
	var privateKeyBytes []byte
	var err error

	privateKeyBytes, err = pkcs8.MarshalPrivateKey(r.KeyInfo.privateKey, r.KeyInfo.passphrase, nil)

	if err != nil {
		return fmt.Errorf("failed to marshal private key to PKCS#8: %v", err)
	}

	// Create PEM block
	privatePEM := &pem.Block{
		Type: func() string {
			if r.KeyInfo.passphrase != nil {
				return "ENCRYPTED PRIVATE KEY"
			}
			return "PRIVATE KEY" // standard type for PKCS#8
		}(),
		Bytes: privateKeyBytes,
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(r.PrivateKeyFile), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Save private key
	privateFile, err := os.OpenFile(r.PrivateKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open private key file: %v", err)
	}
	defer privateFile.Close() //nolint:errcheck

	if err := pem.Encode(privateFile, privatePEM); err != nil {
		return fmt.Errorf("failed to write private key: %v", err)
	}

	return nil
}

// savePublicKey saves the Public Key to file
func (r *Config) savePublicKey() error {
	// Marshal public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(r.KeyInfo.publicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %v", err)
	}

	// Create PEM block
	publicPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(r.PublicKeyFile), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Save public key
	publicFile, err := os.OpenFile(r.PublicKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open public key file: %v", err)
	}
	defer publicFile.Close() //nolint:errcheck

	if err := pem.Encode(publicFile, publicPEM); err != nil {
		return fmt.Errorf("failed to write public key: %v", err)
	}

	return nil
}

// savePassFile saves the private key passphrase into a file
// TODO: encrypt and save
func (r *Config) savePassFile() error {
	// Save password file
	passFilePath := filepath.Join(path.Dir(r.PrivateKeyFile), ".pass")
	passFile, err := os.OpenFile(filepath.Clean(passFilePath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(fmt.Errorf("failed to open pass key file: %v", err))
	}
	defer passFile.Close() //nolint:errcheck
	_, err = passFile.Write(r.KeyInfo.passphrase)
	if err != nil {
		return err
	}
	return nil
}
