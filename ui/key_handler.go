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
	explorerHandler   *ExplorerHandler
}

// NewKeyHandler creates a new KeyHandler instance.
func NewKeyHandler() *KeyHandler {
	return &KeyHandler{
		navigationHandler: NewNavigationHandler(),
		repositoryHandler: NewRepositoryOperationHandler(),
		explorerHandler:   NewExplorerHandler(),
	}
}

// HandleKeyPress dispatches key events to appropriate handlers based on current state.
func (h *KeyHandler) HandleKeyPress(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Log current state and key press
	stateNames := []string{"ListView", "RepoManagementView", "ExplorerView", "DetailsView", "ActionConfigView"}
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
	case RepoManagementView:
		return h.handleRepoManagementViewKeys(m, msg)
	case ExplorerView:
		return h.handleExplorerViewKeys(m, msg)
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
		return h.explorerHandler.NavigateToSelected(m)
	case "m":
		return h.enterRepoManagementMode(m), nil
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

// handleRepoManagementViewKeys handles key events in repository management view.
func (h *KeyHandler) handleRepoManagementViewKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	switch keyStr {
	case "ctrl+c", "esc":
		return h.exitRepoManagementMode(m), nil
	case "up", "k":
		return h.navigationHandler.MoveCursorUp(m), nil
	case "down", "j":
		return h.navigationHandler.MoveCursorDown(m), nil
	case "enter":
		return h.explorerHandler.NavigateToSelected(m)
	case "e":
		return h.explorerHandler.EnterExplorerMode(m)
	case "d":
		return h.repositoryHandler.DeleteSelectedRepository(m)
	case "c":
		// Enter action configuration mode
		m.PreviousState = m.State
		m.State = ActionConfigView
		m.ActionConfigCursor = 0
		m.ActionConfigEditMode = false
		m.ActionConfigAction = nil
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

// handleExplorerViewKeys handles key events in explorer view.
func (h *KeyHandler) handleExplorerViewKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	logging.Get().Debug("explorer key pressed",
		"key", keyStr,
		"length", len(keyStr),
		"bytes", []byte(keyStr))

	switch keyStr {
	case "ctrl+c", "esc", "q":
		m.State = m.PreviousState // Return to the previous state
		return m, nil
	case "up", "k":
		return h.navigationHandler.MoveExplorerCursorUp(m), nil
	case "down", "j":
		return h.navigationHandler.MoveExplorerCursorDown(m), nil
	case "enter":
		return h.explorerHandler.HandleExplorerSelection(m)
	case "space", " ":
		logging.Get().Info("space key detected, calling toggleRepositorySelection", "key", keyStr)
		return h.repositoryHandler.ToggleRepositorySelection(m)
	case "?":
		return h.toggleHelpModal(m), nil
	default:
		// Check for configurable actions
		return h.executeConfiguredActionForExplorer(m, keyStr)
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
		m.State = RepoManagementView
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

// enterRepoManagementMode switches to repository management mode.
func (h *KeyHandler) enterRepoManagementMode(m Model) Model {
	m.PreviousState = m.State
	m.State = RepoManagementView
	m.InputField = ""
	m.InputPrompt = "Enter repository path: "
	return m
}

// exitRepoManagementMode exits repository management mode.
func (h *KeyHandler) exitRepoManagementMode(m Model) Model {
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
	// Check for configurable actions
	if action := m.Config.Keybindings.FindActionByKey(keyStr); action != nil {
		if m.ExplorerCursor >= len(m.ExplorerItems) {
			return m, nil
		}

		selectedItem := m.ExplorerItems[m.ExplorerCursor]
		targetPath := selectedItem.Path

		// If it's not a git repository, try to find the nearest git repository
		if !selectedItem.IsGitRepo {
			// Use the current explorer path as fallback
			targetPath = m.ExplorerPath
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
