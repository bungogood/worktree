package pkg

import (
	"fmt"
)

// Worktree represents a git worktree
type Worktree struct {
	Path         string // Absolute path to the worktree
	Branch       string // Branch name
	Name         string // Worktree name
	RemoteBranch string // Remote name if created from remote branch, empty if local
}

// FindWorktreeByBranch finds a worktree by branch name
func (r *Repo) FindWorktreeByBranch(branch string) *Worktree {
	for i := range r.Worktrees {
		if r.Worktrees[i].Branch == branch {
			return &r.Worktrees[i]
		}
	}
	return nil
}

// FindWorktreeByName finds a worktree by name
func (r *Repo) FindWorktreeByName(name string) *Worktree {
	for i := range r.Worktrees {
		if r.Worktrees[i].Name == name {
			return &r.Worktrees[i]
		}
	}
	return nil
}

// FindWorktree finds a worktree by either name then branch
func (r *Repo) FindWorktree(tree string) *Worktree {
	wt := r.FindWorktreeByName(tree)
	if wt != nil {
		return wt
	}
	return r.FindWorktreeByBranch(tree)
}

// AddExistingBranch creates a worktree for an existing local or remote branch
func (r *Repo) AddExistingBranch(branch, name, remote string) (*Worktree, error) {
	// Check if worktree already exists
	if existing := r.FindWorktreeByName(name); existing != nil {
		return nil, fmt.Errorf("worktree already exists: %s", name)
	}

	if existing := r.FindWorktreeByBranch(branch); existing != nil {
		return existing, fmt.Errorf("worktree already exists for branch '%s'", branch)
	}

	// Ensure the worktrees directory exists
	if err := r.EnsureWorktreesDir(); err != nil {
		return nil, fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	// Get the path for the new worktree using the custom name
	worktreePath := r.GetWorktreePath(name)

	// Check if branch exists locally or on remote
	var err error
	remoteBranch := fmt.Sprintf("%s/%s", remote, branch)
	if r.BranchExists(branch) {
		// Branch exists locally
		_, err = r.RunGitCommand(nil, "worktree", "add", worktreePath, branch)
	} else if r.BranchExists(remoteBranch) {
		// Branch exists on remote, create worktree with tracking
		_, err = r.RunGitCommand(nil, "worktree", "add", "-b", branch, worktreePath, remoteBranch)
	} else {
		return nil, fmt.Errorf("branch '%s' does not exist locally or on remote '%s'", branch, remote)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w", err)
	}

	wt := &Worktree{
		Path:         worktreePath,
		Branch:       branch,
		Name:         name,
		RemoteBranch: remoteBranch,
	}

	r.applyPostCreateSetup(wt)

	return wt, nil
}

// CreateNewBranch creates a worktree with a new branch
func (r *Repo) CreateNewBranch(branch, name string) (*Worktree, error) {
	// Check if branch already exists
	if r.BranchExists(branch) {
		return nil, fmt.Errorf("branch '%s' already exists", branch)
	}

	// Check if worktree already exists
	if existing := r.FindWorktreeByName(name); existing != nil {
		return nil, fmt.Errorf("worktree already exists at: %s", existing.Path)
	}

	// Ensure the worktrees directory exists
	if err := r.EnsureWorktreesDir(); err != nil {
		return nil, fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	// Get the path for the new worktree using the custom name
	worktreePath := r.GetWorktreePath(name)

	// Create the new worktree with a new branch
	_, err := r.RunGitCommand(nil, "worktree", "add", "-b", branch, worktreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w", err)
	}

	wt := &Worktree{
		Path:   worktreePath,
		Branch: branch,
		Name:   name,
	}

	r.applyPostCreateSetup(wt)

	return wt, nil
}

// applyPostCreateSetup applies all post-create operations to a worktree
func (r *Repo) applyPostCreateSetup(wt *Worktree) {
	// Apply skip-worktree settings to the new worktree
	if err := r.applySkipSettingsToWorktree(wt); err != nil {
		// Log error but don't fail the worktree creation
		fmt.Printf("Warning: failed to apply skip settings: %v\n", err)
	}

	// Apply always-copy settings
	if err := r.ApplyAlwaysCopy(wt); err != nil {
		// Log error but don't fail the worktree creation
		fmt.Printf("Warning: failed to apply always-copy: %v\n", err)
	}

	// Run post-create commands
	if err := r.RunPostCreateCommands(wt); err != nil {
		// Log error but don't fail the worktree creation
		fmt.Printf("Warning: failed to run post-create commands: %v\n", err)
	}
}

// RemoveWorktree removes a worktree and optionally force deletes the branch
func (r *Repo) RemoveWorktree(wt *Worktree, forceDeleteBranch bool) error {
	// Protect the main worktree
	if r.IsMainWorktree(wt) {
		return fmt.Errorf("cannot remove the main worktree (contains .git directory)")
	}

	// Remove the worktree
	_, err := r.RunGitCommand(nil, "worktree", "remove", wt.Path, "--force")
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	// Determine if we should delete the branch
	shouldDeleteBranch := forceDeleteBranch || (r.Config != nil && r.Config.DeleteBranchWithWorktree)

	// Delete the branch if requested
	if shouldDeleteBranch {
		_, err := r.RunGitCommand(r.MainWorktree, "branch", "-D", wt.Branch)
		if err != nil {
			return fmt.Errorf("failed to force delete branch: %w", err)
		}
	}

	return nil
}
