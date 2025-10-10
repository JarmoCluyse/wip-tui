package cli

import (
	"flag"
	"fmt"
	"os"
)

// Args holds all parsed command line arguments
type Args struct {
	ConfigPath string
	Help       bool
	Version    bool
}

// Parser handles CLI argument parsing
type Parser struct {
	flagSet *flag.FlagSet
	args    *Args
}

// NewParser creates a new CLI argument parser
func NewParser() *Parser {
	args := &Args{}
	flagSet := flag.NewFlagSet("git-tui", flag.ExitOnError)

	// Configure flag output
	flagSet.Usage = func() {
		fmt.Fprintf(flagSet.Output(), "Git TUI - Terminal User Interface for Git Repository Management\n\n")
		fmt.Fprintf(flagSet.Output(), "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(flagSet.Output(), "Options:\n")
		flagSet.PrintDefaults()
		fmt.Fprintf(flagSet.Output(), "\nExamples:\n")
		fmt.Fprintf(flagSet.Output(), "  %s                          # Use default config\n", os.Args[0])
		fmt.Fprintf(flagSet.Output(), "  %s -c ~/.config/git-tui.toml # Use custom config\n", os.Args[0])
		fmt.Fprintf(flagSet.Output(), "  %s --config /path/to/config.toml\n", os.Args[0])
	}

	return &Parser{
		flagSet: flagSet,
		args:    args,
	}
}

// Define configures all available command line arguments
func (p *Parser) Define() {
	p.flagSet.StringVar(&p.args.ConfigPath, "config", "", "Path to configuration file")
	p.flagSet.StringVar(&p.args.ConfigPath, "c", "", "Path to configuration file (shorthand)")
	p.flagSet.BoolVar(&p.args.Help, "help", false, "Show help information")
	p.flagSet.BoolVar(&p.args.Help, "h", false, "Show help information (shorthand)")
	p.flagSet.BoolVar(&p.args.Version, "version", false, "Show version information")
	p.flagSet.BoolVar(&p.args.Version, "v", false, "Show version information (shorthand)")
}

// Parse processes the command line arguments
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

// ParseFromOS parses arguments from os.Args
func (p *Parser) ParseFromOS() (*Args, error) {
	return p.Parse(os.Args[1:])
}

// GetArgs returns the current parsed arguments
func (p *Parser) GetArgs() *Args {
	return p.args
}
