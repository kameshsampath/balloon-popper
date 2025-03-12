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
	"github.com/spf13/cobra"
	"os"
)

var verbosity string

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "balloon-popper",
		Short: "A balloon popping game",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize logger
			err := logger.Init(logger.Config{
				Level:       verbosity,
				Output:      os.Stdout,
				WithCaller:  true,
				Development: true,
			})
			if err != nil {
				return fmt.Errorf("failed to initialize logger: %w", err)
			}

			logger.Get().Debugf("Logger initialized with level: %s", verbosity)
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			_ = logger.Sync()
		},
	}

	rootCmd.PersistentFlags().StringVarP(&verbosity, "verbose", "v", "info",
		"Log level (debug, info, warn, error)")

	rootCmd.AddCommand(NewVersionCommand())
	rootCmd.AddCommand(NewServerCommand())
	rootCmd.AddCommand(NewJWTKeysCommand())
	rootCmd.AddCommand(NewUserCommand())

	return rootCmd
}
