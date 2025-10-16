package ui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/logging"
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

	switch keyStr {
	case "ctrl+c", "esc":
		return h.exitSettingsMode(m), nil
	case "up", "k":
		m.SettingsCursor--
		if m.SettingsCursor < 0 {
			m.SettingsCursor = 0
		}
		return m, nil
	case "down", "j":
		// Get appropriate max based on current section
		var maxItems int
		switch m.SettingsSection {
		case "actions":
			maxItems = len(m.Config.Keybindings.Actions) - 1
		case "theme":
			maxItems = 4 // Number of theme items
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
		// Keep tab functionality as fallback (forward only)
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
	case "enter":
		// Navigate to selected item details (for repositories)
		if m.SettingsSection == "repositories" || m.SettingsSection == "" {
			navigableItems := m.getNavigableItems()
			if m.SettingsCursor < len(navigableItems) {
				selectedItem := navigableItems[m.SettingsCursor]
				if selectedItem.Type == "repository" || selectedItem.Type == "worktree" {
					m.State = DetailsView
					m.SelectedNavItem = &selectedItem
				}
			}
		}
		return m, nil
	case "e":
		// Edit functionality for various sections
		return m, nil
	case "d":
		// Delete functionality (mainly for repositories)
		if m.SettingsSection == "repositories" || m.SettingsSection == "" {
			return h.repositoryHandler.DeleteSelectedRepository(m)
		}
		return m, nil
	case "a":
		// Add functionality (mainly for actions)
		if m.SettingsSection == "actions" {
			m.PreviousState = m.State
			m.State = ActionConfigView
			m.ActionConfigCursor = 0
			m.ActionConfigEditMode = true
			m.ActionConfigIsNew = true
			m.ActionConfigAction = &config.Action{}
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
