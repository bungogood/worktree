# worktree

[![Build](../../actions/workflows/build.yml/badge.svg)](../../actions/workflows/build.yml)

A CLI for creating, switching, and managing Git worktrees with fast shell navigation.

Work on multiple features in separate directories and switch between them instantly. Each worktree is an independent working directory linked to the same repository, so you can keep in-progress changes isolated without stashing or committing.

## Installation

```bash
go install github.com/bungogood/worktree@latest
```

If `worktree` is not found, add Go's bin directory to your `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

### Full Bash Setup (recommended)

Run once to install the `wrk` shell shim and persistent bash completion:

```bash
mkdir -p "$HOME/.local/scripts" "$HOME/.local/share/bash-completion/completions"
worktree init bash > "$HOME/.local/scripts/wrk-shim.sh"
worktree completion bash > "$HOME/.local/share/bash-completion/completions/worktree"
ln -sf "$HOME/.local/share/bash-completion/completions/worktree" "$HOME/.local/share/bash-completion/completions/wrk"
```

Add this line to `~/.bashrc`:

```bash
source "$HOME/.local/scripts/wrk-shim.sh"
```

Temporary session only (no files written):

```bash
eval "$(worktree init bash)"
source <(worktree completion bash)
```

## Usage

Use `wrk` for interactive commands that switch directories, or `worktree` for scripting (`wrk` is to `worktree` what `z` is to `zoxide`).

```bash
# Create and jump to a new worktree
wrk new feature/my-change

# Switch between worktrees
wrk switch
wrk switch feature/*

# List and remove
wrk list
wrk rm feature/my-change
```

All commands are also available through `worktree`.

### Command List

```
A CLI tool for managing git worktrees with automatic organisation and navigation.

Usage:
  wrk [command]

Available Commands:
  add         Add an existing branch as a worktree
  completion  Generate shell completion script
  copy        Copy files from another worktree
  exclude     Manage excluded untracked files
  help        Help about any command
  list        List all worktrees
  new         Create a new branch as a worktree
  remove      Remove worktrees
  skip        Manage skipped file changes
  switch      Switch to a worktree

Flags:
  -h, --help      help for worktree
  -v, --verbose   Show all git commands being executed
```

## Examples

```bash
# List all worktrees
wrk list

# Create new worktree from new branch
wrk new feature-branch
wrk new feature-branch custom-worktree-name

# Add worktree from existing branch
wrk add existing-branch
wrk add existing-branch another-worktree-name
wrk add existing-branch --remote upstream  # Create tracking branch from remote

# Switch to a worktree (changes directory)
wrk switch  # Switch to main worktree
wrk switch feature-branch
wrk switch JIRA-123-*  # Glob pattern matching

# Remove worktrees
wrk rm # Removes current worktree and switches to main worktree
wrk rm feature-branch
wrk rm branch-1 branch-2 branch-3
wrk rm -D feature-branch  # Deletes branch

# Skip file changes across all worktrees
wrk skip  # Lists skipped files
wrk skip config/local.json
wrk skip --rm config/local.json
wrk skip --local file.txt  # Only current worktree

# Exclude files from git (uses .git/info/exclude)
wrk exclude  # Lists excluded files
wrk exclude build/
wrk exclude --rm build/

# Copy files between worktrees
wrk copy config/settings.json  # Copy from main worktree
wrk copy --from feature-branch src/utils.go src/

# Always-copy: automatically copy to new worktrees
wrk copy --always  # List always-copy paths
wrk copy --always .env  # Add to always-copy
wrk copy --always-rm .env  # Remove from always-copy
```

## Configuration

Optional config is stored in `.{repo}.worktrees/.config.yml`:

```yaml
# Automatically delete the branch when removing a worktree
deleteBranchWithWorktree: true

# Files to automatically copy to new worktrees
copy:
    - .env
    - config/local.json

# Commands to run after creating new worktrees
commands:
    - npm install
    - go mod download
```

## About

### Why `wrk` and `worktree`?

This tool provides both a binary (`worktree`) and a bash wrapper function (`wrk`). The wrapper is required for directory switching, as processes cannot change their parent shell's working directory.

The `wrk` function intercepts the output from the `worktree` binary and automatically executes `cd` commands when switching between worktrees, for simple navigation.

### Worktree Organization

All worktrees are created in a `.{repo-name}.worktrees` directory at the repository root, keeping your workspace organised.

```
projects/
├── .my-repo.worktrees/
│   ├── .config.yml (optional stores wrk config)
│   ├── another-worktree-name/
│   └── feature-branch/
└── my-repo/
```

### Glob Pattern Matching

Commands like `switch` and `remove` support glob patterns for matching worktrees by name or branch:

-   `*` - Matches any characters (e.g., `JIRA-*` matches all worktrees starting with "JIRA-")
-   `?` - Matches a single character (e.g., `test-?` matches "test-1", "test-a", etc.)
-   `[...]` - Matches character ranges (e.g., `feature-[0-9]*` matches "feature-1", "feature-123", etc.)

Patterns match against both worktree **names** and **branch names**. **Tab completion** expands glob patterns to show matching worktrees.

### Skip-Worktree Strategy

When skipping files, the main worktree retains the original files while other worktrees use symlinks pointing to the main worktree. This ensures consistency across all worktrees while also allowing local modifications when needed.

## References

-   [zoxide](https://github.com/ajeetdsouza/zoxide) - Inspiration for the shell integration and directory switching pattern
-   [brachlet](https://github.com/raghavpillai/branchlet) - A similar typescript tool for managing git branches and worktrees
-   [git-worktree](https://git-scm.com/docs/git-worktree) - Git's official worktree documentation
-   [git-update-index](https://git-scm.com/docs/git-update-index) - Documentation for skip-worktree functionality
