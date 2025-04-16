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
	"github.com/kameshsampath/balloon-popper/pkg/security"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

// UserCreateOptions defines the structure for storing credentials
type UserCreateOptions struct {
	secretName   string
	name         string
	email        string
	hashPassword bool
	password     string
	role         string
	username     string
}

func (u *UserCreateOptions) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVarP(&u.username, "user-name", "u", "",
		"Create user with specified username.")
	flags.StringVarP(&u.password, "user-password", "p", "",
		"Password for user")
	flags.StringVarP(&u.name, "name", "n", "",
		"Full name of the user.If not provided user-name will be used.")
	flags.StringVarP(&u.role, "role", "r", "user",
		"Role of the user")
	flags.StringVarP(&u.email, "email", "e", "",
		"Email of the user")
	flags.BoolVarP(&u.hashPassword, "std-out", "", false,
		"Output bcrypt hash for a password and exit")

	// Mark required flags
	cobra.CheckErr(cmd.MarkFlagRequired("user-name"))
	cobra.CheckErr(cmd.MarkFlagRequired("user-password"))
}

func (u *UserCreateOptions) Validate(_ *cobra.Command, _ []string) error {
	return nil
}

func (u *UserCreateOptions) Execute(_ *cobra.Command, _ []string) error {
	u.secretName = fmt.Sprintf("bgd-user-%s", u.username)
	// If hash-password flag is set, just output the hash and exit
	if u.hashPassword {
		password := u.password
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		fmt.Println(string(hash))
		return nil
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if u.name == "" {
		u.name = u.username
	}

	// Create user
	user := security.UserCredentials{
		Username: u.username,
		Password: string(hash),
		Name:     u.name,
		Email:    u.email,
		Role:     u.role,
	}

	if err := user.WriteCredentials(u.secretName); err != nil {
		return err
	}

	return nil
}

var userCommandExample = fmt.Sprintf(`
  # Create a new user
  %[1]s user --user-name admin --user-password secretpassword

  # Create a user with explicit path to credentials file
  %[1]s user -u manager -w managerpass -s kameshs-demo-admin-user

  # Generate a bcrypt hash for a password
  %[1]s user -u someuser -p complexpass --std-out

  # Short form with all options
  %[1]s user -u admin -p adminpass -s kameshs-demo-admin-user
`, ExamplePrefix())

func NewUserCommand() *cobra.Command {

	userCreateOpts := &UserCreateOptions{}

	userCommand := &cobra.Command{
		Use:   "user",
		Short: "Create a new user and save it AWS Secrets Manager",
		Long: `
Create a new user and save it AWS Secrets Manager. A secret will be generated for each user.
`,
		Example: userCommandExample,
		RunE:    userCreateOpts.Execute,
		PreRunE: userCreateOpts.Validate,
	}

	userCreateOpts.AddFlags(userCommand)

	return userCommand
}

var _ Command = (*UserCreateOptions)(nil)
