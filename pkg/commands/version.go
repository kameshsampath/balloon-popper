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

	"github.com/spf13/cobra"
)

// These are populated at link time, see ./hack/build-flags.sh
var (
	// Version is the version string at which the CLI is built.
	Version string
	// BuildDate is the date on which this CLI binary was built
	BuildDate string
	// Commit is the git commit from which this CLI binary was built.
	Commit string
	// BuiltBy is the release program that built this binary
	BuiltBy string
)

// NewVersionCommand implements 'balloon-popper version' commands
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the plugin version",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "Version:      %s\n", Version)
			_, _ = fmt.Fprintf(out, "Build Date:   %s\n", BuildDate)
			_, _ = fmt.Fprintf(out, "Git Revision: %s\n", Commit)
			_, _ = fmt.Fprintf(out, "Built-By: %s\n", BuiltBy)
			return nil
		},
	}
}
