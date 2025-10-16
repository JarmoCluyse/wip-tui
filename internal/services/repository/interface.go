// Package repository provides repository management services for the application.
package repository

import (
	"github.com/jarmocluyse/git-dash/internal/git"
	"github.com/jarmocluyse/git-dash/internal/repository"
)

// Service defines the interface for repository management operations.
type Service interface {
	// Repository Management
	LoadRepositories(paths []string) error
	GetRepositories() []repository.Repository
	AddRepository(name, path string) error
	RemoveRepository(index int) error
	RemoveRepositoryByPath(path string) error
	GetRepositoryPaths() []string

	// Repository Status Operations
	UpdateRepositoryStatus(index int) error
	UpdateAllRepositoryStatuses() error

	// Navigation Items
	GetNavigableItems() ([]repository.NavigableItem, error)
	RefreshNavigableItems() error

	// Repository Information
	GetRepositoryByIndex(index int) (*repository.Repository, error)
	GetRepositoryByPath(path string) (*repository.Repository, error)
	GetRepositoryCount() int

	// Summary Calculations
	GetRepositorySummary() repository.SummaryData
	GetNavigableItemSummary() repository.SummaryData

	// Validation
	IsValidRepositoryIndex(index int) bool
	ContainsRepository(path string) bool

	// Internal access for rendering
	GetGitChecker() git.StatusChecker
}
