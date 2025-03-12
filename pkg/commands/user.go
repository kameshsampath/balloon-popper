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
	"encoding/json"
	"fmt"
	"github.com/kameshsampath/balloon-popper/pkg/logger"
	"github.com/kameshsampath/balloon-popper/pkg/models"
	"github.com/kameshsampath/balloon-popper/pkg/security"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"os"
	"path/filepath"
)

// UserCreateOptions defines the structure for storing credentials
type UserCreateOptions struct {
	credentialsFile string
	name            string
	email           string
	hashPassword    bool
	password        string
	role            string
	username        string
}

func (u *UserCreateOptions) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVarP(&u.credentialsFile, "credentials-file", "c", "./config/users.json",
		"Path to JSON file containing user credentials")
	flags.StringVarP(&u.username, "user-name", "u", "",
		"Create user with specified username if credentials file doesn't exist")
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

func (u *UserCreateOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (u *UserCreateOptions) Execute(cmd *cobra.Command, args []string) error {
	log := logger.Get()

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

	// Create directory if needed
	dir := filepath.Dir(u.credentialsFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
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
	user := models.UserCredentials{
		Username: u.username,
		Password: string(hash),
		Name:     u.name,
		Email:    u.email,
		Role:     u.role,
	}
	// Check if credentials file exists
	users := make([]models.UserCredentials, 0)
	if _, err := os.Stat(u.credentialsFile); os.IsNotExist(err) {
		users = append(users, user)
		if err := u.writeUsers(users); err != nil {
			return err
		}
		log.Infof("Created user '%s' in %s", u.username, u.credentialsFile)
		return nil
	}
	if users, err = security.LoadCredentials(u.credentialsFile); err != nil {
		return err
	}
	users = append(users, user)
	if err := u.writeUsers(users); err != nil {
		return err
	}
	log.Infof("Added user '%s' in %s", u.username, u.credentialsFile)

	return nil
}

func (u *UserCreateOptions) writeUsers(users []models.UserCredentials) error {
	// Write to file
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	if err := os.WriteFile(u.credentialsFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

var userCommandExample = fmt.Sprintf(`
  # Create a new user
  %[1]s user --user-name admin --user-password secretpassword

  # Create a user with explicit path to credentials file
  %[1]s user -u manager -w managerpass -c /app/data/users.json

  # Generate a bcrypt hash for a password
  %[1]s user -u someuser -w complexpass -h

  # Short form with all options
  %[1]s user -u admin -w adminpass -c /custom/path/creds.json
`, ExamplePrefix())

func NewUserCommand() *cobra.Command {

	userCreateOpts := &UserCreateOptions{}

	userCommand := &cobra.Command{
		Use:     "user",
		Short:   "Create a new user",
		Example: userCommandExample,
		RunE:    userCreateOpts.Execute,
		PreRunE: userCreateOpts.Validate,
	}

	userCreateOpts.AddFlags(userCommand)

	return userCommand
}

var _ Command = (*UserCreateOptions)(nil)
