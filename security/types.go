package security

import (
	"crypto/rsa"
	"errors"
)

// RSAKeyGenerator handles RSA key pair generation and verification
type RSAKeyGenerator interface {
	// GenerateKeyPair creates a new RSA key pair
	GenerateKeyPair() error // Fixed typo from "GeneratorKeyPair"
	//VerifyKeyPair verifies the Generated Key pair like able to decrypt with passphrase
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
	//PrivateKeyFile is the absolute path to save the generated Private Key
	PrivateKeyFile string
	//PublicKeyFile is the absolute path to save the generated Public Key
	PublicKeyFile string
	keyInfo       *KeyInfo
}

// PrivateKeyDecryptorConfig defines the Private Key decryptor config
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

var (
	ErrEmptyPEM         = errors.New("PEM data is empty")
	ErrNotPEMFormat     = errors.New("not a PEM format")
	ErrNotRSAPrivateKey = errors.New("not an RSA private key")
	ErrInvalidKeyFormat = errors.New("invalid key format: only PKCS#8 format is supported")
)

func (k *KeyInfo) Bits() int                       { return k.bits }
func (k *KeyInfo) PublicKey() *rsa.PublicKey       { return k.publicKey }
func (k *KeyInfo) SetPassPhrase(passphrase string) { k.passphrase = []byte(passphrase) }
