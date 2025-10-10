package theme

// Theme represents the complete theme configuration
type Theme struct {
	Colors     Colors     `toml:"colors"`
	Indicators Indicators `toml:"indicators"`
}

// Colors defines all color values used in the UI
type Colors struct {
	Title           string `toml:"title"`
	TitleBackground string `toml:"title_background"`
	Selected        string `toml:"selected"`
	StatusClean     string `toml:"status_clean"`
	StatusDirty     string `toml:"status_dirty"`
	StatusUnpushed  string `toml:"status_unpushed"`
	StatusUntracked string `toml:"status_untracked"`
	StatusError     string `toml:"status_error"`
	StatusNotAdded  string `toml:"status_not_added"`
	Help            string `toml:"help"`
	Border          string `toml:"border"`
	ModalBackground string `toml:"modal_background"`
	Branch          string `toml:"branch"`
}

// Indicators defines all status indicator symbols
type Indicators struct {
	Clean     string `toml:"clean"`
	Dirty     string `toml:"dirty"`
	Unpushed  string `toml:"unpushed"`
	Untracked string `toml:"untracked"`
	Error     string `toml:"error"`
	NotAdded  string `toml:"not_added"`
}

// Default returns the default theme configuration
func Default() Theme {
	return Theme{
		Colors: Colors{
			Title:           "#FAFAFA",
			TitleBackground: "#7D56F4",
			Selected:        "#7D56F4",
			StatusClean:     "#6BCF7F",
			StatusDirty:     "#FF6B6B",
			StatusUnpushed:  "#FFD93D",
			StatusUntracked: "#FFA500",
			StatusError:     "#FF0000",
			StatusNotAdded:  "#626262",
			Help:            "#626262",
			Border:          "#7D56F4",
			ModalBackground: "#1E1E1E",
			Branch:          "#00D4AA",
		},
		Indicators: Indicators{
			Clean:     "🟢",
			Dirty:     "🔴",
			Unpushed:  "🟡",
			Untracked: "🟠",
			Error:     "❌",
			NotAdded:  "⚪",
		},
	}
}

// MergeWithDefault takes a user theme and fills in any missing values with defaults
func MergeWithDefault(userTheme Theme) Theme {
	defaultTheme := Default()

	// Merge colors - use user value if not empty, otherwise use default
	if userTheme.Colors.Title == "" {
		userTheme.Colors.Title = defaultTheme.Colors.Title
	}
	if userTheme.Colors.TitleBackground == "" {
		userTheme.Colors.TitleBackground = defaultTheme.Colors.TitleBackground
	}
	if userTheme.Colors.Selected == "" {
		userTheme.Colors.Selected = defaultTheme.Colors.Selected
	}
	if userTheme.Colors.StatusClean == "" {
		userTheme.Colors.StatusClean = defaultTheme.Colors.StatusClean
	}
	if userTheme.Colors.StatusDirty == "" {
		userTheme.Colors.StatusDirty = defaultTheme.Colors.StatusDirty
	}
	if userTheme.Colors.StatusUnpushed == "" {
		userTheme.Colors.StatusUnpushed = defaultTheme.Colors.StatusUnpushed
	}
	if userTheme.Colors.StatusUntracked == "" {
		userTheme.Colors.StatusUntracked = defaultTheme.Colors.StatusUntracked
	}
	if userTheme.Colors.StatusError == "" {
		userTheme.Colors.StatusError = defaultTheme.Colors.StatusError
	}
	if userTheme.Colors.StatusNotAdded == "" {
		userTheme.Colors.StatusNotAdded = defaultTheme.Colors.StatusNotAdded
	}
	if userTheme.Colors.Help == "" {
		userTheme.Colors.Help = defaultTheme.Colors.Help
	}
	if userTheme.Colors.Border == "" {
		userTheme.Colors.Border = defaultTheme.Colors.Border
	}
	if userTheme.Colors.ModalBackground == "" {
		userTheme.Colors.ModalBackground = defaultTheme.Colors.ModalBackground
	}
	if userTheme.Colors.Branch == "" {
		userTheme.Colors.Branch = defaultTheme.Colors.Branch
	}

	// Merge indicators
	if userTheme.Indicators.Clean == "" {
		userTheme.Indicators.Clean = defaultTheme.Indicators.Clean
	}
	if userTheme.Indicators.Dirty == "" {
		userTheme.Indicators.Dirty = defaultTheme.Indicators.Dirty
	}
	if userTheme.Indicators.Unpushed == "" {
		userTheme.Indicators.Unpushed = defaultTheme.Indicators.Unpushed
	}
	if userTheme.Indicators.Untracked == "" {
		userTheme.Indicators.Untracked = defaultTheme.Indicators.Untracked
	}
	if userTheme.Indicators.Error == "" {
		userTheme.Indicators.Error = defaultTheme.Indicators.Error
	}
	if userTheme.Indicators.NotAdded == "" {
		userTheme.Indicators.NotAdded = defaultTheme.Indicators.NotAdded
	}

	return userTheme
}
