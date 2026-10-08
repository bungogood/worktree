# wrk

[![Build](../../actions/workflows/build.yml/badge.svg)](../../actions/workflows/build.yml)

A CLI for creating, switching, and managing Git worktrees with fast shell navigation.

Work on multiple features in separate directories and switch between them instantly. Each worktree is an independent working directory linked to the same repository, so you can keep in-progress changes isolated without stashing or committing.

![Worktree demo](docs/worktree-demo.gif)

## Features

- Create and jump to new or existing branch worktrees in one command
- Shell-native directory switching with `wrk` (like `z` for `zoxide`)
- Glob-aware worktree selection with tab completion (`wrk switch feature/*`)
- In-place branch switching without changing directory (`wrk switch -b`)
- Status overview with dirty state and ahead/behind per worktree (`wrk status`)
- Cleanup of stale worktrees by last-commit age (`wrk clean`)
- Built-in skip/exclude/copy workflows to reduce repeated setup across worktrees

## Installation

```bash
go install github.com/bungogood/wrk@latest
eval "$(wrk hook bash)"
eval "$(wrk completion bash)"
```

The binary is `wrk`.

If `wrk` is not found, add Go's bin directory to your `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

## Usage

Use `wrk` for everything, including interactive commands that switch directories (`wrk` is to plain directory switching what `z` is to `zoxide`).

```bash
# Create and jump to a new worktree
wrk new feature/my-change

# Switch between worktrees
wrk switch
wrk switch feature/*

# List, check status, and remove
wrk list
wrk status
wrk rm feature/my-change
```

Use `wrk --help` for full command details.

## Examples

```bash
# Create a worktree from a new branch
wrk new feature-branch
wrk new feature-branch custom-worktree-name

# Add a worktree from an existing branch
wrk add existing-branch
wrk add existing-branch another-worktree-name
wrk add existing-branch --remote upstream  # Create tracking branch from remote

# Switch to a worktree (changes directory)
wrk switch  # Switch to main worktree
wrk switch feature-branch
wrk switch JIRA-123-*  # Glob pattern matching

# Switch branch in place (no directory change)
wrk switch -b  # Main worktree: remote default branch, linked: branch matching directory name
wrk switch -b feature-branch

# Status overview (dirty state, ahead/behind, last-commit age)
wrk status

# Remove worktrees
wrk rm  # Removes current worktree and switches to main worktree
wrk rm feature-branch
wrk rm branch-1 branch-2 branch-3
wrk rm -D feature-branch  # Deletes branch

# Clean up stale worktrees (no commits for longer than the threshold)
wrk clean --dry-run  # Preview (also try -o 4w, -f for dirty worktrees)
wrk clean

# Skip file changes across all worktrees
wrk skip  # List skipped files
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

## How It Works

### `wrk` function vs `wrk` binary

This tool provides both a binary (`wrk`) and a bash wrapper function (also `wrk`, shadowing the binary in interactive shells). The wrapper is required for directory switching, as processes cannot change their parent shell's working directory.

The `wrk` function intercepts the output from the `wrk` binary and automatically executes `cd` commands when switching between worktrees, for simple navigation.

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

- `*` matches any characters (for example, `JIRA-*`)
- `?` matches a single character (for example, `test-?`)
- `[...]` matches character ranges (for example, `feature-[0-9]*`)

Patterns match against both worktree **names** and **branch names**. **Tab completion** expands glob patterns to show matching worktrees.

### Skip-Worktree Strategy

When skipping files, the main worktree retains the original files while other worktrees use symlinks pointing to the main worktree. This ensures consistency across all worktrees while also allowing local modifications when needed.

## Development

Run tests:

```bash
go test ./...
```

## References

- [zoxide](https://github.com/ajeetdsouza/zoxide) - Inspiration for shell integration and directory switching
- [branchlet](https://github.com/raghavpillai/branchlet) - Similar tool for branch and worktree workflows
- [git-worktree](https://git-scm.com/docs/git-worktree) - Git's official worktree documentation
- [git-update-index](https://git-scm.com/docs/git-update-index) - Documentation for skip-worktree behavior

## License

[MIT](LICENSE)
