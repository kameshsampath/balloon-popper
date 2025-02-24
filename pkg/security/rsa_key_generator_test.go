package security

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestGenerateKeyPairWithPassphrase(t *testing.T) {
	kgc := NewRSAKeyGenerator(0)
	assert.NotNil(t, kgc)
	kgc.PrivateKeyFile = DefaultPrivateKeyFileName
	kgc.PublicKeyFile = DefaultPublicKeyFileName
	kgc.KeyInfo.passphrase = []byte("password123")
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
	_ = os.Remove(kgc.PrivateKeyFile)
	_ = os.Remove(kgc.PublicKeyFile)
	_ = os.Remove(passFile)
}

func TestGenerateKeyPairWithoutPassphrase(t *testing.T) {
	kgc := NewRSAKeyGenerator(0)
	assert.NotNil(t, kgc)
	kgc.PrivateKeyFile = DefaultPrivateKeyFileName
	kgc.PublicKeyFile = DefaultPublicKeyFileName
	err := kgc.GenerateKeyPair()
	assert.Nil(t, err)
	_, err = os.Stat(kgc.PrivateKeyFile)
	assert.Nil(t, err)
	_, err = os.Stat(kgc.PublicKeyFile)
	assert.Nil(t, err)
	passFile := filepath.Join(path.Dir(kgc.PrivateKeyFile), ".pass")
	_, err = os.Stat(passFile)
	assert.ErrorIs(t, err, os.ErrNotExist, fmt.Sprintf("Expected file %s not to exist", passFile))
	err = kgc.VerifyKeyPair()
	assert.Nil(t, err)
	_ = os.Remove(kgc.PrivateKeyFile)
	_ = os.Remove(kgc.PublicKeyFile)
}
