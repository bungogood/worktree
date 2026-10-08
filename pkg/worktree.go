package pkg

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Worktree represents a git worktree
type Worktree struct {
	Path         string // Absolute path to the worktree
	Branch       string // Branch name
	Name         string // Worktree name
	RemoteBranch string // Remote name if created from remote branch, empty if local
}

// WorktreeProbe holds one worktree's gathered read-only state.
type WorktreeProbe struct {
	Worktree   Worktree
	LastCommit time.Time
	HasDate    bool
	Dirty      bool
	DirtyErr   error
	Ahead      int
	Behind     int
	Counted    bool // ahead/behind resolved against base
}

// ProbeWorktrees gathers dirty state, last-commit date and ahead/behind vs
// base for the given worktrees concurrently (bounded). Branch dates come
// from a single batched lookup unless a precomputed map is passed;
// detached worktrees fall back to a per-worktree log. An empty base skips
// ahead/behind. Individual failures are recorded on the probe and never
// abort the sweep.
func (r *Repo) ProbeWorktrees(worktrees []Worktree, base string, dates map[string]time.Time) []WorktreeProbe {
	probes := make([]WorktreeProbe, len(worktrees))

	if dates == nil {
		if batched, err := r.BranchLastCommitMap(); err == nil {
			dates = batched
		}
	}

	const maxParallel = 8
	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup
	for i := range worktrees {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			wt := worktrees[i]
			p := WorktreeProbe{Worktree: wt}

			if dates != nil {
				if last, ok := dates[wt.Branch]; ok {
					p.LastCommit, p.HasDate = last, true
				}
			}
			if !p.HasDate {
				// Covers detached worktrees and a failed batch lookup.
				if last, err := r.WorktreeLastActivity(&wt); err == nil {
					p.LastCommit, p.HasDate = last, true
				}
			}

			if dirty, err := r.IsWorktreeDirty(&wt); err == nil {
				p.Dirty = dirty
			} else {
				p.DirtyErr = err
			}

			if base != "" {
				if ahead, behind, err := r.AheadBehind(&wt, base); err == nil {
					p.Ahead, p.Behind, p.Counted = ahead, behind, true
				}
			}

			probes[i] = p
		}(i)
	}
	wg.Wait()
	return probes
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

func (r *Repo) WorktreeAliases() []string {
	var aliases []string
	for _, wt := range r.Worktrees {
		aliases = append(aliases, wt.Name)
		// Skip empty (detached) and duplicate branch aliases so each
		// worktree completes exactly once per distinct name.
		if wt.Branch != "" && wt.Branch != wt.Name {
			aliases = append(aliases, wt.Branch)
		}
	}
	return aliases
}

// FindWorktreeByBranchGlob finds all worktrees matching a glob pattern against their branch
func (r *Repo) FindWorktreeGlob(pattern string) ([]*Worktree, []string) {
	wtMatchSet := make(map[*Worktree]bool)
	aliasMatchSet := make(map[string]bool)
	for i := range r.Worktrees {
		wt := &r.Worktrees[i]
		if matched, _ := filepath.Match(pattern, wt.Branch); matched {
			wtMatchSet[wt] = true
			aliasMatchSet[wt.Branch] = true
		}
		if matched, _ := filepath.Match(pattern, wt.Name); matched {
			wtMatchSet[wt] = true
			aliasMatchSet[wt.Name] = true
		}
	}

	var wtMatches []*Worktree
	var aliasMatches []string

	for wt := range wtMatchSet {
		wtMatches = append(wtMatches, wt)
	}
	for alias := range aliasMatchSet {
		aliasMatches = append(aliasMatches, alias)
	}

	return wtMatches, aliasMatches
}

// FindWorktree finds a unique worktree matching a pattern (exact or glob)
// Returns error if no matches or multiple matches found
func (r *Repo) FindWorktree(pattern string) (*Worktree, error) {
	matches, aliasMatches := r.FindWorktreeGlob(pattern)

	if len(matches) == 0 {
		return nil, fmt.Errorf("no worktree found matching '%s'", pattern)
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	return nil, fmt.Errorf("pattern '%s' matches multiple worktrees:\n  %s", pattern, strings.Join(aliasMatches, "\n  "))
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

	// Check if branch exists locally or on remote (empty remote means no
	// remotes are configured, so only local branches qualify).
	var err error
	remoteTracking := ""
	if r.BranchExists(branch) {
		// Branch exists locally
		_, err = r.RunGitCommand(nil, "worktree", "add", worktreePath, branch)
	} else if remote != "" && r.BranchExists(fmt.Sprintf("%s/%s", remote, branch)) {
		// Branch exists on remote, create worktree with tracking
		remoteBranch := fmt.Sprintf("%s/%s", remote, branch)
		_, err = r.RunGitCommand(nil, "worktree", "add", "-b", branch, worktreePath, remoteBranch)
		remoteTracking = remoteBranch
	} else if remote != "" {
		return nil, fmt.Errorf("branch '%s' does not exist locally or on remote '%s'", branch, remote)
	} else {
		return nil, fmt.Errorf("branch '%s' does not exist locally (no remotes are configured)", branch)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w", err)
	}

	wt := &Worktree{
		Path:         worktreePath,
		Branch:       branch,
		Name:         name,
		RemoteBranch: remoteTracking,
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
		color.Yellow("Warning: failed to apply skip settings: %v\n", err)
	}

	// Apply always-copy settings
	if err := r.ApplyAlwaysCopy(wt); err != nil {
		// Log error but don't fail the worktree creation
		color.Yellow("Warning: failed to apply always-copy: %v\n", err)
	}

	// Run post-create commands
	if err := r.RunPostCreateCommands(wt); err != nil {
		// Log error but don't fail the worktree creation
		color.Yellow("Warning: failed to run post-create commands: %v\n", err)
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
