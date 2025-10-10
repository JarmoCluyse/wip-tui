package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize structured logging
	initLogger()

	program := tea.NewProgram(createInitialModel(), tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		log.Fatal(err)
		os.Exit(1)
	}
}
