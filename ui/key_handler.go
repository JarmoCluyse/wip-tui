package ui

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/logging"
	"github.com/jarmocluyse/git-dash/internal/theme"
	"github.com/jarmocluyse/git-dash/ui/components/direxplorer"
)

// KeyHandler manages keyboard input handling for different view states.
type KeyHandler struct {
	navigationHandler *NavigationHandler
	repositoryHandler *RepositoryOperationHandler
}

// NewKeyHandler creates a new KeyHandler instance.
func NewKeyHandler() *KeyHandler {
	return &KeyHandler{
		navigationHandler: NewNavigationHandler(),
		repositoryHandler: NewRepositoryOperationHandler(),
	}
}

// HandleKeyPress dispatches key events to appropriate handlers based on current state.
func (h *KeyHandler) HandleKeyPress(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Log current state and key press
	stateNames := []string{"ListView", "SettingsView", "DetailsView", "ActionConfigView"}
	stateName := "Unknown"
	if int(m.State) < len(stateNames) {
		stateName = stateNames[m.State]
	}
	logging.Get().Debug("key pressed", "key", msg.String(), "state", stateName)

	// Global help modal toggle
	if msg.String() == "?" {
		m.ShowHelpModal = !m.ShowHelpModal
		return m, nil
	}

	// If help modal is open, handle its keys
	if m.ShowHelpModal {
		return h.handleHelpModalKeys(m, msg)
	}

	switch m.State {
	case ListView:
		return h.handleListViewKeys(m, msg)
	case SettingsView:
		return h.handleSettingsViewKeys(m, msg)
	case DetailsView:
		return h.handleDetailsViewKeys(m, msg)
	case ActionConfigView:
		return h.handleActionConfigViewKeys(m, msg)
	default:
		return m, nil
	}
}

// handleHelpModalKeys handles keyboard input when the help modal is open.
func (h *KeyHandler) handleHelpModalKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "?":
		m.ShowHelpModal = false
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// handleListViewKeys handles key events in the main list view.
func (h *KeyHandler) handleListViewKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	switch keyStr {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		return h.navigationHandler.MoveCursorUp(m), nil
	case "down", "j":
		return h.navigationHandler.MoveCursorDown(m), nil
	case "enter":
		// Navigate to selected repository details
		navigableItems := m.getNavigableItems()
		if m.Cursor < len(navigableItems) {
			selectedItem := navigableItems[m.Cursor]
			if selectedItem.Type == "repository" || selectedItem.Type == "worktree" {
				m.State = DetailsView
				m.SelectedNavItem = &selectedItem
			}
		}
		return m, nil
	case "s":
		return h.enterSettingsMode(m), nil
	case "w":
		return h.discoverWorktrees(m)
	case "r":
		return m, m.updateRepositoryStatuses()
	case "?":
		return h.toggleHelpModal(m), nil
	default:
		// Check for configurable actions
		return h.executeConfiguredAction(m, keyStr)
	}
}

// handleSettingsViewKeys handles key events in settings view.
func (h *KeyHandler) handleSettingsViewKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	// If we're in theme edit mode, handle those keys differently
	if m.ThemeEditMode {
		return h.handleThemeEditKeys(m, msg)
	}

	// If we're in action edit mode, handle those keys differently
	if m.ActionEditMode {
		return h.handleActionEditKeysInSettings(m, msg)
	}

	// If we're in repository section and dealing with paste mode
	if (m.SettingsSection == "repositories" || m.SettingsSection == "") && m.RepoPasteMode {
		return h.handleRepositoryPasteKeys(m, msg)
	}

	switch keyStr {
	case "ctrl+c", "esc":
		return h.exitSettingsMode(m), nil
	case "up", "k":
		// If in repository section, handle navigation based on active section
		if m.SettingsSection == "repositories" || m.SettingsSection == "" {
			return h.handleRepositoryUpNavigation(m)
		}
		m.SettingsCursor--
		if m.SettingsCursor < 0 {
			m.SettingsCursor = 0
		}
		return m, nil
	case "down", "j":
		// If in repository section, handle navigation based on active section
		if m.SettingsSection == "repositories" || m.SettingsSection == "" {
			return h.handleRepositoryDownNavigation(m)
		}
		// Get appropriate max based on current section
		var maxItems int
		switch m.SettingsSection {
		case "actions":
			maxItems = len(m.Config.Keybindings.Actions) - 1
		case "theme":
			// Count total theme items - hardcoded for now as 32 items
			maxItems = 31 // 0-indexed, so 32 items = max index 31
		default: // repositories
			maxItems = len(m.Dependencies.GetRepoManager().GetItems()) - 1
		}
		if maxItems >= 0 {
			m.SettingsCursor++
			if m.SettingsCursor > maxItems {
				m.SettingsCursor = maxItems
			}
		}
		return m, nil
	case "]":
		// Switch to next tab
		switch m.SettingsSection {
		case "repositories", "":
			m.SettingsSection = "actions"
		case "actions":
			m.SettingsSection = "theme"
		case "theme":
			m.SettingsSection = "repositories"
		}
		m.SettingsCursor = 0
		return m, nil
	case "[":
		// Switch to previous tab
		switch m.SettingsSection {
		case "actions":
			m.SettingsSection = "repositories"
		case "theme":
			m.SettingsSection = "actions"
		case "repositories", "":
			m.SettingsSection = "theme"
		}
		m.SettingsCursor = 0
		return m, nil
	case "tab":
		// Handle tab navigation for repositories section
		if m.SettingsSection == "repositories" || m.SettingsSection == "" {
			return h.handleRepositoryTabNavigation(m)
		}
		// Keep tab functionality as fallback (forward only) for other sections
		switch m.SettingsSection {
		case "actions":
			m.SettingsSection = "theme"
		case "theme":
			m.SettingsSection = "repositories"
		}
		m.SettingsCursor = 0
		return m, nil
	case "enter":
		// Handle enter key for repositories section
		if m.SettingsSection == "repositories" || m.SettingsSection == "" {
			return h.handleRepositoryEnterNavigation(m)
		}
		return m, nil
	case " ":
		// Handle space key for repositories section
		if m.SettingsSection == "repositories" || m.SettingsSection == "" {
			return h.handleRepositorySpaceToggle(m)
		}
		return m, nil
	case "e":
		// Edit functionality for various sections
		if m.SettingsSection == "theme" {
			// Start theme editing mode
			return h.startThemeEdit(m)
		} else if m.SettingsSection == "actions" {
			// Start action editing mode
			return h.startActionEdit(m)
		}
		return m, nil
	case "d":
		// Delete functionality
		if m.SettingsSection == "repositories" || m.SettingsSection == "" {
			return h.repositoryHandler.DeleteSelectedRepository(m)
		} else if m.SettingsSection == "actions" {
			return h.deleteSelectedActionInSettings(m)
		}
		return m, nil
	case "a":
		// Add functionality
		if m.SettingsSection == "actions" {
			return h.addNewActionInSettings(m)
		} else if m.SettingsSection == "repositories" || m.SettingsSection == "" {
			// Activate paste mode for adding repositories
			m.RepoActiveSection = "paste"
			m.RepoPasteMode = true
			return m, nil
		}
		return m, nil
	case "r":
		return m, m.updateRepositoryStatuses()
	case "q":
		return m, tea.Quit
	case "?":
		return h.toggleHelpModal(m), nil
	default:
		return m, nil
	}
}

// handleDetailsViewKeys handles key events in details view.
func (h *KeyHandler) handleDetailsViewKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	switch keyStr {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "b", "esc":
		// Return to list view
		m.State = ListView
		m.SelectedNavItem = nil
		return m, nil
	case "?":
		return h.toggleHelpModal(m), nil
	}

	return m, nil
}

// handleActionConfigViewKeys handles key events in action configuration view.
func (h *KeyHandler) handleActionConfigViewKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	if m.ActionConfigEditMode {
		return h.handleActionEditKeys(m, msg)
	}

	switch keyStr {
	case "ctrl+c", "esc":
		m.State = SettingsView
		return m, nil
	case "up", "k":
		if m.ActionConfigCursor > 0 {
			m.ActionConfigCursor--
		}
		return m, nil
	case "down", "j":
		actions := m.Config.Keybindings.Actions
		if m.ActionConfigCursor < len(actions)-1 {
			m.ActionConfigCursor++
		}
		return m, nil
	case "a":
		// Add new action
		m.ActionConfigAction = &config.Action{}
		m.ActionConfigEditMode = true
		m.ActionConfigIsNew = true
		m.ActionConfigFieldIdx = 0
		return m, nil
	case "enter", "e":
		// Edit selected action
		actions := m.Config.Keybindings.Actions
		if len(actions) > 0 && m.ActionConfigCursor < len(actions) {
			// Make a copy of the action to edit
			selectedAction := actions[m.ActionConfigCursor]
			m.ActionConfigAction = &config.Action{
				Name:        selectedAction.Name,
				Key:         selectedAction.Key,
				Command:     selectedAction.Command,
				Args:        make([]string, len(selectedAction.Args)),
				Description: selectedAction.Description,
			}
			copy(m.ActionConfigAction.Args, selectedAction.Args)
			m.ActionConfigEditMode = true
			m.ActionConfigIsNew = false
			m.ActionConfigFieldIdx = 0
		}
		return m, nil
	case "d":
		// Delete selected action
		return h.deleteSelectedAction(m)
	case "q":
		return m, tea.Quit
	case "?":
		return h.toggleHelpModal(m), nil
	default:
		return m, nil
	}
}

// handleActionEditKeys handles key events when editing an action.
func (h *KeyHandler) handleActionEditKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	switch keyStr {
	case "ctrl+c", "esc":
		// Cancel editing
		m.ActionConfigEditMode = false
		m.ActionConfigAction = nil
		m.ActionConfigIsNew = false
		return m, nil
	case "up", "k":
		if m.ActionConfigFieldIdx > 0 {
			m.ActionConfigFieldIdx--
		}
		return m, nil
	case "down", "j":
		if m.ActionConfigFieldIdx < 4 { // 5 fields: name, key, description, command, args
			m.ActionConfigFieldIdx++
		}
		return m, nil
	case "ctrl+s":
		// Save action
		return h.saveAction(m)
	case "enter":
		// Start editing current field (placeholder - would need proper input handling)
		return m, nil
	case "?":
		return h.toggleHelpModal(m), nil
	default:
		return m, nil
	}
}

// enterSettingsMode switches to settings mode.
func (h *KeyHandler) enterSettingsMode(m Model) Model {
	m.PreviousState = m.State
	m.State = SettingsView
	m.SettingsSection = "repositories"
	m.SettingsCursor = 0
	return m
}

// exitSettingsMode exits settings mode.
func (h *KeyHandler) exitSettingsMode(m Model) Model {
	m.State = m.PreviousState
	return m
}

// toggleHelpModal toggles the help modal display.
func (h *KeyHandler) toggleHelpModal(m Model) Model {
	m.ShowHelpModal = !m.ShowHelpModal
	return m
}

// discoverWorktrees initiates worktree discovery.
func (h *KeyHandler) discoverWorktrees(m Model) (Model, tea.Cmd) {
	// Implementation would go here - placeholder for now
	return m, nil
}

// executeConfiguredAction executes a configured action based on key binding.
func (h *KeyHandler) executeConfiguredAction(m Model, keyStr string) (tea.Model, tea.Cmd) {
	// Check for configurable actions
	if action := m.Config.Keybindings.FindActionByKey(keyStr); action != nil {
		navigableItems := m.getNavigableItems()
		if m.Cursor >= len(navigableItems) {
			return m, nil
		}

		selectedItem := navigableItems[m.Cursor]
		var targetPath string

		if selectedItem.Type == "worktree" {
			targetPath = selectedItem.WorktreeInfo.Path
		} else if selectedItem.Type == "repository" {
			targetPath = selectedItem.Repository.Path
		} else {
			return m, nil
		}

		// Use the configured action
		cmd := action.ExecuteOpenAction(targetPath)

		return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
			if err != nil {
				logging.Get().Error("failed to run configured action",
					"error", err,
					"path", targetPath,
					"action", action.Name,
					"key", action.Key)
			}
			return nil
		})
	}
	return m, nil
}

// executeConfiguredActionForExplorer executes a configured action in explorer view.
func (h *KeyHandler) executeConfiguredActionForExplorer(m Model, keyStr string) (tea.Model, tea.Cmd) {
	// Explorer functionality temporarily disabled
	return m, nil
}

// deleteSelectedAction deletes the currently selected action from the configuration.
func (h *KeyHandler) deleteSelectedAction(m Model) (Model, tea.Cmd) {
	actions := m.Config.Keybindings.Actions
	if len(actions) == 0 || m.ActionConfigCursor >= len(actions) {
		return m, nil
	}

	// Remove the selected action
	m.Config.Keybindings.Actions = append(actions[:m.ActionConfigCursor], actions[m.ActionConfigCursor+1:]...)

	// Adjust cursor if necessary
	if m.ActionConfigCursor >= len(m.Config.Keybindings.Actions) && len(m.Config.Keybindings.Actions) > 0 {
		m.ActionConfigCursor = len(m.Config.Keybindings.Actions) - 1
	} else if len(m.Config.Keybindings.Actions) == 0 {
		m.ActionConfigCursor = 0
	}

	// Save configuration
	if err := m.Dependencies.GetConfigService().Save(m.Config); err != nil {
		logging.Get().Error("failed to save config after deleting action", "error", err)
	}

	return m, nil
}

// saveAction saves the currently edited action to the configuration.
func (h *KeyHandler) saveAction(m Model) (Model, tea.Cmd) {
	if m.ActionConfigAction == nil {
		return m, nil
	}

	// Validate action
	if m.ActionConfigAction.Name == "" || m.ActionConfigAction.Key == "" || m.ActionConfigAction.Command == "" {
		// TODO: Could show error message to user
		return m, nil
	}

	if m.ActionConfigIsNew {
		// Add new action
		m.Config.Keybindings.Actions = append(m.Config.Keybindings.Actions, *m.ActionConfigAction)
	} else {
		// Update existing action
		actions := m.Config.Keybindings.Actions
		if m.ActionConfigCursor < len(actions) {
			actions[m.ActionConfigCursor] = *m.ActionConfigAction
		}
	}

	// Save configuration
	if err := m.Dependencies.GetConfigService().Save(m.Config); err != nil {
		logging.Get().Error("failed to save config after saving action", "error", err)
	}

	// Exit edit mode
	m.ActionConfigEditMode = false
	m.ActionConfigAction = nil
	m.ActionConfigIsNew = false

	return m, nil
}

// startActionEdit starts action editing mode for the selected action in settings
func (h *KeyHandler) startActionEdit(m Model) (Model, tea.Cmd) {
	actions := m.Config.Keybindings.Actions
	if len(actions) == 0 || m.SettingsCursor >= len(actions) {
		return m, nil
	}

	// Start editing the name field by default
	m.ActionEditMode = true
	m.ActionEditItemIndex = m.SettingsCursor
	m.ActionEditFieldType = "name"

	// Get the current value based on field type
	action := actions[m.SettingsCursor]
	m.ActionEditValue = action.Name

	return m, nil
}

// addNewActionInSettings adds a new action in the settings view
func (h *KeyHandler) addNewActionInSettings(m Model) (Model, tea.Cmd) {
	// Add an empty action to the config
	newAction := config.Action{
		Name:    "New Action",
		Key:     "",
		Command: "",
		Args:    []string{},
	}

	m.Config.Keybindings.Actions = append(m.Config.Keybindings.Actions, newAction)

	// Move cursor to the new action and start editing
	m.SettingsCursor = len(m.Config.Keybindings.Actions) - 1
	m.ActionEditMode = true
	m.ActionEditItemIndex = m.SettingsCursor
	m.ActionEditFieldType = "name"
	m.ActionEditValue = "New Action"

	return m, nil
}

// deleteSelectedActionInSettings deletes the selected action in settings
func (h *KeyHandler) deleteSelectedActionInSettings(m Model) (Model, tea.Cmd) {
	actions := m.Config.Keybindings.Actions
	if len(actions) == 0 || m.SettingsCursor >= len(actions) {
		return m, nil
	}

	// Remove the selected action
	m.Config.Keybindings.Actions = append(actions[:m.SettingsCursor], actions[m.SettingsCursor+1:]...)

	// Adjust cursor if necessary
	if m.SettingsCursor >= len(m.Config.Keybindings.Actions) && len(m.Config.Keybindings.Actions) > 0 {
		m.SettingsCursor = len(m.Config.Keybindings.Actions) - 1
	} else if len(m.Config.Keybindings.Actions) == 0 {
		m.SettingsCursor = 0
	}

	// Save configuration
	if err := m.Dependencies.GetConfigService().Save(m.Config); err != nil {
		logging.Get().Error("failed to save config after deleting action", "error", err)
	}

	return m, nil
}

// handleActionEditKeysInSettings handles key events when editing an action in settings
func (h *KeyHandler) handleActionEditKeysInSettings(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	switch keyStr {
	case "ctrl+c", "esc":
		// Cancel editing
		m.ActionEditMode = false
		m.ActionEditValue = ""
		m.ActionEditFieldType = ""
		return m, nil
	case "enter":
		// Save current field and move to next field or finish
		if err := h.saveActionFieldEdit(m); err != nil {
			// TODO: Show error to user
			return m, nil
		}

		// Move to next field or finish editing
		return h.moveToNextActionField(m), nil
	case "tab":
		// Move to next field without saving current changes
		return h.moveToNextActionField(m), nil
	case "backspace":
		// Remove last character
		if len(m.ActionEditValue) > 0 {
			m.ActionEditValue = m.ActionEditValue[:len(m.ActionEditValue)-1]
		}
		return m, nil
	default:
		// Add character to the edit value
		if len(keyStr) == 1 {
			m.ActionEditValue += keyStr
		}
		return m, nil
	}
}

// saveActionFieldEdit saves the current field value being edited
func (h *KeyHandler) saveActionFieldEdit(m Model) error {
	actions := m.Config.Keybindings.Actions
	if m.ActionEditItemIndex >= len(actions) {
		return nil
	}

	action := &actions[m.ActionEditItemIndex]

	switch m.ActionEditFieldType {
	case "name":
		action.Name = m.ActionEditValue
	case "key":
		action.Key = m.ActionEditValue
	case "command":
		action.Command = m.ActionEditValue
	case "args":
		// Split space-separated arguments
		if strings.TrimSpace(m.ActionEditValue) == "" {
			action.Args = nil
		} else {
			action.Args = strings.Fields(m.ActionEditValue)
		}
	}

	// Update the action in the slice
	m.Config.Keybindings.Actions[m.ActionEditItemIndex] = *action

	// Save configuration
	return m.Dependencies.GetConfigService().Save(m.Config)
}

// moveToNextActionField moves to the next field in action editing
func (h *KeyHandler) moveToNextActionField(m Model) Model {
	actions := m.Config.Keybindings.Actions
	if m.ActionEditItemIndex >= len(actions) {
		m.ActionEditMode = false
		return m
	}

	action := actions[m.ActionEditItemIndex]

	switch m.ActionEditFieldType {
	case "name":
		m.ActionEditFieldType = "key"
		m.ActionEditValue = action.Key
	case "key":
		m.ActionEditFieldType = "command"
		m.ActionEditValue = action.Command
	case "command":
		m.ActionEditFieldType = "args"
		m.ActionEditValue = strings.Join(action.Args, " ")
	case "args":
		// Finished editing all fields
		m.ActionEditMode = false
		m.ActionEditValue = ""
		m.ActionEditFieldType = ""
	default:
		// Start with name if no field type set
		m.ActionEditFieldType = "name"
		m.ActionEditValue = action.Name
	}

	return m
}

// handleThemeEditKeys handles key events when editing theme in settings view
func (h *KeyHandler) handleThemeEditKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	switch keyStr {
	case "ctrl+c", "esc":
		// Cancel theme editing
		m.ThemeEditMode = false
		m.ThemeEditValue = ""
		return m, nil
	case "enter":
		// Save the theme property value
		if m.ThemeEditValue != "" {
			if err := h.saveThemePropertyEdit(m); err != nil {
				// TODO: Show error to user
				return m, nil
			}
		}
		m.ThemeEditMode = false
		m.ThemeEditValue = ""
		return m, nil
	case "backspace":
		// Remove last character
		if len(m.ThemeEditValue) > 0 {
			m.ThemeEditValue = m.ThemeEditValue[:len(m.ThemeEditValue)-1]
		}
		return m, nil
	default:
		// Add character to theme property value
		if len(keyStr) == 1 {
			m.ThemeEditValue += keyStr
		}
		return m, nil
	}
}

// startThemeEdit starts theme editing mode in settings
func (h *KeyHandler) startThemeEdit(m Model) (tea.Model, tea.Cmd) {
	m.ThemeEditMode = true
	m.ThemeEditItemIndex = m.SettingsCursor

	// Get the current value of the theme property being edited
	themeItems := h.getAllThemeItems(m.Config.Theme)
	if m.SettingsCursor >= 0 && m.SettingsCursor < len(themeItems) {
		m.ThemeEditValue = themeItems[m.SettingsCursor].value
	}

	return m, nil
}

// saveThemePropertyEdit saves the edited theme property
func (h *KeyHandler) saveThemePropertyEdit(m Model) error {
	themeItems := h.getAllThemeItems(m.Config.Theme)
	if m.ThemeEditItemIndex < 0 || m.ThemeEditItemIndex >= len(themeItems) {
		return nil
	}

	item := themeItems[m.ThemeEditItemIndex]

	// Update the appropriate field based on the item name and type
	switch item.name {
	// Colors
	case "Title":
		m.Config.Theme.Colors.Title = m.ThemeEditValue
	case "Title Background":
		m.Config.Theme.Colors.TitleBackground = m.ThemeEditValue
	case "Selected":
		m.Config.Theme.Colors.Selected = m.ThemeEditValue
	case "Selected Background":
		m.Config.Theme.Colors.SelectedBackground = m.ThemeEditValue
	case "Help Text":
		m.Config.Theme.Colors.Help = m.ThemeEditValue
	case "Border":
		m.Config.Theme.Colors.Border = m.ThemeEditValue
	case "Modal Background":
		m.Config.Theme.Colors.ModalBackground = m.ThemeEditValue
	case "Branch":
		m.Config.Theme.Colors.Branch = m.ThemeEditValue
	case "Regular Icon":
		m.Config.Theme.Colors.IconRegular = m.ThemeEditValue
	case "Bare Icon":
		m.Config.Theme.Colors.IconBare = m.ThemeEditValue
	case "Worktree Icon":
		m.Config.Theme.Colors.IconWorktree = m.ThemeEditValue
	case "Clean Status Color":
		m.Config.Theme.Colors.StatusClean = m.ThemeEditValue
	case "Dirty Status Color":
		m.Config.Theme.Colors.StatusDirty = m.ThemeEditValue
	case "Unpushed Status Color":
		m.Config.Theme.Colors.StatusUnpushed = m.ThemeEditValue
	case "Untracked Status Color":
		m.Config.Theme.Colors.StatusUntracked = m.ThemeEditValue
	case "Error Status Color":
		m.Config.Theme.Colors.StatusError = m.ThemeEditValue
	case "Not Added Status Color":
		m.Config.Theme.Colors.StatusNotAdded = m.ThemeEditValue

	// Indicators
	case "Clean Status Icon":
		m.Config.Theme.Indicators.Clean = m.ThemeEditValue
	case "Dirty Status Icon":
		m.Config.Theme.Indicators.Dirty = m.ThemeEditValue
	case "Unpushed Status Icon":
		m.Config.Theme.Indicators.Unpushed = m.ThemeEditValue
	case "Untracked Status Icon":
		m.Config.Theme.Indicators.Untracked = m.ThemeEditValue
	case "Error Status Icon":
		m.Config.Theme.Indicators.Error = m.ThemeEditValue
	case "Not Added Status Icon":
		m.Config.Theme.Indicators.NotAdded = m.ThemeEditValue
	case "Selected Indicator":
		m.Config.Theme.Indicators.Selected = m.ThemeEditValue
	case "Selected End":
		m.Config.Theme.Indicators.SelectedEnd = m.ThemeEditValue

	// Icons
	case "Regular Repository":
		m.Config.Theme.Icons.Repository.Regular = m.ThemeEditValue
	case "Bare Repository":
		m.Config.Theme.Icons.Repository.Bare = m.ThemeEditValue
	case "Worktree Repository":
		m.Config.Theme.Icons.Repository.Worktree = m.ThemeEditValue
	case "Branch Icon":
		m.Config.Theme.Icons.Branch.Icon = m.ThemeEditValue
	case "Tree Branch":
		m.Config.Theme.Icons.Tree.Branch = m.ThemeEditValue
	case "Tree Last":
		m.Config.Theme.Icons.Tree.Last = m.ThemeEditValue
	case "Folder Icon":
		m.Config.Theme.Icons.Folder.Icon = m.ThemeEditValue
	}

	// Save configuration
	return m.Dependencies.GetConfigService().Save(m.Config)
}

// ThemeItem represents a single editable theme item (copied from settings renderer for consistency)
type ThemeItem struct {
	name     string
	value    string
	itemType string // "color", "icon", "indicator"
	category string
}

// getAllThemeItems returns all editable theme items (copied from settings renderer for consistency)
func (h *KeyHandler) getAllThemeItems(themeConfig theme.Theme) []ThemeItem {
	var items []ThemeItem

	// Colors
	items = append(items, []ThemeItem{
		{"Title", themeConfig.Colors.Title, "color", "Colors"},
		{"Title Background", themeConfig.Colors.TitleBackground, "color", "Colors"},
		{"Selected", themeConfig.Colors.Selected, "color", "Colors"},
		{"Selected Background", themeConfig.Colors.SelectedBackground, "color", "Colors"},
		{"Help Text", themeConfig.Colors.Help, "color", "Colors"},
		{"Border", themeConfig.Colors.Border, "color", "Colors"},
		{"Modal Background", themeConfig.Colors.ModalBackground, "color", "Colors"},
		{"Branch", themeConfig.Colors.Branch, "color", "Colors"},
		{"Regular Icon", themeConfig.Colors.IconRegular, "color", "Colors"},
		{"Bare Icon", themeConfig.Colors.IconBare, "color", "Colors"},
		{"Worktree Icon", themeConfig.Colors.IconWorktree, "color", "Colors"},
	}...)

	// Status colors and indicators
	items = append(items, []ThemeItem{
		{"Clean Status Color", themeConfig.Colors.StatusClean, "color", "Status Indicators"},
		{"Clean Status Icon", themeConfig.Indicators.Clean, "indicator", "Status Indicators"},
		{"Dirty Status Color", themeConfig.Colors.StatusDirty, "color", "Status Indicators"},
		{"Dirty Status Icon", themeConfig.Indicators.Dirty, "indicator", "Status Indicators"},
		{"Unpushed Status Color", themeConfig.Colors.StatusUnpushed, "color", "Status Indicators"},
		{"Unpushed Status Icon", themeConfig.Indicators.Unpushed, "indicator", "Status Indicators"},
		{"Untracked Status Color", themeConfig.Colors.StatusUntracked, "color", "Status Indicators"},
		{"Untracked Status Icon", themeConfig.Indicators.Untracked, "indicator", "Status Indicators"},
		{"Error Status Color", themeConfig.Colors.StatusError, "color", "Status Indicators"},
		{"Error Status Icon", themeConfig.Indicators.Error, "indicator", "Status Indicators"},
		{"Not Added Status Color", themeConfig.Colors.StatusNotAdded, "color", "Status Indicators"},
		{"Not Added Status Icon", themeConfig.Indicators.NotAdded, "indicator", "Status Indicators"},
	}...)

	// Repository icons
	items = append(items, []ThemeItem{
		{"Regular Repository", themeConfig.Icons.Repository.Regular, "icon", "Repository Icons"},
		{"Bare Repository", themeConfig.Icons.Repository.Bare, "icon", "Repository Icons"},
		{"Worktree Repository", themeConfig.Icons.Repository.Worktree, "icon", "Repository Icons"},
	}...)

	// UI icons
	items = append(items, []ThemeItem{
		{"Selected Indicator", themeConfig.Indicators.Selected, "indicator", "UI Icons"},
		{"Selected End", themeConfig.Indicators.SelectedEnd, "indicator", "UI Icons"},
		{"Branch Icon", themeConfig.Icons.Branch.Icon, "icon", "UI Icons"},
		{"Tree Branch", themeConfig.Icons.Tree.Branch, "icon", "UI Icons"},
		{"Tree Last", themeConfig.Icons.Tree.Last, "icon", "UI Icons"},
		{"Folder Icon", themeConfig.Icons.Folder.Icon, "icon", "UI Icons"},
	}...)

	return items
}

// handleRepositoryTabNavigation handles Tab navigation within repository sections
func (h *KeyHandler) handleRepositoryTabNavigation(m Model) (Model, tea.Cmd) {
	switch m.RepoActiveSection {
	case "list":
		m.RepoActiveSection = "explorer"
		// Initialize explorer if not already done
		if m.RepoExplorer == nil {
			m = h.initializeExplorer(m)
		} else {
			// Update repository list when switching to explorer
			h.updateExplorerRepositoryList(m)
		}
	case "explorer":
		m.RepoActiveSection = "list"
	case "paste":
		// Paste mode stays in paste, use 'a' to activate and 'esc' to exit
		m.RepoActiveSection = "list"
	default:
		m.RepoActiveSection = "list"
	}
	return m, nil
}

// handleRepositoryPasteKeys handles key events when in paste input mode
func (h *KeyHandler) handleRepositoryPasteKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	switch keyStr {
	case "ctrl+c", "esc":
		// Exit paste mode
		m.RepoPasteMode = false
		m.RepoPasteValue = ""
		return m, nil
	case "enter":
		// Add repository from paste input
		if m.RepoPasteValue != "" {
			return h.addRepositoryFromPath(m, m.RepoPasteValue)
		}
		return m, nil
	case "backspace":
		// Remove last character
		if len(m.RepoPasteValue) > 0 {
			m.RepoPasteValue = m.RepoPasteValue[:len(m.RepoPasteValue)-1]
		}
		return m, nil
	default:
		// Add character to paste value
		if len(keyStr) == 1 {
			m.RepoPasteValue += keyStr
		}
		return m, nil
	}
}

// initializeExplorer initializes the directory explorer
func (h *KeyHandler) initializeExplorer(m Model) Model {
	// Start from user's home directory
	homeDir := "."
	if userHome := os.Getenv("HOME"); userHome != "" {
		homeDir = userHome
	}

	// Create explorer styles based on current theme
	explorerStyles := direxplorer.StyleConfig{
		Directory:    lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.Branch)),
		File:         lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.Help)),
		GitRepo:      lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.Selected)).Bold(true),
		BareRepo:     lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.IconBare)).Bold(true),
		AlreadyAdded: lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.StatusClean)),
		Selected:     lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.Selected)).Bold(true),
		CurrentPath:  lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.Title)).Bold(true),
		EmptyState:   lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.Help)),
	}

	m.RepoExplorer = direxplorer.NewExplorer(homeDir, explorerStyles, m.Config.Theme)

	// Update the explorer with already-added repository paths
	h.updateExplorerRepositoryList(m)

	return m
}

// updateExplorerRepositoryList updates the explorer with the current list of added repositories
func (h *KeyHandler) updateExplorerRepositoryList(m Model) {
	if m.RepoExplorer != nil {
		repoItems := m.Dependencies.GetRepoManager().GetItems()
		var repoPaths []string
		for _, item := range repoItems {
			repoPaths = append(repoPaths, item.Path)
		}
		m.RepoExplorer.UpdateAddedRepositories(repoPaths)
	}
}

// addRepositoryFromPath adds a repository from the given path
func (h *KeyHandler) addRepositoryFromPath(m Model, path string) (Model, tea.Cmd) {
	// Clean the path
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" {
		return m, nil
	}

	// Add the repository using the repository handler
	if err := h.repositoryHandler.AddRepository(m, cleanPath); err != nil {
		// TODO: Show error to user
		logging.Get().Error("failed to add repository from path", "error", err, "path", cleanPath)
		return m, nil
	}

	// Clear paste input and exit paste mode
	m.RepoPasteMode = false
	m.RepoPasteValue = ""

	// Update the explorer's repository list
	h.updateExplorerRepositoryList(m)

	// Update repository list
	return m, m.updateRepositoryStatuses()
}

// handleRepositoryUpNavigation handles up navigation in repository sections
func (h *KeyHandler) handleRepositoryUpNavigation(m Model) (Model, tea.Cmd) {
	switch m.RepoActiveSection {
	case "list":
		// Navigate up in repository list
		m.SettingsCursor--
		if m.SettingsCursor < 0 {
			m.SettingsCursor = 0
		}
	case "explorer":
		// Navigate up in directory explorer
		if m.RepoExplorer != nil {
			m.RepoExplorer.MoveCursorUp()
		}
	case "paste":
		// In paste mode, up navigation doesn't do anything special
	}
	return m, nil
}

// handleRepositoryDownNavigation handles down navigation in repository sections
func (h *KeyHandler) handleRepositoryDownNavigation(m Model) (Model, tea.Cmd) {
	switch m.RepoActiveSection {
	case "list":
		// Navigate down in repository list
		maxItems := len(m.Dependencies.GetRepoManager().GetItems()) - 1
		if maxItems >= 0 {
			m.SettingsCursor++
			if m.SettingsCursor > maxItems {
				m.SettingsCursor = maxItems
			}
		}
	case "explorer":
		// Navigate down in directory explorer
		if m.RepoExplorer != nil {
			m.RepoExplorer.MoveCursorDown()
		}
	case "paste":
		// In paste mode, down navigation doesn't do anything special
	}
	return m, nil
}

// handleRepositoryEnterNavigation handles enter key in repository sections
func (h *KeyHandler) handleRepositoryEnterNavigation(m Model) (Model, tea.Cmd) {
	switch m.RepoActiveSection {
	case "list":
		// Navigate to selected item details (for repositories)
		navigableItems := m.getNavigableItems()
		if m.SettingsCursor < len(navigableItems) {
			selectedItem := navigableItems[m.SettingsCursor]
			if selectedItem.Type == "repository" || selectedItem.Type == "worktree" {
				m.State = DetailsView
				m.SelectedNavItem = &selectedItem
			}
		}
	case "explorer":
		// Enter directory only (space is used for repository toggle)
		if m.RepoExplorer != nil {
			selectedItem := m.RepoExplorer.GetSelectedItem()
			if selectedItem != nil && selectedItem.IsDir && !selectedItem.IsGitRepo {
				// Enter directory (only for non-git directories)
				m.RepoExplorer.NavigateInto()
			}
		}
	case "paste":
		// Add repository from paste input
		if m.RepoPasteValue != "" {
			return h.addRepositoryFromPath(m, m.RepoPasteValue)
		}
	}
	return m, nil
}

// handleRepositorySpaceToggle handles space key for toggling repositories
func (h *KeyHandler) handleRepositorySpaceToggle(m Model) (Model, tea.Cmd) {
	switch m.RepoActiveSection {
	case "explorer":
		// Toggle repository add/remove with space
		if m.RepoExplorer != nil {
			selectedItem := m.RepoExplorer.GetSelectedItem()
			if selectedItem != nil && selectedItem.IsGitRepo {
				if selectedItem.IsAlreadyAdded {
					// Remove repository
					return h.removeRepositoryFromPath(m, selectedItem.Path)
				} else {
					// Add repository
					return h.addRepositoryFromPath(m, selectedItem.Path)
				}
			}
		}
	case "list":
		// In list view, space could toggle repository enable/disable (future feature)
		// For now, do nothing
	case "paste":
		// In paste mode, space adds a space character
		m.RepoPasteValue += " "
	}
	return m, nil
}

// removeRepositoryFromPath removes a repository from the given path and updates the UI
func (h *KeyHandler) removeRepositoryFromPath(m Model, path string) (Model, tea.Cmd) {
	// Remove the repository using the repository handler
	h.repositoryHandler.RemoveRepositoryByPath(m, path)

	// Update the explorer's repository list
	h.updateExplorerRepositoryList(m)

	// Update repository list
	return m, m.updateRepositoryStatuses()
}
