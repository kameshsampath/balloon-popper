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

package security

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const (
	// Character sets for password generation
	lowerChars = "abcdefghijklmnopqrstuvwxyz"
	upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers    = "0123456789"
	symbols    = "!@#$%^&*()_+-=[]{}|;:,.<>?"
)

type PasswordConfig struct {
	Length           int
	IncludeLower     bool
	IncludeUpper     bool
	IncludeNumbers   bool
	IncludeSymbols   bool
	RequireAll       bool
	ExcludeSimilar   bool
	ExcludeAmbiguous bool
}

// NewPasswordGenerator gets a new instance of the password generator with sane defaults
func NewPasswordGenerator() *PasswordConfig {
	return &PasswordConfig{
		Length:           16,
		IncludeLower:     true,
		IncludeUpper:     true,
		IncludeNumbers:   true,
		IncludeSymbols:   true,
		RequireAll:       true,
		ExcludeSimilar:   true,
		ExcludeAmbiguous: true,
	}
}

// GeneratePassword creates a cryptographically secure random password
func (pc *PasswordConfig) GeneratePassword() (string, error) {
	// Validate configuration
	if pc.Length < 8 {
		return "", fmt.Errorf("password length must be at least 8")
	}

	// Build character set
	var chars strings.Builder
	var requiredChars strings.Builder

	if pc.IncludeLower {
		chars.WriteString(lowerChars)
		if pc.RequireAll {
			// Add one random lowercase character
			randomChar, err := getRandomChar(lowerChars)
			if err != nil {
				return "", err
			}
			requiredChars.WriteString(string(randomChar))
		}
	}

	if pc.IncludeUpper {
		chars.WriteString(upperChars)
		if pc.RequireAll {
			randomChar, err := getRandomChar(upperChars)
			if err != nil {
				return "", err
			}
			requiredChars.WriteString(string(randomChar))
		}
	}

	if pc.IncludeNumbers {
		chars.WriteString(numbers)
		if pc.RequireAll {
			randomChar, err := getRandomChar(numbers)
			if err != nil {
				return "", err
			}
			requiredChars.WriteString(string(randomChar))
		}
	}

	if pc.IncludeSymbols {
		chars.WriteString(symbols)
		if pc.RequireAll {
			randomChar, err := getRandomChar(symbols)
			if err != nil {
				return "", err
			}
			requiredChars.WriteString(string(randomChar))
		}
	}

	// Remove similar characters if requested
	if pc.ExcludeSimilar {
		chars.WriteString(strings.NewReplacer(
			"1", "", "l", "", "I", "", "0", "", "O", "", "o", "",
			"5", "", "S", "", "s", "", "8", "", "B", "",
		).Replace(chars.String()))
	}

	// Remove ambiguous characters if requested
	//TODO: improve it to exclude specific characters
	if pc.ExcludeAmbiguous {
		chars.WriteString(strings.NewReplacer(
			"`", "", "~", "", "^", "", ",", "", ".", "", ";", "", ":", "",
			"<", "", ">", "", "[", "", "]", "", "{", "", "}", "", "(", "",
			")", "", "/", "", "\\", "", "|", "",
		).Replace(chars.String()))
	}

	if chars.Len() == 0 {
		return "", fmt.Errorf("no characters available with current configuration")
	}

	// Calculate remaining length after required characters
	remainingLength := pc.Length - requiredChars.Len()
	if remainingLength < 0 {
		return "", fmt.Errorf("password length too short for required characters")
	}

	// Generate random characters for remaining length
	result := requiredChars.String()
	charset := chars.String()
	for i := 0; i < remainingLength; i++ {
		randomChar, err := getRandomChar(charset)
		if err != nil {
			return "", fmt.Errorf("failed to generate random character: %v", err)
		}
		result += string(randomChar)
	}

	// Shuffle the final password
	shuffled, err := shuffleString(result)
	if err != nil {
		return "", fmt.Errorf("failed to shuffle password: %v", err)
	}

	return shuffled, nil
}

// getRandomChar returns a random character from the given string
func getRandomChar(chars string) (byte, error) {
	if len(chars) == 0 {
		return 0, fmt.Errorf("empty character set")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
	if err != nil {
		return 0, err
	}
	return chars[n.Int64()], nil
}

// shuffleString randomly shuffles a string
func shuffleString(s string) (string, error) {
	runes := []rune(s)
	for i := len(runes) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		j := n.Int64()
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes), nil
}
