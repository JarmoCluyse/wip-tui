package explorer

import (
	"github.com/jarmocluyse/git-dash/internal/git"
	"github.com/jarmocluyse/git-dash/internal/repository"
)

// Explorer provides backward compatibility interface for directory exploration.
type Explorer interface {
	ListDirectory(path string, repositories []repository.Repository) ([]Item, error)
	GetParentDirectory(path string) string
}

// Item represents a legacy filesystem item for backward compatibility.
type Item struct {
	Name           string
	Path           string
	IsDirectory    bool
	IsGitRepo      bool
	IsAdded        bool
	IsWorktree     bool
	WorktreeInfo   *git.WorktreeInfo
	HasUncommitted bool
	HasUnpushed    bool
	HasUntracked   bool
	HasError       bool
}

// LegacyExplorerAdapter adapts the modern DirectoryExplorer to the legacy Explorer interface.
type LegacyExplorerAdapter struct {
	modernExplorer DirectoryExplorer
}

// NewLegacyExplorerAdapter creates a new LegacyExplorerAdapter instance.
func NewLegacyExplorerAdapter(modernExplorer DirectoryExplorer) Explorer {
	return &LegacyExplorerAdapter{modernExplorer: modernExplorer}
}

// ListDirectory returns legacy Item structs for directory contents.
func (a *LegacyExplorerAdapter) ListDirectory(path string, repositories []repository.Repository) ([]Item, error) {
	entries, err := a.modernExplorer.ExploreDirectory(path, repositories)
	if err != nil {
		return nil, err
	}

	var items []Item
	for _, entry := range entries {
		item := a.convertToLegacyItem(entry)
		items = append(items, item)
	}

	return items, nil
}

// GetParentDirectory returns the parent directory path.
func (a *LegacyExplorerAdapter) GetParentDirectory(path string) string {
	return a.modernExplorer.NavigateToParent(path)
}

// convertToLegacyItem converts a FileSystemEntry to a legacy Item.
func (a *LegacyExplorerAdapter) convertToLegacyItem(entry FileSystemEntry) Item {
	repoInfo := entry.RepositoryInfo()

	return Item{
		Name:           entry.Name(),
		Path:           entry.Path(),
		IsDirectory:    entry.IsDirectory(),
		IsGitRepo:      repoInfo.IsGitRepository(),
		IsAdded:        repoInfo.IsAlreadyManaged(),
		IsWorktree:     repoInfo.IsWorktree(),
		WorktreeInfo:   repoInfo.WorktreeDetails(),
		HasUncommitted: repoInfo.HasUncommittedWork(),
		HasUnpushed:    repoInfo.HasUnpushedCommits(),
		HasUntracked:   repoInfo.HasUntrackedFiles(),
		HasError:       repoInfo.HasEncounteredError(),
	}
}
