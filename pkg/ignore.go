package pkg

import (
	"fmt"
	"strings"
)

// IgnoreFile marks a file to have its changes ignored using git skip-worktree
func (r *Repo) IgnoreFile(file string) error {
	_, err := r.RunGitCommand(nil, "update-index", "--skip-worktree", file)
	if err != nil {
		return fmt.Errorf("failed to ignore file: %w", err)
	}
	return nil
}

// UnignoreFile unmarks a file so its changes will be tracked again
func (r *Repo) UnignoreFile(file string) error {
	_, err := r.RunGitCommand(nil, "update-index", "--no-skip-worktree", file)
	if err != nil {
		return fmt.Errorf("failed to unignore file: %w", err)
	}
	return nil
}

// ListIgnoredFiles returns a list of files marked with skip-worktree
func (r *Repo) ListIgnoredFiles() ([]string, error) {
	output, err := r.RunGitCommand(nil, "ls-files", "-v")
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	ignored := make([]string, 0)

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		// Files with skip-worktree flag show as 'S' in the first column
		if strings.HasPrefix(line, "S ") {
			file := strings.TrimPrefix(line, "S ")
			ignored = append(ignored, file)
		}
	}

	return ignored, nil
}
