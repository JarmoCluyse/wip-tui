package explorer

import "github.com/jarmocluyse/git-dash/internal/git"

// New creates a new Explorer instance with backward compatibility.
func New(gitChecker git.StatusChecker, config any) Explorer {
	modernExplorer := NewCleanFileSystemExplorer(gitChecker)
	return NewLegacyExplorerAdapter(modernExplorer)
}
