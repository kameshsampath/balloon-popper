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
func (pc *PasswordConfig) GeneratePassword() ([]byte, error) {
	// Validate configuration
	if pc.Length < 8 {
		return nil, fmt.Errorf("password length must be at least 8")
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
				return nil, err
			}
			requiredChars.WriteString(string(randomChar))
		}
	}

	if pc.IncludeUpper {
		chars.WriteString(upperChars)
		if pc.RequireAll {
			randomChar, err := getRandomChar(upperChars)
			if err != nil {
				return nil, err
			}
			requiredChars.WriteString(string(randomChar))
		}
	}

	if pc.IncludeNumbers {
		chars.WriteString(numbers)
		if pc.RequireAll {
			randomChar, err := getRandomChar(numbers)
			if err != nil {
				return nil, err
			}
			requiredChars.WriteString(string(randomChar))
		}
	}

	if pc.IncludeSymbols {
		chars.WriteString(symbols)
		if pc.RequireAll {
			randomChar, err := getRandomChar(symbols)
			if err != nil {
				return nil, err
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
		return nil, fmt.Errorf("no characters available with current configuration")
	}

	// Calculate remaining length after required characters
	remainingLength := pc.Length - requiredChars.Len()
	if remainingLength < 0 {
		return nil, fmt.Errorf("password length too short for required characters")
	}

	// Generate random characters for remaining length
	result := requiredChars.String()
	charset := chars.String()
	for i := 0; i < remainingLength; i++ {
		randomChar, err := getRandomChar(charset)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random character: %v", err)
		}
		result += string(randomChar)
	}

	// Shuffle the final password
	shuffled, err := shuffleToBytes(result)
	if err != nil {
		return nil, fmt.Errorf("failed to shuffle password: %v", err)
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

// shuffleToBytes randomly shuffles a string
func shuffleToBytes(s string) ([]byte, error) {
	b := []byte(s)
	for i := len(b) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return nil, err
		}
		j := n.Int64()
		b[i], b[j] = b[j], b[i]
	}
	return b, nil
}
