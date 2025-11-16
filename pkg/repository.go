package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repo represents a git repository with its worktrees
type Repo struct {
	Name            string     // Repository name
	WorktreesDir    string     // Path to .{repo}.worktrees directory
	Worktrees       []Worktree // All worktrees in the repo
	MainWorktree    *Worktree  // The main worktree (contains .git directory)
	CurrentWorktree *Worktree  // The worktree we're currently in
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

func (r *Repo) AllBranches() ([]string, error) {
	// find all branches of the repo not just work tree one
	output, err := r.RunGitCommand(nil, "branch", "--all", "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	branches := make([]string, 0, len(lines))
	for _, line := range lines {
		branch := strings.TrimSpace(line)
		if branch != "" {
			branches = append(branches, branch)
		}
	}

	return branches, nil
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
