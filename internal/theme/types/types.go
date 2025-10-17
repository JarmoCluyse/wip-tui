package types

// Theme represents the complete theme configuration.
type Theme struct {
	Colors     Colors     `yaml:"colors"`
	Indicators Indicators `yaml:"indicators"`
	Icons      Icons      `yaml:"icons"`
}

// Colors defines all color values used in the UI.
type Colors struct {
	Title              string `yaml:"title"`
	TitleBackground    string `yaml:"title_background"`
	Selected           string `yaml:"selected"`
	SelectedBackground string `yaml:"selected_background"`
	StatusClean        string `yaml:"status_clean"`
	StatusDirty        string `yaml:"status_dirty"`
	StatusUnpushed     string `yaml:"status_unpushed"`
	StatusUntracked    string `yaml:"status_untracked"`
	StatusError        string `yaml:"status_error"`
	StatusNotAdded     string `yaml:"status_not_added"`
	Help               string `yaml:"help"`
	Border             string `yaml:"border"`
	ModalBackground    string `yaml:"modal_background"`
	Branch             string `yaml:"branch"`
	IconRegular        string `yaml:"icon_regular"`
	IconBare           string `yaml:"icon_bare"`
	IconWorktree       string `yaml:"icon_worktree"`
}

// Indicators defines all status indicator symbols.
type Indicators struct {
	Clean       string `yaml:"clean"`
	Dirty       string `yaml:"dirty"`
	Unpushed    string `yaml:"unpushed"`
	Untracked   string `yaml:"untracked"`
	Error       string `yaml:"error"`
	NotAdded    string `yaml:"not_added"`
	Selected    string `yaml:"selected"`
	SelectedEnd string `yaml:"selected_end"`
}

// Icons defines all icon symbols used in the UI.
type Icons struct {
	Repository struct {
		Regular  string `yaml:"regular"`
		Bare     string `yaml:"bare"`
		Worktree string `yaml:"worktree"`
	} `yaml:"repository"`
	Branch struct {
		Icon string `yaml:"icon"`
	} `yaml:"branch"`
	Tree struct {
		Branch string `yaml:"branch"`
		Last   string `yaml:"last"`
	} `yaml:"tree"`
	Folder struct {
		Icon string `yaml:"icon"`
	} `yaml:"folder"`
}
