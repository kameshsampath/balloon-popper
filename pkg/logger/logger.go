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

package logger

import (
	"fmt"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

// Config holds logger configuration
type Config struct {
	Level       string
	Output      io.Writer
	WithCaller  bool
	Development bool
}

// Init initializes the global logger with the given configuration
func Init(cfg Config) error {
	logger, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	log = logger
	return nil
}

// NewLogger creates a new sugared logger instance
func NewLogger(cfg Config) (*zap.SugaredLogger, error) {
	if cfg.Level == "" {
		cfg.Level = "info"
	}

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level '%s': %w", cfg.Level, err)
	}

	var config zap.Config
	if cfg.Development {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level = zap.NewAtomicLevelAt(level)
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.OutputPaths = []string{"stdout"}
	config.DisableStacktrace = !cfg.Development

	options := []zap.Option{
		zap.AddCallerSkip(1),
	}
	if cfg.WithCaller {
		options = append(options, zap.WithCaller(true))
	}

	logger, err := config.Build(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger.Sugar(), nil
}

// Get returns the global logger instance
func Get() *zap.SugaredLogger {
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	return log
}

// Sync flushes any buffered log entries
func Sync() error {
	if log != nil {
		return log.Sync()
	}
	return nil
}
