package config

import (
	"os/exec"
	"strings"
)

// Action represents a configurable action with key binding and command.
type Action struct {
	Name        string   `yaml:"name"`        // Display name for the action
	Key         string   `yaml:"key"`         // Key binding (e.g., "l", "o", "ctrl+o")
	Command     string   `yaml:"command"`     // The command to execute
	Args        []string `yaml:"args"`        // Arguments to pass to the command
	Description string   `yaml:"description"` // Description of what this action does
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
