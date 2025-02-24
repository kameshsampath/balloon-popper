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

package commands

import (
	"fmt"
	"github.com/kameshsampath/balloon-popper-server/pkg/logger"
	"github.com/kameshsampath/balloon-popper-server/pkg/security"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

type JWTKeysOptions struct {
	bits               int
	outDir             string
	privateKeyFilename string
	publicKeyFilename  string
	usePassphrase      bool
}

func (jwtOpts *JWTKeysOptions) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVarP(&jwtOpts.outDir, "out-dir", "d", "keys",
		"Directory where the keys are stored")
	flags.StringVarP(&jwtOpts.privateKeyFilename, "private-key-file", "k", "jwt-private-key",
		"RSA private key filename")
	flags.IntVarP(&jwtOpts.bits, "bits", "b", 4096,
		"RSA key size in bits")
	flags.BoolVarP(&jwtOpts.usePassphrase, "use-passphrase", "p", true,
		"Encrypt private key with passphrase (stored in .pass file)")
}

func (jwtOpts *JWTKeysOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (jwtOpts *JWTKeysOptions) Execute(cmd *cobra.Command, args []string) error {
	log := logger.Get()
	var err error

	// Validate key size
	validKeySizes := map[int]bool{2048: true, 3072: true, 4096: true}
	if !validKeySizes[jwtOpts.bits] {
		log.Warnf("Warning: Non-standard key size. Recommended sizes are 2048 or 4096 bits")
	}

	// Prepare file paths
	jwtOpts.publicKeyFilename = jwtOpts.privateKeyFilename + ".pub"
	privateKeyPath := filepath.Join(jwtOpts.outDir, jwtOpts.privateKeyFilename)
	publicKeyPath := filepath.Join(jwtOpts.outDir, jwtOpts.publicKeyFilename)

	rsaKeyPairConfig := security.NewRSAKeyGenerator(jwtOpts.bits)
	rsaKeyPairConfig.PrivateKeyFile = privateKeyPath
	rsaKeyPairConfig.PublicKeyFile = publicKeyPath

	if jwtOpts.usePassphrase {
		var passphrase string
		if v, ok := os.LookupEnv("JWT_SIGNING_KEY_PASSPHRASE"); ok {
			log.Info("Using JWT_SIGNING_KEY_PASSPHRASE from environment variable.")
			passphrase = v
		} else {
			log.Info("No JWT_SIGNING_KEY_PASSPHRASE, generating a new passphrase.")
			pwdGenerator := security.NewPasswordGenerator()
			passphrase, err = pwdGenerator.GeneratePassword()
			if err != nil {
				log.Fatal(err)
			}

		}
		rsaKeyPairConfig.KeyInfo.SetPassPhrase(passphrase)
	}

	// Generate key pair
	err = rsaKeyPairConfig.GenerateKeyPair()
	if err != nil {
		log.Fatal(err)
	}
	// Verify the key pair
	if err := rsaKeyPairConfig.VerifyKeyPair(); err != nil {
		log.Fatal(err)
	}

	return nil
}

var jwtKeysCommandExample = fmt.Sprintf(`
  # Generate keys with default settings (4096 bits, encrypted, in ./keys directory)
  %[1]s jwt-keys

  # Generate unencrypted keys in custom directory
  %[1]s jwt-keys --out-dir /my/keys --use-passphrase=false

  # Generate 2048-bit keys with custom filenames
  %[1]s jwt-keys -b 2048 --private-key-file my-private --public-key-file my-public

  # Generate keys in specific directory with custom private key name
  %[1]s jwt-keys -d /app/certs -k custom-key

  # Generate keys with all custom options
  %[1]s jwt-keys --out-dir /certs \
    --private-key-file auth-private \
    --public-key-file auth-public \
    --bits 3072 \
    --use-passphrase=true

  # Short form with all options
  %[1]s jwt-keys -d /certs -k auth-private -f auth-public -b 3072 -p
`, ExamplePrefix())

// NewJWTKeysCommand starts the Balloon Popper Server
func NewJWTKeysCommand() *cobra.Command {
	jwtOpts := &JWTKeysOptions{}

	jwtKeyCommand := &cobra.Command{
		Use:     "jwt-keys ",
		Short:   "Generate RSA keys for signing the JWT Tokens",
		Example: jwtKeysCommandExample,
		RunE:    jwtOpts.Execute,
		PreRunE: jwtOpts.Validate,
		PostRunE: func(cmd *cobra.Command, args []string) error {
			log := logger.Get()

			log.Infof("\nSuccessfully generated %d-bit RSA key pair:\n", jwtOpts.bits)
			log.Infof("Private key (PKCS#8): %s\n", jwtOpts.privateKeyFilename)
			log.Infof("Public key: %s\n", jwtOpts.publicKeyFilename)
			if jwtOpts.usePassphrase {
				log.Infof("Private key is encrypted with passphrase")
			}
			log.Infof("\nFile permissions:\n")
			log.Infof("Private key: 0600 (rw-------)\n")
			log.Infof("Public key:  0644 (rw-r--r--)\n")

			return nil
		},
	}

	jwtOpts.AddFlags(jwtKeyCommand)

	return jwtKeyCommand
}

var _ Command = (*JWTKeysOptions)(nil)
