package pkg

import (
	"fmt"
	"path/filepath"
)

// CopyFromWorktree copies a file or directory from one worktree to another
func (r *Repo) CopyFromWorktree(sourceWt *Worktree, destWt *Worktree, srcPath string, dstPath string) error {
	// Check if trying to copy from self to self
	if sourceWt == destWt {
		return fmt.Errorf("cannot copy from worktree to itself")
	}

	// Build full paths
	fullSrcPath := filepath.Join(sourceWt.Path, srcPath)
	fullDstPath := filepath.Join(destWt.Path, dstPath)

	// Perform the copy
	if err := CopyPath(fullSrcPath, fullDstPath); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	return nil
}

// GetSourceWorktree determines the source worktree based on a name/branch or defaults to main
func (r *Repo) GetSourceWorktree(name string) (*Worktree, error) {
	if name != "" {
		return r.FindWorktree(name)
	}
	// Default to main worktree
	return r.MainWorktree, nil
}
