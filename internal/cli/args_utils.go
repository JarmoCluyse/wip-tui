package cli

import (
	"flag"
	"fmt"
	"os"
)

// Define configures all available command line arguments.
func (p *Parser) Define() {
	p.flagSet.StringVar(&p.args.ConfigPath, "config", "", "Path to configuration file")
	p.flagSet.StringVar(&p.args.ConfigPath, "c", "", "Path to configuration file (shorthand)")

	p.flagSet.BoolVar(&p.args.Help, "help", false, "Show help information")
	p.flagSet.BoolVar(&p.args.Help, "h", false, "Show help information (shorthand)")

	p.flagSet.BoolVar(&p.args.Version, "version", false, "Show version information")
	p.flagSet.BoolVar(&p.args.Version, "v", false, "Show version information (shorthand)")
}

// PrintUsage adds the usage to flagSet
func PrintUsage(flagSet *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(flagSet.Output(), "Git TUI - Terminal User Interface for Git Repository Management\n\n")
		fmt.Fprintf(flagSet.Output(), "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(flagSet.Output(), "Options:\n")
		flagSet.PrintDefaults()
		fmt.Fprintf(flagSet.Output(), "\nExamples:\n")
		fmt.Fprintf(flagSet.Output(), "  %s                          # Use default config\n", os.Args[0])
		fmt.Fprintf(flagSet.Output(), "  %s -c ~/.config/git-tui.toml # Use custom config\n", os.Args[0])
		fmt.Fprintf(flagSet.Output(), "  %s --config /path/to/config.toml\n", os.Args[0])
	}
}

// ParseFromOS parses arguments from os.Args.
func (p *Parser) ParseFromOS() (*Args, error) {
	return p.Parse(os.Args[1:])
}

// GetArgs returns the current parsed arguments.
func (p *Parser) GetArgs() *Args {
	return p.args
}
