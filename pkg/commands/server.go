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
	"github.com/kameshsampath/balloon-popper/pkg/producer"
	"github.com/kameshsampath/balloon-popper/pkg/routes"
	"github.com/kameshsampath/balloon-popper/pkg/security"
	"github.com/kameshsampath/balloon-popper/pkg/web"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

var appLogger = logger.Get()

type ServerOptions struct {
	privateKeyFile        string
	privateKeyPassword    string
	kafkaBootstrapServers string
	kafkaTopic            string
	port                  int
	userCredentialsFile   string
	verbose               bool
}

func (s *ServerOptions) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	// Add flags
	flags.StringVarP(&s.privateKeyFile, "key-file", "k", "", "Path to the private key file")
	flags.StringVarP(&s.privateKeyPassword, "key-password", "p", "", "Password for the private key")
	flags.StringVarP(&s.kafkaBootstrapServers, "kafka-servers", "s", "localhost:19094", "Kafka bootstrap servers")
	flags.StringVarP(&s.kafkaTopic, "kafka-topic", "t", "balloon-game", "Kafka topic to send balloon game scores")
	flags.IntVarP(&s.port, "port", "P", 8080, "Server port")
	flags.StringVarP(&s.userCredentialsFile, "credentials-file", "c", "", "Path to user credentials file")
	flags.BoolVarP(&s.verbose, "verbose", "v", false, "Enable verbose mode")

	// Mark required flags
	err := cmd.MarkFlagRequired("key-file")
	if err != nil {
		appLogger.Fatal(err)
	}
	err = cmd.MarkFlagRequired("credentials-file")
	if err != nil {
		appLogger.Fatal(err)
	}
}

func (s *ServerOptions) Validate(_ *cobra.Command, _ []string) error {
	return nil
}

func (s *ServerOptions) Execute(_ *cobra.Command, _ []string) error {
	var err error
	logLevel := "info"
	lc := logger.Config{
		Level:  logLevel,
		Output: os.Stdout,
	}
	if s.verbose {
		lc.Level = "debug"
		lc.WithCaller = true
		lc.Development = true
	}
	if appLogger, err = logger.NewLogger(lc); err != nil {
		appLogger.Warnf("Unable to initialize logger: %v.Using defaults.", err)
	}
	//check to see if the private key file exists
	if _, err := os.Stat(s.privateKeyFile); err != nil {
		return fmt.Errorf("Error: error loading %s  private key file: %v", s.privateKeyFile, err)
	}
	// create endpoint with JWT config
	ec, err := routes.NewEndpoints(s.privateKeyFile, s.privateKeyPassword)
	if err != nil {
		return err
	}
	ec.Logger = appLogger
	//Load Users
	if c, err := security.LoadCredentials(s.userCredentialsFile); err != nil {
		return err
	} else {
		ec.Users = c
	}
	// Initialize Kafka kafkaScoreProducer
	kp, err := producer.NewKafkaScoreProducer(s.kafkaBootstrapServers, s.kafkaTopic)
	if err != nil {
		return err
	}
	ec.KafkaProducer = kp
	// Start Kafka producer
	if err := ec.KafkaProducer.Start(); err != nil {
		return fmt.Errorf("failed to start Kafka producer: %v", err)
	}
	//Create a new Server
	server := web.NewServer(appLogger, s.port, ec)
	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		if err := server.Stop(); err != nil {
			appLogger.Errorf("Error stopping server: %v", err)
		}
		os.Exit(0)
	}()
	//Start the server
	if err := server.Start(); err != nil {
		appLogger.Fatalf("Failed to start server: %v", err)
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
