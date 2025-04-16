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
	"github.com/kameshsampath/balloon-popper/pkg/logger"
	"github.com/kameshsampath/balloon-popper/pkg/security"
	"github.com/spf13/cobra"
	"os"
)

type JWTKeysOptions struct {
	bits          int
	secretName    string
	usePassphrase bool
}

func (jwtOpts *JWTKeysOptions) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVarP(&jwtOpts.secretName, "secret-name", "s", fmt.Sprintf("%s-jwt-keys", os.Getenv("USER")),
		"Secret name to store the RSA Key Pair in AWS Secrets Manager")
	flags.IntVarP(&jwtOpts.bits, "bits", "b", 4096,
		"RSA key size in bits")
	flags.BoolVarP(&jwtOpts.usePassphrase, "use-passphrase", "p", true,
		"Encrypt private key with passphrase.")
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

	rsaKeyPairConfig, err := security.NewRSAKeyGenerator(jwtOpts.bits)
	if err != nil {
		return err
	}
	rsaKeyPairConfig.SecretName = jwtOpts.secretName

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
	err = rsaKeyPairConfig.GenerateAndSaveKeyPair()
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
  # Generate keys with default settings (4096 bits, encrypted, in AWS Secrets Manager with key name "$USER-jwt-keys")
  %[1]s jwt-keys

  # Generate unencrypted keys and save in AWS Secrets Manager	
  %[1]s jwt-keys --secret-name my-secret-jwt-keys --use-passphrase=false

  # Generate 2048-bit keys
  %[1]s jwt-keys -b 2048 ---secret-name my-secret-jwt-keys --use-passphrase=false
`, ExamplePrefix())

// NewJWTKeysCommand starts the Balloon Popper Server
func NewJWTKeysCommand() *cobra.Command {
	jwtOpts := &JWTKeysOptions{}

	jwtKeyCommand := &cobra.Command{
		Use:     "jwt-keys",
		Short:   "Generate RSA keys for signing the JWT Tokens and store the keys in AWS Secrets Manager",
		Example: jwtKeysCommandExample,
		RunE:    jwtOpts.Execute,
		PreRunE: jwtOpts.Validate,
		PostRunE: func(cmd *cobra.Command, args []string) error {
			log := logger.Get()

			log.Infoln("Successfully generated RSA key pair")
			if jwtOpts.usePassphrase {
				log.Infoln("Private key is encrypted with passphrase")
			}

			return nil
		},
	}

	jwtOpts.AddFlags(jwtKeyCommand)

	return jwtKeyCommand
}

var _ Command = (*JWTKeysOptions)(nil)
