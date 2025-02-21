package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/youmark/pkcs8"
	"os"
	"path"
	"path/filepath"
)

var _ RSAKeyGenerator = (*Config)(nil)

// NewRSAKeyGenerator creates the new instance of generator
func NewRSAKeyGenerator(bits int) *Config {
	if bits == 0 {
		bits = 2048
	}

	return &Config{
		keyInfo: &KeyInfo{
			bits: bits,
		},
	}
}

func (r *Config) GenerateKeyPair() error {
	key, err := rsa.GenerateKey(rand.Reader, r.keyInfo.bits)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key pair: %v", err)
	}

	//if no error set the key
	r.keyInfo = &KeyInfo{
		publicKey:  &key.PublicKey,
		privateKey: key,
	}

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

	// Save password file
	passFilePath := filepath.Join(path.Dir(r.PrivateKeyFile), ".pass")
	passFile, err := os.OpenFile(passFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to open pass key file: %v", err))
	}
	defer passFile.Close()
	_, err = passFile.Write(r.keyInfo.passphrase)
	if err != nil {
		return err
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
	privateKey, err = pkcs8.ParsePKCS8PrivateKey(block.Bytes, r.keyInfo.passphrase)

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

	privateKeyBytes, err = pkcs8.MarshalPrivateKey(r.keyInfo.privateKey, r.keyInfo.passphrase, nil)

	if err != nil {
		return fmt.Errorf("failed to marshal private key to PKCS#8: %v", err)
	}

	// Create PEM block
	privatePEM := &pem.Block{
		Type: func() string {
			if r.keyInfo.passphrase != nil {
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
	defer privateFile.Close()

	if err := pem.Encode(privateFile, privatePEM); err != nil {
		return fmt.Errorf("failed to write private key: %v", err)
	}

	return nil
}

// savePublicKey saves the Public Key to file
func (r *Config) savePublicKey() error {
	// Marshal public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(r.keyInfo.publicKey)
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
	publicFile, err := os.OpenFile(r.PublicKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open public key file: %v", err)
	}
	defer publicFile.Close()

	if err := pem.Encode(publicFile, publicPEM); err != nil {
		return fmt.Errorf("failed to write public key: %v", err)
	}

	return nil
}
