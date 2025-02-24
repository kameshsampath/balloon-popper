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
	"github.com/kameshsampath/balloon-popper-server/pkg/web"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type ServerOptions struct {
	privateKeyFile     string
	privateKeyPassword string
	port               int
}

func (s *ServerOptions) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVarP(&s.privateKeyFile, "private-key-file", "f", "",
		"RSA private key file for JWT token signing")
	flags.StringVarP(&s.privateKeyPassword, "private-key-password", "w", "",
		"RSA private key password")
	flags.IntVarP(&s.port, "port", "p", 8080,
		"Server listening port")

	// Mark required flags
	cobra.CheckErr(cmd.MarkFlagRequired("private-key-file"))
}

func (s *ServerOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (s *ServerOptions) Execute(cmd *cobra.Command, args []string) error {

	server := web.NewServer(logger.Get())

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		if err := server.Stop(); err != nil {
			log.Printf("Error stopping server: %v", err)
		}
		os.Exit(0)
	}()

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	return nil
}

var serverCommandExample = fmt.Sprintf(`
  # Run server with unencrypted private key
  %[1]s server --private-key-file /keys/foo
  # Run server with unencrypted private key
  %[1]s server --private-key-file /keys/foo --private-key-password password123
`, ExamplePrefix())

// NewServerCommand starts the Balloon Popper Server
func NewServerCommand() *cobra.Command {
	serverOpts := &ServerOptions{}

	serverCommand := &cobra.Command{
		Use:     "server",
		Short:   "Start the Balloon Popper Server",
		Example: serverCommandExample,
		RunE:    serverOpts.Execute,
		PreRunE: serverOpts.Validate,
	}

	serverOpts.AddFlags(serverCommand)

	return serverCommand
}

var _ Command = (*ServerOptions)(nil)
