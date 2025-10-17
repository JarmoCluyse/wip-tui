package direxplorer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/git-dash/internal/theme/types"
)

// DirItem represents a directory or file in the explorer
type DirItem struct {
	Name           string
	Path           string
	IsDir          bool
	IsGitRepo      bool
	IsBareRepo     bool
	IsAlreadyAdded bool
	Size           int64
}

// Explorer handles directory exploration functionality
type Explorer struct {
	currentPath    string
	items          []DirItem
	cursor         int
	styles         StyleConfig
	theme          types.Theme
	addedRepoPaths map[string]bool // Track which repositories are already added
}

// StyleConfig holds styling configuration for the directory explorer
type StyleConfig struct {
	Directory    lipgloss.Style
	File         lipgloss.Style
	GitRepo      lipgloss.Style
	BareRepo     lipgloss.Style
	AlreadyAdded lipgloss.Style
	Selected     lipgloss.Style
	CurrentPath  lipgloss.Style
	EmptyState   lipgloss.Style
}

// NewExplorer creates a new directory explorer
func NewExplorer(startPath string, styles StyleConfig, themeConfig types.Theme) *Explorer {
	explorer := &Explorer{
		currentPath:    startPath,
		cursor:         0,
		styles:         styles,
		theme:          themeConfig,
		addedRepoPaths: make(map[string]bool),
	}

	// Load initial directory
	explorer.LoadDirectory(startPath)

	return explorer
}

// UpdateAddedRepositories updates the list of already-added repository paths
func (e *Explorer) UpdateAddedRepositories(repoPaths []string) {
	e.addedRepoPaths = make(map[string]bool)
	for _, path := range repoPaths {
		e.addedRepoPaths[path] = true
	}
	// Preserve cursor position when reloading
	currentCursor := e.cursor
	e.LoadDirectory(e.currentPath)
	// Restore cursor position if it's still valid
	if currentCursor < len(e.items) {
		e.cursor = currentCursor
	}
}

// LoadDirectory loads the contents of the specified directory
func (e *Explorer) LoadDirectory(path string) error {
	// Clean and validate path
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Check if directory exists and is accessible
	info, err := os.Stat(cleanPath)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", cleanPath)
	}

	// Read directory contents
	entries, err := os.ReadDir(cleanPath)
	if err != nil {
		return err
	}

	// Clear existing items and reset cursor
	e.items = []DirItem{}
	e.cursor = 0
	e.currentPath = cleanPath

	// Add parent directory entry (unless we're at root)
	if cleanPath != filepath.Dir(cleanPath) {
		e.items = append(e.items, DirItem{
			Name:  "..",
			Path:  filepath.Dir(cleanPath),
			IsDir: true,
		})
	}

	// Process directory entries
	for _, entry := range entries {
		// Skip hidden files/directories (starting with .)
		if strings.HasPrefix(entry.Name(), ".") && entry.Name() != ".." {
			continue
		}

		entryPath := filepath.Join(cleanPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't stat
		}

		item := DirItem{
			Name:  entry.Name(),
			Path:  entryPath,
			IsDir: info.IsDir(),
			Size:  info.Size(),
		}

		// Check if directory is a git repository
		if item.IsDir {
			gitPath := filepath.Join(entryPath, ".git")
			headsPath := filepath.Join(entryPath, "HEAD")

			// Check for normal git repository (.git directory)
			if _, err := os.Stat(gitPath); err == nil {
				item.IsGitRepo = true
			} else if _, err := os.Stat(headsPath); err == nil {
				// Check for bare repository (has HEAD file but no .git directory)
				refsPath := filepath.Join(entryPath, "refs")
				objectsPath := filepath.Join(entryPath, "objects")
				if _, refErr := os.Stat(refsPath); refErr == nil {
					if _, objErr := os.Stat(objectsPath); objErr == nil {
						item.IsGitRepo = true
						item.IsBareRepo = true
					}
				}
			}

			// Check if this repository is already added
			if item.IsGitRepo {
				item.IsAlreadyAdded = e.addedRepoPaths[entryPath]
			}
		}

		e.items = append(e.items, item)
	}

	// Sort items: directories first, then files, alphabetically within each group
	sort.Slice(e.items, func(i, j int) bool {
		// Keep ".." at the top
		if e.items[i].Name == ".." {
			return true
		}
		if e.items[j].Name == ".." {
			return false
		}

		// Directories before files
		if e.items[i].IsDir != e.items[j].IsDir {
			return e.items[i].IsDir
		}

		// Alphabetical within same type
		return strings.ToLower(e.items[i].Name) < strings.ToLower(e.items[j].Name)
	})

	return nil
}

// GetCurrentPath returns the current directory path
func (e *Explorer) GetCurrentPath() string {
	return e.currentPath
}

// GetSelectedItem returns the currently selected item
func (e *Explorer) GetSelectedItem() *DirItem {
	if e.cursor >= 0 && e.cursor < len(e.items) {
		return &e.items[e.cursor]
	}
	return nil
}

// MoveCursorUp moves the cursor up
func (e *Explorer) MoveCursorUp() {
	if e.cursor > 0 {
		e.cursor--
	}
}

// MoveCursorDown moves the cursor down
func (e *Explorer) MoveCursorDown() {
	if e.cursor < len(e.items)-1 {
		e.cursor++
	}
}

// NavigateInto navigates into the selected directory
func (e *Explorer) NavigateInto() error {
	selected := e.GetSelectedItem()
	if selected != nil && selected.IsDir {
		return e.LoadDirectory(selected.Path)
	}
	return fmt.Errorf("selected item is not a directory")
}

// Render renders the directory explorer
func (e *Explorer) Render(width, height int) string {
	var content strings.Builder

	// Current path header
	pathStyle := e.styles.CurrentPath
	content.WriteString(pathStyle.Render(fmt.Sprintf("📁 %s", e.currentPath)) + "\n\n")

	// Check if we have items
	if len(e.items) == 0 {
		content.WriteString(e.styles.EmptyState.Render("Directory is empty or inaccessible") + "\n")
		return content.String()
	}

	// Calculate visible items based on height
	maxVisible := height - 3 // Reserve space for header and padding
	if maxVisible < 1 {
		maxVisible = 10
	}

	// Calculate scroll offset
	scrollOffset := 0
	if e.cursor >= maxVisible {
		scrollOffset = e.cursor - maxVisible + 1
	}

	// Render visible items
	for i := scrollOffset; i < len(e.items) && i < scrollOffset+maxVisible; i++ {
		item := e.items[i]
		isSelected := i == e.cursor

		// Choose style based on item type and selection
		var style lipgloss.Style
		var icon string

		if isSelected {
			style = e.styles.Selected
			icon = e.theme.Indicators.Selected
		} else {
			icon = strings.Repeat(" ", lipgloss.Width(e.theme.Indicators.Selected))
		}

		// Add type-specific styling and icons
		var typeIcon string
		var nameStyle lipgloss.Style
		var statusIndicator string

		if item.Name == ".." {
			typeIcon = "↑"
			nameStyle = e.styles.Directory
		} else if item.IsGitRepo {
			// Choose icon based on repository type
			if item.IsBareRepo {
				typeIcon = e.theme.Icons.Repository.Bare
				nameStyle = e.styles.BareRepo
			} else {
				typeIcon = e.theme.Icons.Repository.Regular
				nameStyle = e.styles.GitRepo
			}

			// Add status indicator for already-added repositories
			if item.IsAlreadyAdded {
				statusIndicator = "✓"
				nameStyle = e.styles.AlreadyAdded
			}
		} else if item.IsDir {
			typeIcon = e.theme.Icons.Folder.Icon
			nameStyle = e.styles.Directory
		} else {
			typeIcon = "·"
			nameStyle = e.styles.File
		}

		// Format the line with status indicator
		var line string
		if statusIndicator != "" {
			line = fmt.Sprintf(" %s%s %s %s", icon, typeIcon, item.Name, statusIndicator)
		} else {
			line = fmt.Sprintf(" %s%s %s", icon, typeIcon, item.Name)
		}

		if isSelected {
			content.WriteString(style.Render(line) + "\n")
		} else {
			content.WriteString(nameStyle.Render(line) + "\n")
		}
	}

	// Show scroll indicator if needed
	if len(e.items) > maxVisible {
		totalItems := len(e.items)
		visibleStart := scrollOffset + 1
		visibleEnd := min(scrollOffset+maxVisible, totalItems)

		scrollInfo := fmt.Sprintf("(%d-%d of %d)", visibleStart, visibleEnd, totalItems)
		content.WriteString("\n" + e.styles.EmptyState.Render(scrollInfo))
	}

	return content.String()
}

// GetGitRepositories returns a list of git repositories found in the current view
func (e *Explorer) GetGitRepositories() []DirItem {
	var repos []DirItem
	for _, item := range e.items {
		if item.IsGitRepo {
			repos = append(repos, item)
		}
	}
	return repos
}

// GetItemCount returns the total number of items
func (e *Explorer) GetItemCount() int {
	return len(e.items)
}
