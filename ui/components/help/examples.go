package help

// Example of how to use the detailed help functionality
// This can be used when implementing a help modal (triggered by '?' key)

/*
Example usage for detailed help modal:

helpBuilder := help.NewBuilder(styles.Help)

sections := []help.HelpSection{
	{
		Title: "Navigation",
		Bindings: []help.KeyBinding{
			{"↑", "move up"},
			{"↓", "move down"},
			{"k", "move up (vim style)"},
			{"j", "move down (vim style)"},
			{"Enter", "select item"},
		},
	},
	{
		Title: "Actions",
		Bindings: []help.KeyBinding{
			{"l", "open in Lazygit"},
			{"c", "open in VS Code"},
			{"m", "manage repositories"},
		},
	},
	{
		Title: "General",
		Bindings: []help.KeyBinding{
			{"Esc", "go back"},
			{"q", "quit application"},
			{"?", "show/hide this help"},
		},
	},
}

detailedHelp := helpBuilder.BuildDetailedHelp(sections)
*/
