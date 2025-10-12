package config

import (
	"os/exec"
	"strings"
)

// Action represents a configurable action with key binding and command.
type Action struct {
	Name        string   `toml:"name"`        // Display name for the action
	Key         string   `toml:"key"`         // Key binding (e.g., "l", "o", "ctrl+o")
	Command     string   `toml:"command"`     // The command to execute
	Args        []string `toml:"args"`        // Arguments to pass to the command
	Description string   `toml:"description"` // Description of what this action does
}

// ExecuteOpenAction executes the configured action with the given path.
func (a *Action) ExecuteOpenAction(path string) *exec.Cmd {
	// Replace {path} placeholder in command and args
	command := strings.ReplaceAll(a.Command, "{path}", path)

	var args []string
	for _, arg := range a.Args {
		args = append(args, strings.ReplaceAll(arg, "{path}", path))
	}

	return exec.Command(command, args...)
}
