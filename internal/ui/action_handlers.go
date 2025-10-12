package ui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/jarmocluyse/wip-tui/internal/config"
	"github.com/jarmocluyse/wip-tui/internal/logging"
)

// executeConfiguredAction executes a user-configured action on the currently selected repository.
func (m Model) executeConfiguredAction(action config.Action) (tea.Model, tea.Cmd) {
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
