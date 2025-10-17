// Package theme provides theming configuration for the application.
package theme

import "github.com/jarmocluyse/git-dash/internal/theme/types"

// Default returns the default theme configuration.
// This is the authoritative source for all theme defaults.
// Config files can override individual values which will be merged with these defaults.
func Default() types.Theme {
	return types.Theme{
		Colors: types.Colors{
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
		Indicators: types.Indicators{
			Clean:       "󰄬 ",
			Dirty:       "󰏫 ",
			Unpushed:    "󰕒 ",
			Untracked:   "󰈔 ",
			Error:       " ",
			NotAdded:    "󰝒 ",
			Selected:    "󰒊 ",
			SelectedEnd: "▌",
		},
		Icons: types.Icons{
			Repository: struct {
				Regular  string `yaml:"regular"`
				Bare     string `yaml:"bare"`
				Worktree string `yaml:"worktree"`
			}{
				Regular:  "󰊢 ",
				Bare:     "󰉋 ",
				Worktree: "󰐅 ",
			},
			Branch: struct {
				Icon string `yaml:"icon"`
			}{
				Icon: "󰘬 ",
			},
			Tree: struct {
				Branch string `yaml:"branch"`
				Last   string `yaml:"last"`
			}{
				Branch: "├─",
				Last:   "└─",
			},
			Folder: struct {
				Icon string `yaml:"icon"`
			}{
				Icon: "󰉋 ",
			},
		},
	}
}

// MergeWithDefault takes a user theme and fills in any missing values with defaults.
// This allows config files to specify only the values they want to override,
// while ensuring all theme properties have valid values.
func MergeWithDefault(userTheme types.Theme) types.Theme {
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
