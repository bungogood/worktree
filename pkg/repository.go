package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Repo represents a git repository with its worktrees
type Repo struct {
	Name            string     // Repository name
	WorktreesDir    string     // Path to .{repo}.worktrees directory
	Worktrees       []Worktree // All worktrees in the repo
	MainWorktree    *Worktree  // The main worktree (contains .git directory)
	CurrentWorktree *Worktree  // The worktree we're currently in
	Config          *Config    // Configuration settings
}

// LoadRepo discovers the git repository and all its worktrees
func LoadRepo() (*Repo, error) {
	// Find the main git directory (the one with .git directory, not file)
	mainGitDir, err := findMainGitDir()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}

	// Get repository name from the directory name
	repoName := filepath.Base(mainGitDir)

	// Construct worktrees directory path - in the same parent directory as the repo
	parentDir := filepath.Dir(mainGitDir)
	worktreesDir := filepath.Join(parentDir, fmt.Sprintf(".%s.worktrees", repoName))

	repo := &Repo{
		Name:         repoName,
		WorktreesDir: worktreesDir,
		Worktrees:    make([]Worktree, 0),
	}

	// Load all worktrees
	if err := repo.loadWorktrees(); err != nil {
		return nil, err
	}

	// Find and set the main worktree (the one with .git directory)
	for i := range repo.Worktrees {
		gitPath := filepath.Join(repo.Worktrees[i].Path, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			repo.MainWorktree = &repo.Worktrees[i]
			break
		}
	}

	// Determine current worktree
	cwd, _ := os.Getwd()
	for i := range repo.Worktrees {
		if strings.HasPrefix(cwd, repo.Worktrees[i].Path) {
			repo.CurrentWorktree = &repo.Worktrees[i]
			break
		}
	}

	if repo.CurrentWorktree == nil {
		return nil, fmt.Errorf("current directory is not inside any worktree")
	}

	// Load configuration
	config, err := repo.LoadConfig()
	if err != nil {
		return nil, err
	}
	repo.Config = config

	return repo, nil
}

// findMainGitDir finds the main git directory (the one with .git as a directory, not a file)
func findMainGitDir() (string, error) {
	// Use git to find the common git directory (the main .git directory)
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository: %w", err)
	}

	gitCommonDir := strings.TrimSpace(string(output))

	// The common dir is the .git directory, so get its parent
	mainDir := filepath.Dir(gitCommonDir)

	// Make it absolute if it's relative
	if !filepath.IsAbs(mainDir) {
		cwd, _ := os.Getwd()
		mainDir = filepath.Join(cwd, mainDir)
	}

	return mainDir, nil
}

// loadWorktrees loads all worktrees from git
func (r *Repo) loadWorktrees() error {
	output, err := r.RunGitCommand(nil, "worktree", "list", "--porcelain")

	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var wt *Worktree

	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			if wt != nil {
				r.Worktrees = append(r.Worktrees, *wt)
			}
			path := strings.TrimPrefix(line, "worktree ")
			name := filepath.Base(path)
			wt = &Worktree{
				Path: path,
				Name: name,
			}
		} else if strings.HasPrefix(line, "branch ") {
			if wt != nil {
				branch := strings.TrimPrefix(line, "branch refs/heads/")
				wt.Branch = branch
			}
		} else if line == "" && wt != nil {
			r.Worktrees = append(r.Worktrees, *wt)
			wt = nil
		}
	}

	return nil
}

func (r *Repo) AllBranches(remote string) ([]string, error) {
	// List local branches and branches from the selected remote.
	// An empty remote lists local branches only.
	refs := []string{"refs/heads"}
	if remote != "" {
		// Trailing slash keeps prefix remotes (origin vs origin2) apart.
		refs = append(refs, fmt.Sprintf("refs/remotes/%s/", remote))
	}
	output, err := r.RunGitCommand(nil, append([]string{"for-each-ref", "--format=%(refname)"}, refs...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	// Parse full refnames (short names are ambiguous: the "<remote>/HEAD"
	// symref shortens to a bare "<remote>", which is not a branch).
	lines := strings.Split(string(output), "\n")
	branches := make([]string, 0, len(lines))
	seen := make(map[string]bool, len(lines))
	remoteHEAD := "refs/remotes/" + remote + "/HEAD"
	remotePrefix := "refs/remotes/" + remote + "/"

	for _, line := range lines {
		full := strings.TrimSpace(line)
		if full == "" {
			continue
		}
		var branch string
		if after, ok := strings.CutPrefix(full, "refs/heads/"); ok {
			branch = after
		} else if remote == "" || full == remoteHEAD {
			continue
		} else if after, ok := strings.CutPrefix(full, remotePrefix); ok {
			branch = after
		} else {
			continue
		}
		if seen[branch] {
			continue
		}
		seen[branch] = true
		branches = append(branches, branch)
	}

	return branches, nil
}

// Remotes lists the configured git remotes.
func (r *Repo) Remotes() ([]string, error) {
	output, err := r.RunGitCommand(r.MainWorktree, "remote")
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	var remotes []string
	for _, line := range strings.Split(string(output), "\n") {
		if remote := strings.TrimSpace(line); remote != "" {
			remotes = append(remotes, remote)
		}
	}
	sort.Strings(remotes)

	return remotes, nil
}

// DefaultRemote returns the preferred remote: "origin" when configured,
// otherwise the first remote alphabetically, or "" when there are none.
func (r *Repo) DefaultRemote() string {
	remotes, err := r.Remotes()
	if err != nil || len(remotes) == 0 {
		return ""
	}
	if slices.Contains(remotes, "origin") {
		return "origin"
	}
	return remotes[0]
}

// ResolveRemote picks the remote to use: an explicit --remote flag value wins,
// otherwise the default remote. Returns "" when no remotes are configured.
func (r *Repo) ResolveRemote(explicit string) string {
	if explicit != "" {
		return explicit
	}
	return r.DefaultRemote()
}

// RequireRemote errors when the named remote is not configured.
func (r *Repo) RequireRemote(remote string) error {
	remotes, err := r.Remotes()
	if err != nil {
		return err
	}
	if !slices.Contains(remotes, remote) {
		if len(remotes) == 0 {
			return fmt.Errorf("unknown remote '%s': no remotes are configured", remote)
		}
		return fmt.Errorf("unknown remote '%s' (available: %s)", remote, strings.Join(remotes, ", "))
	}
	return nil
}

// DefaultBranch resolves the default remote's default branch offline (no fetch/pull).
// It reads the local refs/remotes/<remote>/HEAD symbolic ref and returns the
// short branch name (e.g. "main", not "origin/main").
func (r *Repo) DefaultBranch() (string, error) {
	return r.DefaultBranchFor(r.DefaultRemote())
}

// DefaultBranchFor resolves the given remote's default branch offline. An
// empty remote skips straight to the local main/master fallback.
func (r *Repo) DefaultBranchFor(remote string) (string, error) {
	if remote != "" {
		headRef := fmt.Sprintf("refs/remotes/%s/HEAD", remote)
		output, err := r.RunGitCommand(r.MainWorktree, "symbolic-ref", headRef)
		if err == nil {
			ref := strings.TrimSpace(string(output))
			if branch, ok := strings.CutPrefix(ref, fmt.Sprintf("refs/remotes/%s/", remote)); ok && branch != "" {
				return branch, nil
			}
		}
	}

	// Fallback when the remote HEAD is not set locally: prefer main, then master.
	// Qualified to local branches so a stale remote-only ref cannot win.
	if r.BranchExists("refs/heads/main") {
		return "main", nil
	}
	if r.BranchExists("refs/heads/master") {
		return "master", nil
	}

	return "", fmt.Errorf("could not determine default branch: no remote HEAD is set and neither 'main' nor 'master' exists locally")
}

// CurrentBranch returns the current branch name, or "" when detached.
func (r *Repo) CurrentBranch(wt *Worktree) (string, error) {
	output, err := r.RunGitCommand(wt, "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// SwitchBranch switches the worktree to the target branch.
func (r *Repo) SwitchBranch(wt *Worktree, target string) error {
	output, err := r.RunGitCommand(wt, "switch", target)
	if err != nil {
		return fmt.Errorf("failed to switch to branch '%s': %s: %w", target, strings.TrimSpace(string(output)), err)
	}
	return nil
}

// BranchLastCommit returns the time of the last commit on a local branch.
func (r *Repo) BranchLastCommit(branch string) (time.Time, error) {
	output, err := r.runGitRead(nil, "for-each-ref", "--format=%(committerdate:unix)", "refs/heads/"+branch)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last commit for branch '%s': %w", branch, err)
	}
	unix, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse last commit for branch '%s': %w", branch, err)
	}
	return time.Unix(unix, 0), nil
}

// BranchLastCommitMap returns the last-commit time of every local branch in
// a single git invocation, so status-style sweeps don't fork per worktree.
func (r *Repo) BranchLastCommitMap() (map[string]time.Time, error) {
	output, err := r.runGitRead(r.MainWorktree, "for-each-ref", "--format=%(refname) %(committerdate:unix)", "refs/heads/")
	if err != nil {
		return nil, fmt.Errorf("failed to list branch dates: %w", err)
	}

	dates := make(map[string]time.Time)
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Fields(strings.TrimSpace(line))
		if len(parts) != 2 {
			continue
		}
		branch, ok := strings.CutPrefix(parts[0], "refs/heads/")
		if !ok || branch == "" {
			continue
		}
		if unix, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
			dates[branch] = time.Unix(unix, 0)
		}
	}
	return dates, nil
}

// WorktreeLastActivity returns the last commit reachable from a worktree:
// the tip of its branch, or HEAD when detached.
func (r *Repo) WorktreeLastActivity(wt *Worktree) (time.Time, error) {
	if wt.Branch != "" {
		return r.BranchLastCommit(wt.Branch)
	}
	output, err := r.runGitRead(wt, "log", "-1", "--format=%ct", "HEAD")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last commit for worktree '%s': %w", wt.Name, err)
	}
	unix, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse last commit for worktree '%s': %w", wt.Name, err)
	}
	return time.Unix(unix, 0), nil
}

// IsWorktreeDirty reports whether a worktree has uncommitted changes.
func (r *Repo) IsWorktreeDirty(wt *Worktree) (bool, error) {
	output, err := r.runGitRead(wt, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("failed to check status for worktree '%s': %w", wt.Name, err)
	}
	return strings.TrimSpace(string(output)) != "", nil
}

// readOnlyEnv marks git invocations that must never take locks or refresh
// the index (status-style reads across many worktrees).
var readOnlyEnv = []string{"GIT_OPTIONAL_LOCKS=0"}

// runGitRead runs a git command that only reads, never writes.
func (r *Repo) runGitRead(wt *Worktree, args ...string) ([]byte, error) {
	if wt != nil {
		args = append([]string{"-C", wt.Path}, args...)
	}
	return RunCommandEnv("git", readOnlyEnv, args...)
}

// PruneWorktrees clears git metadata for worktree directories that no
// longer exist on disk. Returns how many entries were pruned.
func (r *Repo) PruneWorktrees() (int, error) {
	missing := 0
	for i := range r.Worktrees {
		if _, err := os.Stat(r.Worktrees[i].Path); os.IsNotExist(err) {
			missing++
		}
	}
	if missing == 0 {
		return 0, nil
	}
	if _, err := r.RunGitCommand(r.MainWorktree, "worktree", "prune"); err != nil {
		return 0, fmt.Errorf("failed to prune worktrees: %w", err)
	}
	return missing, nil
}

// AheadBehind counts how many commits the worktree's HEAD is ahead of and
// behind the base branch. Uses local refs only (no fetch).
func (r *Repo) AheadBehind(wt *Worktree, base string) (ahead, behind int, err error) {
	output, err := r.runGitRead(wt, "rev-list", "--left-right", "--count", base+"...HEAD")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to compare with '%s': %w", base, err)
	}
	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output %q", strings.TrimSpace(string(output)))
	}
	if behind, err = strconv.Atoi(parts[0]); err != nil {
		return 0, 0, fmt.Errorf("unexpected rev-list output %q: %w", parts[0], err)
	}
	if ahead, err = strconv.Atoi(parts[1]); err != nil {
		return 0, 0, fmt.Errorf("unexpected rev-list output %q: %w", parts[1], err)
	}
	return ahead, behind, nil
}

// GetWorktreePath returns the path where a worktree for the given branch should be
func (r *Repo) GetWorktreePath(branch string) string {
	return filepath.Join(r.WorktreesDir, branch)
}

// EnsureWorktreesDir creates the .{repo}.worktrees directory if it doesn't exist
func (r *Repo) EnsureWorktreesDir() error {
	if _, err := os.Stat(r.WorktreesDir); os.IsNotExist(err) {
		if GlobalFlags.Verbose {
			fmt.Fprintf(os.Stderr, "Creating worktrees directory: %s\n", r.WorktreesDir)
		}
		return os.MkdirAll(r.WorktreesDir, 0755)
	}
	return nil
}

func (r *Repo) IsMainWorktree(wt *Worktree) bool {
	return r.MainWorktree != nil && wt.Path == r.MainWorktree.Path
}

// BranchExists checks if a branch exists
func (r *Repo) BranchExists(branch string) bool {
	_, err := r.RunGitCommand(nil, "rev-parse", "--verify", branch)
	return err == nil
}

func (r *Repo) RunGitCommand(wt *Worktree, args ...string) ([]byte, error) {
	if wt != nil {
		args = append([]string{"-C", wt.Path}, args...)
	}
	return RunCommand("git", args...)
}
