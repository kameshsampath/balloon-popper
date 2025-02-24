package security

import (
	"crypto/rsa"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const (
	DefaultPrivateKeyFileName = "jwt-test-key.pem"
	DefaultPublicKeyFileName  = "jwt-test-key.pub"
	BlockTypeEncrypted        = "ENCRYPTED PRIVATE KEY"
)

var (
	ErrEmptyPEM         = errors.New("PEM data is empty")
	ErrNotPEMFormat     = errors.New("not a PEM format")
	ErrNotRSAPrivateKey = errors.New("not an RSA private key")
	ErrInvalidKeyFormat = errors.New("invalid key format: only PKCS#8 format is supported")
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
	KeyInfo       *KeyInfo
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
