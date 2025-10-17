package repomanager

// AddRepo adds a new repository by path.
func (rm *RepoManager) AddRepo(path string) error {
	// Check if already exists
	for _, item := range rm.items {
		if item.Path == path {
			return nil // Already exists
		}
	}
	// Create new repo item
	item := &RepoItem{
		Name:     extractNameFromPath(path),
		Path:     path,
		SubItems: make([]*SubItem, 0),
	}

	rm.updateRepoStatus(item)
	if item.IsBare {
		rm.loadWorktrees(item)
	}

	rm.items = append(rm.items, item)

	config, err := rm.configService.Load()
	if err != nil {
		return err
	}

	config.AddRepositoryPath(path)
	return rm.configService.Save(config)
}

// RemoveRepo removes a repository by path.
func (rm *RepoManager) RemoveRepo(path string) error {
	for i, item := range rm.items {
		if item.Path == path {
			rm.items = append(rm.items[:i], rm.items[i+1:]...)
			break
		}
	}
	config, err := rm.configService.Load()
	if err != nil {
		return err
	}
	config.RemoveRepositoryPathByValue(path)
	return rm.configService.Save(config)
}
