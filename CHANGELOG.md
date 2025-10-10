# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Enhanced worktree navigation functionality for Git TUI application
- Automatic detection and display of worktrees under bare repositories  
- Tree-style indentation display for worktrees with `├─` visual indicators
- Interactive navigation support for both repositories and worktrees using Enter key
- Real-time Git status indicators for worktrees (●↑?✓✗)
- `NavigableItem` type for unified repository and worktree handling
- Relative path display for worktrees with proper icons (🌳 for worktrees, 📁 for bare repos)
- Action handling that properly distinguishes between repositories and worktrees (delete only works on repos)
- **Lazygit integration**: Press `l` to open selected repository or worktree in lazygit
- Lazygit support in both main view and explorer view for seamless Git operations

### Changed
- Modified explorer view to automatically detect worktrees under bare repositories
- Updated status page to show worktrees when bare repositories are selected
- Enhanced cursor navigation to work with flattened repository and worktree list
- Improved display formatting with clean tree-style layout

### Technical Details
- Added `buildNavigableItems()` function for unified repo/worktree handling
- Implemented `navigateToSelected()` method for Enter key navigation
- Added `RenderNavigable()` method to `ListViewRenderer`
- Enhanced `getWorktreeItems()` function in explorer with worktree status checking
- Updated cursor navigation and action handling throughout the application

## [0.1.0] - Initial Release

### Added
- Basic Git repository management TUI
- Repository status monitoring with visual indicators
- Explorer view for browsing directories
- Configuration management for tracked repositories
- Support for bare repositories
- Basic worktree discovery functionality
- Logging system for debugging
- Manual refresh capability (press 'r')

### Features
- Add repositories to tracking list
- Navigate through repositories and directories
- Visual status indicators for Git repository states
- Delete repositories from tracking
- Explore repository contents
- Discover worktrees for bare repositories (press 'w')