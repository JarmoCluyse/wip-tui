// Package main provides the entry point for the git-dash application.
package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jarmocluyse/git-dash/internal/cli"
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/env"
	"github.com/jarmocluyse/git-dash/internal/logging"
	"github.com/jarmocluyse/git-dash/internal/repomanager"
	themeService "github.com/jarmocluyse/git-dash/internal/services/theme"
	"github.com/jarmocluyse/git-dash/ui"
)

// main initializes and runs the git-dash application.
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

	var configService config.ConfigService
	if args.ConfigPath != "" {
		configService = config.NewFileConfigServiceWithPath(args.ConfigPath)
	} else {
		configService = config.NewFileConfigService()
	}

	// Create services
	repoManager := repomanager.NewRepoManager(configService)
	themeManager := themeService.NewManager(configService)

	model := ui.CreateInitialModel(deps)
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		log.Fatal(err)
		os.Exit(1)
	}
}
