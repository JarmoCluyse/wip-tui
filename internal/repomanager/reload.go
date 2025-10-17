package repomanager

// ReloadWorktrees reloads worktrees for all bare repositories.
func (rm *RepoManager) ReloadWorktrees() error {
	for _, item := range rm.items {
		if item.IsBare {
			rm.loadWorktrees(item)
		}
	}
	return nil
}

// ReloadStatus reloads status for all repositories and their worktrees.
func (rm *RepoManager) ReloadStatus() error {
	for _, item := range rm.items {
		rm.updateRepoStatus(item)

		// Update status for all worktrees
		for _, subItem := range item.SubItems {
			rm.updateSubItemStatus(subItem)
		}
	}
	return nil
}
