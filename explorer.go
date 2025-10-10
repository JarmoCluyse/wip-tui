package main

import (
	"os"
	"path/filepath"
	"strings"
)

type DirectoryExplorer interface {
	ListDirectory(path string, repositories []Repository) ([]ExplorerItem, error)
	GetParentDirectory(path string) string
}

type FileSystemExplorer struct {
	gitChecker GitStatusChecker
}

func NewDirectoryExplorer(gitChecker GitStatusChecker, config *Config) DirectoryExplorer {
	return &FileSystemExplorer{
		gitChecker: gitChecker,
	}
}

func (f *FileSystemExplorer) ListDirectory(path string, repositories []Repository) ([]ExplorerItem, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var items []ExplorerItem

	if path != "/" {
		parentItem := ExplorerItem{
			Name:        "..",
			Path:        f.GetParentDirectory(path),
			IsDirectory: true,
			IsGitRepo:   false,
			IsAdded:     false,
		}
		items = append(items, parentItem)
	}

	for _, entry := range entries {
		if f.shouldSkipEntry(entry.Name()) {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		item := ExplorerItem{
			Name:        entry.Name(),
			Path:        fullPath,
			IsDirectory: entry.IsDir(),
			IsGitRepo:   f.isGitRepository(fullPath),
			IsAdded:     f.isAlreadyAdded(fullPath, repositories),
		}

		// Check if this is a bare repository and add its worktrees
		if item.IsGitRepo && f.gitChecker.IsBareRepository(fullPath) {
			worktreeItems := f.getWorktreeItems(fullPath, repositories)
			items = append(items, item)
			items = append(items, worktreeItems...)
		} else {
			items = append(items, item)
		}
	}

	return items, nil
}

func (f *FileSystemExplorer) GetParentDirectory(path string) string {
	parent := filepath.Dir(path)
	if parent == path {
		return "/"
	}
	return parent
}

func (f *FileSystemExplorer) shouldSkipEntry(name string) bool {
	return strings.HasPrefix(name, ".")
}

func (f *FileSystemExplorer) isGitRepository(path string) bool {
	return f.gitChecker.IsGitRepository(path)
}

func (f *FileSystemExplorer) isAlreadyAdded(path string, repositories []Repository) bool {
	// Normalize the path for comparison
	cleanPath := filepath.Clean(path)

	for _, repo := range repositories {
		cleanRepoPath := filepath.Clean(repo.Path)
		if cleanRepoPath == cleanPath {
			return true
		}
	}
	return false
}

func (f *FileSystemExplorer) getWorktreeItems(bareRepoPath string, repositories []Repository) []ExplorerItem {
	worktrees, err := f.gitChecker.ListWorktrees(bareRepoPath)
	if err != nil {
		return nil
	}

	var items []ExplorerItem
	for _, wt := range worktrees {
		// Skip the bare repository itself
		if wt.Bare || wt.Path == bareRepoPath {
			continue
		}

		// Create a display name for the worktree
		baseName := filepath.Base(bareRepoPath)
		worktreeName := baseName + "-" + wt.Branch

		item := ExplorerItem{
			Name:         "  ├─ " + worktreeName,
			Path:         wt.Path,
			IsDirectory:  true,
			IsGitRepo:    true,
			IsAdded:      f.isAlreadyAdded(wt.Path, repositories),
			IsWorktree:   true,
			WorktreeInfo: &wt,
		}

		// Update status for this worktree
		f.updateWorktreeStatus(&item)
		items = append(items, item)
	}

	return items
}

func (f *FileSystemExplorer) updateWorktreeStatus(item *ExplorerItem) {
	if !f.gitChecker.IsGitRepository(item.Path) {
		item.HasError = true
		return
	}

	item.HasUncommitted = f.gitChecker.HasUncommittedChanges(item.Path)
	item.HasUnpushed = f.gitChecker.HasUnpushedCommits(item.Path)
	item.HasUntracked = f.gitChecker.HasUntrackedFiles(item.Path)
}
