// Package explorer provides filesystem navigation with Git repository detection and worktree support.
package explorer

import "github.com/jarmocluyse/git-dash/internal/git"

// FileSystemEntry represents a single filesystem entry with Git repository information.
type FileSystemEntry struct {
	name           string
	absolutePath   string
	isDirectory    bool
	repositoryInfo RepositoryInfo
}

// Name returns the display name of the filesystem entry.
func (e FileSystemEntry) Name() string {
	return e.name
}

// Path returns the absolute path of the filesystem entry.
func (e FileSystemEntry) Path() string {
	return e.absolutePath
}

// IsDirectory returns true if this entry represents a directory.
func (e FileSystemEntry) IsDirectory() bool {
	return e.isDirectory
}

// RepositoryInfo returns the Git repository information for this entry.
func (e FileSystemEntry) RepositoryInfo() RepositoryInfo {
	return e.repositoryInfo
}

// NewFileSystemEntry creates a new immutable FileSystemEntry.
func NewFileSystemEntry(name, path string, isDir bool, repoInfo RepositoryInfo) FileSystemEntry {
	return FileSystemEntry{
		name:           name,
		absolutePath:   path,
		isDirectory:    isDir,
		repositoryInfo: repoInfo,
	}
}

// RepositoryInfo contains Git repository status and metadata for a filesystem entry.
type RepositoryInfo struct {
	isGitRepository     bool
	isAlreadyManaged    bool
	isWorktree          bool
	worktreeDetails     *git.WorktreeInfo
	hasUncommittedWork  bool
	hasUnpushedCommits  bool
	hasUntrackedFiles   bool
	hasEncounteredError bool
}

// IsGitRepository returns true if this entry is a Git repository.
func (r RepositoryInfo) IsGitRepository() bool {
	return r.isGitRepository
}

// IsAlreadyManaged returns true if this repository is already being managed by the application.
func (r RepositoryInfo) IsAlreadyManaged() bool {
	return r.isAlreadyManaged
}

// IsWorktree returns true if this entry represents a Git worktree.
func (r RepositoryInfo) IsWorktree() bool {
	return r.isWorktree
}

// WorktreeDetails returns the detailed worktree information, or nil if not a worktree.
func (r RepositoryInfo) WorktreeDetails() *git.WorktreeInfo {
	return r.worktreeDetails
}

// HasUncommittedWork returns true if the repository has uncommitted changes.
func (r RepositoryInfo) HasUncommittedWork() bool {
	return r.hasUncommittedWork
}

// HasUnpushedCommits returns true if the repository has unpushed commits.
func (r RepositoryInfo) HasUnpushedCommits() bool {
	return r.hasUnpushedCommits
}

// HasUntrackedFiles returns true if the repository has untracked files.
func (r RepositoryInfo) HasUntrackedFiles() bool {
	return r.hasUntrackedFiles
}

// HasEncounteredError returns true if an error occurred while checking repository status.
func (r RepositoryInfo) HasEncounteredError() bool {
	return r.hasEncounteredError
}

// NewRepositoryInfo creates a new RepositoryInfo with default values.
func NewRepositoryInfo() RepositoryInfo {
	return RepositoryInfo{}
}

// WithGitRepository sets whether this entry is a Git repository.
func (r RepositoryInfo) WithGitRepository(isRepo bool) RepositoryInfo {
	r.isGitRepository = isRepo
	return r
}

// WithManagedStatus sets whether this repository is already being managed.
func (r RepositoryInfo) WithManagedStatus(isManaged bool) RepositoryInfo {
	r.isAlreadyManaged = isManaged
	return r
}

// WithWorktree sets the worktree status and details.
func (r RepositoryInfo) WithWorktree(isWorktree bool, details *git.WorktreeInfo) RepositoryInfo {
	r.isWorktree = isWorktree
	r.worktreeDetails = details
	return r
}

// WithGitStatus sets the Git status flags for uncommitted, unpushed, and untracked changes.
func (r RepositoryInfo) WithGitStatus(uncommitted, unpushed, untracked bool) RepositoryInfo {
	r.hasUncommittedWork = uncommitted
	r.hasUnpushedCommits = unpushed
	r.hasUntrackedFiles = untracked
	return r
}

// WithError sets whether an error was encountered while checking repository status.
func (r RepositoryInfo) WithError(hasError bool) RepositoryInfo {
	r.hasEncounteredError = hasError
	return r
}
