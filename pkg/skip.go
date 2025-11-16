package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SkipFile marks a file to have its changes skipped using git skip-worktree
// across all worktrees. For non-main worktrees, it symlinks the file to the
// main worktree's version.
func (r *Repo) SkipFile(file string) error {
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

		// Skip the file in this worktree
		if err := r.skipFileInWorktree(&wt, file); err != nil {
			errors = append(errors, fmt.Sprintf("worktree %s: %v", wt.Name, err))
			continue
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed in some worktrees:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// UnskipFile unmarks a file so its changes will be tracked again across all
// worktrees. It removes the skip-worktree flag and restores the file to its
// branch-specific version.
func (r *Repo) UnskipFile(file string) error {
	var errors []string

	// Process each worktree
	for _, wt := range r.Worktrees {
		if err := r.unskipFileInWorktree(&wt, file); err != nil {
			errors = append(errors, fmt.Sprintf("worktree %s: %v", wt.Name, err))
			continue
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed in some worktrees:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// LocalSkipFile marks a file to have its changes skipped only in the current worktree.
// This does not work in the main worktree.
func (r *Repo) LocalSkipFile(file string) error {
	if r.CurrentWorktree == nil {
		return fmt.Errorf("not in a worktree")
	}

	if r.IsMainWorktree(r.CurrentWorktree) {
		return fmt.Errorf("local skip does not work in main worktree")
	}

	return r.skipFileInWorktree(r.CurrentWorktree, file)
}

// LocalUnskipFile unmarks a file so its changes will be tracked again only in the current worktree.
// This does not work in the main worktree.
func (r *Repo) LocalUnskipFile(file string) error {
	if r.CurrentWorktree == nil {
		return fmt.Errorf("not in a worktree")
	}

	if r.IsMainWorktree(r.CurrentWorktree) {
		return fmt.Errorf("local unskip does not work in main worktree")
	}

	return r.unskipFileInWorktree(r.CurrentWorktree, file)
}

// skipFileInWorktree skips a file in a specific worktree
func (r *Repo) skipFileInWorktree(wt *Worktree, file string) error {
	wtFilePath := filepath.Join(wt.Path, file)

	// Run git update-index --skip-worktree in this worktree
	_, err := r.RunGitCommand(wt, "update-index", "--skip-worktree", file)
	if err != nil {
		return fmt.Errorf("failed to skip-worktree: %w", err)
	}

	// If this is not the main worktree, replace file with symlink to main
	if !r.IsMainWorktree(wt) {
		mainFilePath := filepath.Join(r.MainWorktree.Path, file)

		// Remove existing file
		if err := os.Remove(wtFilePath); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}

		// Create symlink to main worktree's file
		if err := os.Symlink(mainFilePath, wtFilePath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
	}

	return nil
}

// unskipFileInWorktree unskips a file in a specific worktree
func (r *Repo) unskipFileInWorktree(wt *Worktree, file string) error {
	wtFilePath := filepath.Join(wt.Path, file)

	// Run git update-index --no-skip-worktree
	_, err := r.RunGitCommand(wt, "update-index", "--no-skip-worktree", file)
	if err != nil {
		return fmt.Errorf("failed to unskip: %w", err)
	}

	// If this is not the main worktree, remove symlink and copy from main
	if !r.IsMainWorktree(wt) {
		// Check if file is a symlink and remove it
		info, err := os.Lstat(wtFilePath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to check file: %w", err)
		}

		// Remove symlink if it exists
		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(wtFilePath); err != nil {
				return fmt.Errorf("failed to remove symlink: %w", err)
			}

			// Copy the file from main worktree to this worktree
			mainFilePath := filepath.Join(r.MainWorktree.Path, file)
			err = CopyPath(mainFilePath, wtFilePath)
			if err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}
		}
	}

	return nil
}

// ListSkippedFiles returns a list of files marked with skip-worktree in the main worktree
func (r *Repo) ListSkippedFiles() ([]string, error) {
	skippedMap, err := r.getSkippedFilesInWorktree(r.MainWorktree)
	if err != nil {
		return nil, err
	}

	// Convert map to slice
	files := make([]string, 0, len(skippedMap))
	for file := range skippedMap {
		files = append(files, file)
	}

	return files, nil
}

// PrintSkippedFiles prints a list of skipped files with markers for local differences
func (r *Repo) PrintSkippedFiles() error {
	// Get skipped files from main worktree
	mainSkipped, err := r.getSkippedFilesInWorktree(r.MainWorktree)
	if err != nil {
		return err
	}

	// Get current worktree
	if r.CurrentWorktree == nil {
		return fmt.Errorf("not in a worktree")
	}

	// Get skipped files from current worktree (if not main)
	var currentSkipped map[string]bool
	if !r.IsMainWorktree(r.CurrentWorktree) {
		currentSkipped, err = r.getSkippedFilesInWorktree(r.CurrentWorktree)
		if err != nil {
			return err
		}
	} else {
		// In main worktree, current = main
		currentSkipped = mainSkipped
	}

	if len(mainSkipped) == 0 && len(currentSkipped) == 0 {
		fmt.Println("No files with skipped changes.")
		return nil
	}

	// Collect all unique files
	allFiles := make(map[string]bool)
	for file := range mainSkipped {
		allFiles[file] = true
	}
	for file := range currentSkipped {
		allFiles[file] = true
	}

	// Print files with markers
	for file := range allFiles {
		inMain := mainSkipped[file]
		inCurrent := currentSkipped[file]

		if r.IsMainWorktree(r.CurrentWorktree) {
			// In main worktree, just print the file
			fmt.Println(file)
		} else {
			// In other worktrees, show status
			if inMain && inCurrent {
				// File is skipped globally
				fmt.Println(file)
			} else if inMain && !inCurrent {
				// File is globally skipped but locally unskipped
				fmt.Printf("%s (locally unskipped)\n", file)
			} else if !inMain && inCurrent {
				// File is only locally skipped (shouldn't happen with normal usage)
				fmt.Printf("%s (locally skipped only)\n", file)
			}
		}
	}

	return nil
}

// getSkippedFilesInWorktree returns a map of skipped files in a specific worktree
func (r *Repo) getSkippedFilesInWorktree(wt *Worktree) (map[string]bool, error) {
	output, err := r.RunGitCommand(wt, "ls-files", "-v")
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	skipped := make(map[string]bool)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		// Files with skip-worktree flag show as 'S' in the first column
		if strings.HasPrefix(line, "S ") {
			file := strings.TrimPrefix(line, "S ")
			skipped[file] = true
		}
	}

	return skipped, nil
}

// applySkipSettingsToWorktree applies all skip-worktree settings from the main worktree to a new worktree
func (r *Repo) applySkipSettingsToWorktree(wt *Worktree) error {
	// Don't apply to main worktree
	if r.IsMainWorktree(wt) {
		return nil
	}

	// Get list of skipped files from main worktree
	skippedFiles, err := r.getSkippedFilesInWorktree(r.MainWorktree)
	if err != nil {
		return fmt.Errorf("failed to get skipped files: %w", err)
	}

	// If no files are skipped, nothing to do
	if len(skippedFiles) == 0 {
		return nil
	}

	var errors []string

	// Apply skip and symlink for each file
	for file := range skippedFiles {
		wtFilePath := filepath.Join(wt.Path, file)

		// Check if file exists in this worktree
		if _, err := os.Lstat(wtFilePath); err != nil {
			if os.IsNotExist(err) {
				// File doesn't exist in this worktree, skip it
				continue
			}
			errors = append(errors, fmt.Sprintf("file %s: failed to check: %v", file, err))
			continue
		}

		// Apply skip-worktree and symlink
		if err := r.skipFileInWorktree(wt, file); err != nil {
			errors = append(errors, fmt.Sprintf("file %s: %v", file, err))
			continue
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some files failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}
