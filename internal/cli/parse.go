package cli

// ParseArgs is a convenience function to parse command line arguments
func ParseArgs() (*Args, error) {
	parser := NewParser()
	parser.Define()
	return parser.ParseFromOS()
}
