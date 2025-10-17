package repomanager

// Init initializes the repository manager by loading paths from config and building the item list.
func (rm *RepoManager) Init() error {
	config, err := rm.configService.Load()
	if err != nil {
		return err
	}
	// Clear existing items
	rm.items = make([]*RepoItem, 0, len(config.RepositoryPaths))
	// Load repositories from config paths
	for _, path := range config.RepositoryPaths {
		item := &RepoItem{
			Name:     extractNameFromPath(path),
			Path:     path,
			SubItems: make([]*SubItem, 0),
		}
		// Update status for this repository
		rm.updateRepoStatus(item)
		if item.IsBare {
			rm.loadWorktrees(item)
		}

		rm.items = append(rm.items, item)
	}

	return nil
}
