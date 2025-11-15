package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IgnoreFile marks a file to have its changes ignored using git skip-worktree
// across all worktrees. For non-main worktrees, it symlinks the file to the
// main worktree's version.
func (r *Repo) IgnoreFile(file string) error {
	// Get the absolute path from the main worktree
	mainFilePath := filepath.Join(r.MainWorktree.Path, file)

	// Check if file exists in main worktree
	if _, err := os.Stat(mainFilePath); err != nil {
		return fmt.Errorf("file not found in main worktree: %w", err)
	}

	// Check if file is tracked in git in the main worktree
	_, err := r.RunGitCommand(r.MainWorktree, "ls-files", "--error-unmatch", file)
	if err != nil {
		return fmt.Errorf("file is not tracked by git in main worktree: %s", file)
	}

	var errors []string

	// Process each worktree
	for _, wt := range r.Worktrees {
		wtFilePath := filepath.Join(wt.Path, file)

		// Check if file exists in this worktree (either as file or symlink)
		if _, err := os.Lstat(wtFilePath); err != nil {
			if os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("worktree %s: file does not exist", wt.Name))
				continue
			}
			errors = append(errors, fmt.Sprintf("worktree %s: failed to check file: %v", wt.Name, err))
			continue
		}

		// Run git update-index --skip-worktree in this worktree
		_, err := r.RunGitCommand(&wt, "update-index", "--skip-worktree", file)
		if err != nil {
			errors = append(errors, fmt.Sprintf("worktree %s: failed to skip-worktree: %v", wt.Name, err))
			continue
		}

		// If this is not the main worktree, replace file with symlink to main
		if !r.IsMainWorktree(&wt) {
			// Remove existing file
			if err := os.Remove(wtFilePath); err != nil {
				errors = append(errors, fmt.Sprintf("worktree %s: failed to remove file: %v", wt.Name, err))
				continue
			}

			// Create symlink to main worktree's file
			if err := os.Symlink(mainFilePath, wtFilePath); err != nil {
				errors = append(errors, fmt.Sprintf("worktree %s: failed to create symlink: %v", wt.Name, err))
				continue
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed in some worktrees:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// UnignoreFile unmarks a file so its changes will be tracked again across all
// worktrees. It removes the skip-worktree flag and restores the file to its
// branch-specific version.
func (r *Repo) UnignoreFile(file string) error {
	var errors []string

	// Process each worktree
	for _, wt := range r.Worktrees {
		wtFilePath := filepath.Join(wt.Path, file)

		// Run git update-index --no-skip-worktree
		_, err := r.RunGitCommand(&wt, "update-index", "--no-skip-worktree", file)
		if err != nil {
			errors = append(errors, fmt.Sprintf("worktree %s: failed to unignore: %v", wt.Name, err))
			continue
		}

		// If this is not the main worktree, remove symlink and restore
		if !r.IsMainWorktree(&wt) {
			// Check if file is a symlink and remove it
			info, err := os.Lstat(wtFilePath)
			if err != nil && !os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("worktree %s: failed to check file: %v", wt.Name, err))
				continue
			}

			// Remove symlink if it exists
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				if err := os.Remove(wtFilePath); err != nil {
					errors = append(errors, fmt.Sprintf("worktree %s: failed to remove symlink: %v", wt.Name, err))
					continue
				}
			}

			// Restore the file to its branch-specific version
			// This will fail gracefully if the file doesn't exist in this branch
			output, err := r.RunGitCommand(&wt, "restore", file)
			if err != nil {
				// Check if it's because the file doesn't exist in this branch
				if strings.Contains(string(output), "did not match any file") ||
					strings.Contains(err.Error(), "did not match any file") {
					errors = append(errors, fmt.Sprintf("worktree %s: file does not exist in this branch", wt.Name))
				} else {
					errors = append(errors, fmt.Sprintf("worktree %s: failed to restore: %v", wt.Name, err))
				}
				continue
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed in some worktrees:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// ListIgnoredFiles returns a list of files marked with skip-worktree in the main worktree// ListIgnoredFiles returns a list of files marked with skip-worktree in the main worktree
func (r *Repo) ListIgnoredFiles() ([]string, error) {
	output, err := r.RunGitCommand(r.MainWorktree, "ls-files", "-v")
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
