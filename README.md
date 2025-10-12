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

### Command Line Options

```bash
./git-tui [options]
```

**Options:**
- `-c, --config <path>`: Specify a custom configuration file path
- `-h, --help`: Show help information
- `-v, --version`: Show version information

**Examples:**
```bash
./git-tui                              # Use default config (~/.git-tui.toml)
./git-tui -c ~/.config/git-tui.toml    # Use custom config file
./git-tui --config /path/to/config.toml
./git-tui --help                       # Show detailed help
./git-tui --version                    # Show version information
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
- `l`: Open repository in Lazygit (configurable)
- `c`: Open repository in VS Code (configurable)
- `t`: Open terminal in repository directory (configurable)
- `Enter`: View repository details
- `q`: Quit application
- `?`: Show help modal

**Note:** The `l`, `c`, and `t` actions are configurable through the config file. See the [Configurable Actions](#configurable-actions) section below.

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

Configuration is stored in TOML format. The default location is `~/.git-tui.toml`, but you can specify a custom location using the `--config` flag.

**Default Configuration Locations (in order of priority):**
1. Path specified with `--config` or `-c` flag
2. Path specified in `GIT_TUI_CONFIG` environment variable  
3. `~/.git-tui.toml` (default)

**Example Configuration:**
```toml
repository_paths = [
    "/path/to/repo1",
    "/path/to/repo2"
]

# Configurable keybindings for repository actions
[[keybindings.actions]]
name = "Lazygit"
key = "l"
command = "lazygit"
args = ["-p", "{path}"]
description = "Open repository in Lazygit"

[[keybindings.actions]]
name = "VS Code"
key = "c"
command = "code"
args = ["{path}"]
description = "Open repository in VS Code"

[[keybindings.actions]]
name = "Terminal"
key = "t"
command = "gnome-terminal"
args = ["--working-directory={path}"]
description = "Open terminal in repository directory"

[theme.colors]
title = "#FF6B6B"
title_background = "#000000"
selected = "#4ECDC4"
status_dirty = "#FFE66D"
status_unpushed = "#FF6B6B"
status_untracked = "#95E1D3"
status_error = "#F38BA8"
status_clean = "#A8E6CF"
status_not_added = "#FFB6C1"
help = "#B0B0B0"
border = "#4A4A4A"
modal_background = "#1A1A1A"

[theme.indicators]
clean = "✨"
dirty = "🔥"
unpushed = "⬆️"
untracked = "❓"
error = "💥"
not_added = "➕"
```

**Customizable Theme Options:**
- **Colors**: All UI colors can be customized using hex color codes
- **Indicators**: Status indicators can be customized with any Unicode characters/emojis

### Configurable Actions

You can configure custom keybindings to open repositories in your preferred tools. Actions are defined in the `[keybindings]` section of your config file.

**Default Actions:**
- `l` - Open in Lazygit
- `c` - Open in VS Code  
- `t` - Open terminal in repository directory

**Example Configuration:**
```toml
[[keybindings.actions]]
name = "Lazygit"
key = "l"
command = "lazygit"
args = ["-p", "{path}"]
description = "Open repository in Lazygit"

[[keybindings.actions]]
name = "VS Code"
key = "c" 
command = "code"
args = ["{path}"]
description = "Open repository in VS Code"

[[keybindings.actions]]
name = "Terminal"
key = "t"
command = "gnome-terminal"
args = ["--working-directory={path}"]
description = "Open terminal in repository directory"

# Add your own custom actions
[[keybindings.actions]]
name = "GitHub Desktop"
key = "g"
command = "github-desktop"
args = ["{path}"]
description = "Open repository in GitHub Desktop"

[[keybindings.actions]]
name = "Custom Script"
key = "s"
command = "/path/to/your/script.sh"
args = ["{path}", "--verbose"]
description = "Run custom script on repository"
```

**Configuration Details:**
- `name`: Display name for the action
- `key`: Single character key binding (avoid conflicts with built-in keys)
- `command`: Command to execute
- `args`: Array of command arguments
- `description`: Help text description
- `{path}`: Placeholder that gets replaced with the actual repository path

**Built-in Keys to Avoid:**
- Navigation: `↑`, `↓`, `j`, `k`, `h`, `l` (if you want vim-style navigation)
- Actions: `a`, `e`, `w`, `d`, `r`, `q`, `?`, `Enter`, `Esc`, `Space`

The help text at the bottom of the screen will automatically update to show your configured actions.

## Architecture

This application follows Clean Architecture principles:

- **Entities**: `Repository`, `Config` - Core business objects
- **Use Cases**: `ConfigService`, `GitStatusChecker` - Business logic interfaces  
- **Interface Adapters**: `FileConfigService`, `CommandLineGitChecker` - External system adapters
- **Frameworks**: Bubbletea TUI framework, file I/O

Dependencies flow inward, making the code testable and maintainable.