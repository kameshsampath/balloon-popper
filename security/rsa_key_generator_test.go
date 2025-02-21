package security

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestGenerateKeyPairWithPassphrase(t *testing.T) {
	kgc := NewRSAKeyGenerator(0)
	assert.NotNil(t, kgc)
	kgc.PrivateKeyFile = "jwt-test-key.pem"
	kgc.PublicKeyFile = "jwt-test-key.pub"
	kgc.keyInfo.passphrase = []byte("password123")
	err := kgc.GenerateKeyPair()
	assert.Nil(t, err)
	_, err = os.Stat(kgc.PrivateKeyFile)
	assert.Nil(t, err)
	_, err = os.Stat(kgc.PublicKeyFile)
	assert.Nil(t, err)
	passFile := filepath.Join(path.Dir(kgc.PrivateKeyFile), ".pass")
	_, err = os.Stat(passFile)
	assert.Nil(t, err)
	err = kgc.VerifyKeyPair()
	assert.Nil(t, err)
	os.Remove(kgc.PrivateKeyFile)
	os.Remove(kgc.PublicKeyFile)
	os.Remove(passFile)
}
