package security

import (
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"github.com/youmark/pkcs8"
	"os"
	"path/filepath"
)

var _ PrivateKeyDecryptor = (*PrivateKeyDecryptorConfig)(nil)

// NewRSAKeyDecryptor builds a new RSAKeyDecryptor instance to decrypt RSA Private Keys
func NewRSAKeyDecryptor(privateKeyFile string) (*PrivateKeyDecryptorConfig, error) {
	return loadKey(privateKeyFile)
}

// IsEncrypted checks if the RSA Key is encrypted or not
func (d *PrivateKeyDecryptorConfig) IsEncrypted() bool {
	return d.isLocked
}

// Decrypt the private key
func (d *PrivateKeyDecryptorConfig) Decrypt() error {

	if !d.isLocked {
		return nil // Already decrypted
	}

	block, _ := pem.Decode(d.rawPEM)
	if block == nil {
		return ErrNotPEMFormat
	}

	if block.Type != BlockTypeEncrypted {
		return ErrInvalidKeyFormat
	}

	// Decrypt the private key using PKCS#8
	privateKey, err := pkcs8.ParsePKCS8PrivateKeyRSA(block.Bytes, d.KeyInfo.passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt PKCS#8 key: %w", err)
	}

	d.KeyInfo = &KeyInfo{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}
	d.isLocked = false
	return nil
}

// loadKey loads the Private Key file, loads it if its unencrypted else gets raw bytes for decrypting
func loadKey(privateKeyFile string) (*PrivateKeyDecryptorConfig, error) {
	var pemData []byte
	var err error
	if pemData, err = os.ReadFile(filepath.Clean(privateKeyFile)); err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	if len(pemData) == 0 {
		return nil, ErrEmptyPEM
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrNotPEMFormat
	}

	switch block.Type {
	case "PRIVATE KEY": // Unencrypted PKCS#8
		privateKey, err := pkcs8.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
		}
		rsaKey, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, ErrNotRSAPrivateKey
		}

		return &PrivateKeyDecryptorConfig{
			KeyInfo: &KeyInfo{
				privateKey: rsaKey,
				publicKey:  &rsaKey.PublicKey,
			},
			rawPEM:   pemData,
			isLocked: false,
		}, nil

	case BlockTypeEncrypted: // Encrypted PKCS#8
		return &PrivateKeyDecryptorConfig{
			rawPEM:   pemData,
			isLocked: true,
			KeyInfo:  &KeyInfo{},
		}, nil

	default:
		return nil, ErrInvalidKeyFormat
	}
}
