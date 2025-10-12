package env

import "os"

// SetupTerminal sets environment variables to prevent terminal query issues.
// This ensures consistent color and terminal capabilities across different environments.
func SetupTerminal() {
	// Prevent terminal from querying capabilities that can cause escape sequences
	os.Setenv("COLORTERM", "truecolor")
	os.Setenv("TERM", "xterm-256color")
}
