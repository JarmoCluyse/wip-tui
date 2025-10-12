// Package theme provides theming configuration for the application.
package theme

// Theme represents the complete theme configuration.
type Theme struct {
	Colors     Colors     `toml:"colors"`
	Indicators Indicators `toml:"indicators"`
	Icons      Icons      `toml:"icons"`
}

// Colors defines all color values used in the UI.
type Colors struct {
	Title              string `toml:"title"`
	TitleBackground    string `toml:"title_background"`
	Selected           string `toml:"selected"`
	SelectedBackground string `toml:"selected_background"`
	StatusClean        string `toml:"status_clean"`
	StatusDirty        string `toml:"status_dirty"`
	StatusUnpushed     string `toml:"status_unpushed"`
	StatusUntracked    string `toml:"status_untracked"`
	StatusError        string `toml:"status_error"`
	StatusNotAdded     string `toml:"status_not_added"`
	Help               string `toml:"help"`
	Border             string `toml:"border"`
	ModalBackground    string `toml:"modal_background"`
	Branch             string `toml:"branch"`
	IconRegular        string `toml:"icon_regular"`
	IconBare           string `toml:"icon_bare"`
	IconWorktree       string `toml:"icon_worktree"`
}

// Indicators defines all status indicator symbols.
type Indicators struct {
	Clean       string `toml:"clean"`
	Dirty       string `toml:"dirty"`
	Unpushed    string `toml:"unpushed"`
	Untracked   string `toml:"untracked"`
	Error       string `toml:"error"`
	NotAdded    string `toml:"not_added"`
	Selected    string `toml:"selected"`
	SelectedEnd string `toml:"selected_end"`
}

// Icons defines all icon symbols used in the UI.
type Icons struct {
	Repository struct {
		Regular  string `toml:"regular"`
		Bare     string `toml:"bare"`
		Worktree string `toml:"worktree"`
	} `toml:"repository"`
	Branch struct {
		Icon string `toml:"icon"`
	} `toml:"branch"`
	Tree struct {
		Branch string `toml:"branch"`
		Last   string `toml:"last"`
	} `toml:"tree"`
	Folder struct {
		Icon string `toml:"icon"`
	} `toml:"folder"`
}

// Default returns the default theme configuration.
func Default() Theme {
	return Theme{
		Colors: Colors{
			Title:              "#FAFAFA",
			TitleBackground:    "#7D56F4",
			Selected:           "#7D56F4",
			SelectedBackground: "#7D56F4FF",
			StatusClean:        "#6BCF7F",
			StatusDirty:        "#FF6B6B",
			StatusUnpushed:     "#FFD93D",
			StatusUntracked:    "#FFA500",
			StatusError:        "#FF0000",
			StatusNotAdded:     "#626262",
			Help:               "#626262",
			Border:             "#7D56F4",
			ModalBackground:    "#1E1E1E",
			Branch:             "#00D4AA",
			IconRegular:        "#4A9EFF",
			IconBare:           "#FFA500",
			IconWorktree:       "#32CD32",
		},
		Indicators: Indicators{
			Clean:       "󰄬 ",
			Dirty:       "󰏫 ",
			Unpushed:    "󰕒 ",
			Untracked:   "󰈔 ",
			Error:       " ",
			NotAdded:    "󰝒 ",
			Selected:    "󰒊 ",
			SelectedEnd: "▌",
		},
		Icons: Icons{
			Repository: struct {
				Regular  string `toml:"regular"`
				Bare     string `toml:"bare"`
				Worktree string `toml:"worktree"`
			}{
				Regular:  "󰊢 ",
				Bare:     "󰉋 ",
				Worktree: "󰐅 ",
			},
			Branch: struct {
				Icon string `toml:"icon"`
			}{
				Icon: "󰘬 ",
			},
			Tree: struct {
				Branch string `toml:"branch"`
				Last   string `toml:"last"`
			}{
				Branch: "├─",
				Last:   "└─",
			},
			Folder: struct {
				Icon string `toml:"icon"`
			}{
				Icon: "󰉋 ",
			},
		},
	}
}

// MergeWithDefault takes a user theme and fills in any missing values with defaults.
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
	if userTheme.Colors.SelectedBackground == "" {
		userTheme.Colors.SelectedBackground = defaultTheme.Colors.SelectedBackground
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
	if userTheme.Colors.IconRegular == "" {
		userTheme.Colors.IconRegular = defaultTheme.Colors.IconRegular
	}
	if userTheme.Colors.IconBare == "" {
		userTheme.Colors.IconBare = defaultTheme.Colors.IconBare
	}
	if userTheme.Colors.IconWorktree == "" {
		userTheme.Colors.IconWorktree = defaultTheme.Colors.IconWorktree
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
	if userTheme.Indicators.Selected == "" {
		userTheme.Indicators.Selected = defaultTheme.Indicators.Selected
	}
	if userTheme.Indicators.SelectedEnd == "" {
		userTheme.Indicators.SelectedEnd = defaultTheme.Indicators.SelectedEnd
	}

	// Merge icons
	if userTheme.Icons.Repository.Regular == "" {
		userTheme.Icons.Repository.Regular = defaultTheme.Icons.Repository.Regular
	}
	if userTheme.Icons.Repository.Bare == "" {
		userTheme.Icons.Repository.Bare = defaultTheme.Icons.Repository.Bare
	}
	if userTheme.Icons.Repository.Worktree == "" {
		userTheme.Icons.Repository.Worktree = defaultTheme.Icons.Repository.Worktree
	}
	if userTheme.Icons.Branch.Icon == "" {
		userTheme.Icons.Branch.Icon = defaultTheme.Icons.Branch.Icon
	}
	if userTheme.Icons.Tree.Branch == "" {
		userTheme.Icons.Tree.Branch = defaultTheme.Icons.Tree.Branch
	}
	if userTheme.Icons.Tree.Last == "" {
		userTheme.Icons.Tree.Last = defaultTheme.Icons.Tree.Last
	}
	if userTheme.Icons.Folder.Icon == "" {
		userTheme.Icons.Folder.Icon = defaultTheme.Icons.Folder.Icon
	}

	return userTheme
}
