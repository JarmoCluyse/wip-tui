package explorer

import (
	"os"
	"path/filepath"

	"github.com/jarmocluyse/git-dash/internal/git"
	"github.com/jarmocluyse/git-dash/internal/repository"
)

// CleanFileSystemExplorer implements directory exploration with git repository awareness.
type CleanFileSystemExplorer struct {
	repositoryService *RepositoryStatusService
	worktreeService   *WorktreeService
}

// NewCleanFileSystemExplorer creates a new CleanFileSystemExplorer instance.
func NewCleanFileSystemExplorer(gitChecker git.StatusChecker) DirectoryExplorer {
	repositoryService := NewRepositoryStatusService(gitChecker)
	worktreeService := NewWorktreeService(repositoryService)

	return &CleanFileSystemExplorer{
		repositoryService: repositoryService,
		worktreeService:   worktreeService,
	}
}

// ExploreDirectory returns filesystem entries for the given directory path.
func (e *CleanFileSystemExplorer) ExploreDirectory(directoryPath string, managedRepositories []repository.Repository) ([]FileSystemEntry, error) {
	entries, err := e.readDirectoryEntries(directoryPath)
	if err != nil {
		return nil, err
	}

	var fileSystemEntries []FileSystemEntry

	if e.shouldIncludeParentDirectory(directoryPath) {
		parentEntry := e.createParentDirectoryEntry(directoryPath)
		fileSystemEntries = append(fileSystemEntries, parentEntry)
	}

	for _, entry := range entries {
		if e.shouldIncludeEntry(entry) {
			entryEntries := e.processDirectoryEntry(entry, directoryPath, managedRepositories)
			fileSystemEntries = append(fileSystemEntries, entryEntries...)
		}
	}

	return fileSystemEntries, nil
}

// NavigateToParent returns the parent directory path for the given path.
func (e *CleanFileSystemExplorer) NavigateToParent(currentPath string) string {
	return createParentDirectoryPath(currentPath)
}

// readDirectoryEntries reads the contents of a directory.
func (e *CleanFileSystemExplorer) readDirectoryEntries(directoryPath string) ([]os.DirEntry, error) {
	return os.ReadDir(directoryPath)
}

// shouldIncludeParentDirectory determines if the parent directory entry should be included.
func (e *CleanFileSystemExplorer) shouldIncludeParentDirectory(currentPath string) bool {
	return currentPath != "/"
}

// createParentDirectoryEntry creates a ".." entry for navigating to the parent directory.
func (e *CleanFileSystemExplorer) createParentDirectoryEntry(currentPath string) FileSystemEntry {
	parentPath := e.NavigateToParent(currentPath)
	emptyRepoInfo := NewRepositoryInfo()

	return NewFileSystemEntry("..", parentPath, true, emptyRepoInfo)
}

// shouldIncludeEntry determines if a directory entry should be included in the results.
func (e *CleanFileSystemExplorer) shouldIncludeEntry(entry os.DirEntry) bool {
	return !shouldSkipHiddenEntry(entry.Name())
}

// processDirectoryEntry processes a single directory entry and returns all associated filesystem entries.
func (e *CleanFileSystemExplorer) processDirectoryEntry(entry os.DirEntry, basePath string, managedRepositories []repository.Repository) []FileSystemEntry {
	entryPath := filepath.Join(basePath, entry.Name())
	mainEntry := e.createFileSystemEntry(entry, entryPath, managedRepositories)

	entries := []FileSystemEntry{mainEntry}

	if e.shouldIncludeWorktrees(mainEntry, entryPath) {
		worktreeEntries := e.createWorktreeEntries(entryPath, managedRepositories)
		entries = append(entries, worktreeEntries...)
	}

	return entries
}

// createFileSystemEntry creates a FileSystemEntry from an os.DirEntry.
func (e *CleanFileSystemExplorer) createFileSystemEntry(entry os.DirEntry, entryPath string, managedRepositories []repository.Repository) FileSystemEntry {
	repositoryInfo := e.buildRepositoryInfo(entryPath, managedRepositories)
	return NewFileSystemEntry(entry.Name(), entryPath, entry.IsDir(), repositoryInfo)
}

// buildRepositoryInfo creates repository information for a filesystem entry.
func (e *CleanFileSystemExplorer) buildRepositoryInfo(entryPath string, managedRepositories []repository.Repository) RepositoryInfo {
	isGitRepo := e.repositoryService.IsGitRepository(entryPath)
	isManaged := e.repositoryService.IsAlreadyManaged(entryPath, managedRepositories)

	return NewRepositoryInfo().
		WithGitRepository(isGitRepo).
		WithManagedStatus(isManaged)
}

// shouldIncludeWorktrees determines if worktree entries should be included for a git repository.
func (e *CleanFileSystemExplorer) shouldIncludeWorktrees(entry FileSystemEntry, entryPath string) bool {
	return entry.RepositoryInfo().IsGitRepository() && e.repositoryService.IsBareRepository(entryPath)
}

// createWorktreeEntries creates filesystem entries for git worktrees.
func (e *CleanFileSystemExplorer) createWorktreeEntries(bareRepoPath string, managedRepositories []repository.Repository) []FileSystemEntry {
	worktreeEntries, err := e.worktreeService.CreateWorktreeEntries(bareRepoPath, managedRepositories)
	if err != nil {
		return nil
	}
	return worktreeEntries
}
