// Package cli provides command line argument parsing functionality.
package cli

import (
	"flag"
	"fmt"
	"os"
)

// Args holds all parsed command line arguments.
type Args struct {
	ConfigPath string // comand line config file to load
	Help       bool   // show help
	Version    bool   // show the version
}

// Parser handles CLI argument parsing.
type Parser struct {
	flagSet *flag.FlagSet
	args    *Args
}

// NewParser creates a new CLI argument parser.
func NewParser() *Parser {
	args := &Args{}
	flagSet := flag.NewFlagSet("git-tui", flag.ExitOnError)

	flagSet.Usage = PrintUsage(flagSet)

	return &Parser{
		flagSet: flagSet,
		args:    args,
	}
}

// Parse processes the command line arguments.
func (p *Parser) Parse(args []string) (*Args, error) {
	if err := p.flagSet.Parse(args); err != nil {
		return nil, err
	}

	// Handle special flags
	if p.args.Help {
		p.flagSet.Usage()
		os.Exit(0)
	}

	if p.args.Version {
		fmt.Println("git-tui version 1.0.0")
		os.Exit(0)
	}

	return p.args, nil
}
