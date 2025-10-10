# Git TUI - Repository Status Monitor

A clean, minimal TUI application for monitoring Git repository status across multiple repositories, with full support for bare repositories, worktrees, and a built-in folder explorer.

## Features

- **📁 Folder Explorer**: Browse your filesystem to discover and add Git repositories
- **Repository Management**: Add and remove Git repositories from your monitoring list
- **Bare Repository Support**: Full support for bare repositories and their worktrees
- **Worktree Auto-Discovery**: Automatically discover and add worktrees from bare repositories
- **Smart Detection**: Shows which repositories are already added in the explorer
- **Status Indicators**: 
  - `📁` Bare repository
  - `🌳` Worktree  
  - `●` Red dot indicates uncommitted changes
  - `↑` Yellow arrow indicates unpushed commits
  - `?` Orange question mark indicates untracked files (WIP not added to git)
  - `✗` Red X indicates invalid repository (folder is not a git repo)
  - `✓` Green checkmark indicates clean repository
- **Keyboard Navigation**: Vim-style navigation with hjkl keys
- **Real-time Updates**: Refresh repository status with a single key press

## Usage

### Installation
```bash
go build .
```

### Running
```bash
./git-tui
```

### Repository Discovery Workflow

1. **Open Explorer**: Press `e` to open the folder explorer
2. **Navigate**: Use arrow keys or hjkl to browse directories
3. **Add Repositories**: Press `Space` on Git repositories to toggle them
4. **Already Added**: Repositories show `✓` if already monitored, `○` if not
5. **Return**: Press `Esc` to return to the main view

### Bare Repository Workflow

1. **Add your bare repository**: Use explorer or manual add
2. **Discover worktrees**: Select the bare repo and press `w` to auto-discover all worktrees
3. **Monitor status**: All worktrees will now be monitored for uncommitted/unpushed changes

### Controls

**List View:**
- `↑/k`: Move cursor up
- `↓/j`: Move cursor down  
- `a`: Add new repository (manual input)
- `e`: Open folder explorer
- `w`: Discover worktrees from selected bare repository
- `d`: Delete selected repository
- `r`: Refresh all repository statuses
- `q`: Quit application

**Explorer View:**
- `↑/k`: Move cursor up
- `↓/j`: Move cursor down
- `Enter`: Navigate into directory
- `Space`: Toggle Git repository (add/remove from monitoring)
- `Esc/q`: Return to list view

**Add Repository View:**
- Type repository path
- `Enter`: Add repository
- `Esc`: Cancel and return to list

## Icons & Indicators

**Explorer Icons:**
- `📁` Directory
- `🔗` Git repository
- `📄` Regular file
- `✓` Repository already added
- `○` Repository not added

**Status Indicators:**
- `📁` Bare repository
- `🌳` Worktree
- `●` Uncommitted changes
- `↑` Unpushed commits
- `✓` Clean repository

## Bare Repository + Worktree Support

This TUI is designed with bare repository workflows in mind:

- **Bare Repository Detection**: Automatically detects bare repos using `git rev-parse --is-bare-repository`
- **Worktree Enumeration**: Uses `git worktree list --porcelain` to discover all worktrees
- **Individual Worktree Status**: Each worktree is monitored independently for changes
- **Unified Management**: Manage your entire bare repo + worktrees setup from one interface

## Configuration

Repository configurations are stored in `~/.git-tui.json`

## Architecture

This application follows Clean Architecture principles:

- **Entities**: `Repository`, `Config` - Core business objects
- **Use Cases**: `ConfigService`, `GitStatusChecker` - Business logic interfaces  
- **Interface Adapters**: `FileConfigService`, `CommandLineGitChecker` - External system adapters
- **Frameworks**: Bubbletea TUI framework, file I/O

Dependencies flow inward, making the code testable and maintainable.