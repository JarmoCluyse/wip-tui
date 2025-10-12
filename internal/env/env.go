// Package env provides environment variable loading and terminal setup utilities.
package env

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnvFile loads environment variables from .env and .env.development files.
// The .env.development file values override .env values if both exist.
func LoadEnvFile() {
	// Load base .env file first
	loadEnvFromFile(".env")

	// Load .env.development file (overrides .env values)
	loadEnvFromFile(".env.development")
}

// loadEnvFromFile loads environment variables from a specific file.
// Env files are optional - missing files are silently ignored.
func loadEnvFromFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		// Env files are optional, so just return if they don't exist
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Set environment variable (will override if already exists)
			os.Setenv(key, value)
		}
	}
}

// SetupTerminal sets environment variables to prevent terminal query issues.
// This ensures consistent color and terminal capabilities across different environments.
func SetupTerminal() {
	// Prevent terminal from querying capabilities that can cause escape sequences
	os.Setenv("COLORTERM", "truecolor")
	os.Setenv("TERM", "xterm-256color")
}
