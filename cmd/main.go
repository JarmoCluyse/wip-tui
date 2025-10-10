package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jarmocluyse/wip-tui/internal/cli"
	"github.com/jarmocluyse/wip-tui/internal/env"
	"github.com/jarmocluyse/wip-tui/internal/logging"
	"github.com/jarmocluyse/wip-tui/internal/ui"
)

func main() {
	// Parse command line arguments
	args, err := cli.ParseArgs()
	if err != nil {
		fmt.Printf("Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	env.SetupTerminal()
	env.LoadEnvFile()
	logging.Init()

	deps := NewAppDependencies(args.ConfigPath)
	model := ui.CreateInitialModel(deps)
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		log.Fatal(err)
		os.Exit(1)
	}
}
