package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree
type Worktree struct {
	Path   string // Absolute path to the worktree
	Branch string // Branch name
	Name   string // Worktree name
}

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
	output, err := RunCommand("git", "worktree", "list", "--porcelain")

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

func (r *Repo) RunGitCommand(wt *Worktree, args ...string) ([]byte, error) {
	if wt != nil {
		args = append([]string{"-C", wt.Path}, args...)
	}
	return RunCommand("git", args...)
}

// AddExistingBranch creates a worktree for an existing local or remote branch
func (r *Repo) AddExistingBranch(branch, name string) (*Worktree, error) {
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

	// Check if branch exists locally
	_, err := r.RunGitCommand(nil, "rev-parse", "--verify", branch)
	localExists := err == nil

	// If not local, check if it exists on remote
	if !localExists {
		_, err = r.RunGitCommand(nil, "rev-parse", "--verify", fmt.Sprintf("origin/%s", branch))
		if err != nil {
			return nil, fmt.Errorf("branch '%s' does not exist locally or on remote", branch)
		}
		// Branch exists on remote, create worktree with tracking
		_, err = r.RunGitCommand(nil, "worktree", "add", "-b", branch, worktreePath, fmt.Sprintf("origin/%s", branch))
	} else {
		// Branch exists locally
		_, err = r.RunGitCommand(nil, "worktree", "add", worktreePath, branch)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w", err)
	}

	return &Worktree{
		Path:   worktreePath,
		Branch: branch,
		Name:   name,
	}, nil
}

// CreateNewBranch creates a worktree with a new branch
func (r *Repo) CreateNewBranch(branch, name string) (*Worktree, error) {
	// Check if branch already exists
	_, err := r.RunGitCommand(nil, "rev-parse", "--verify", branch)
	if err == nil {
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
	_, err = r.RunGitCommand(nil, "worktree", "add", "-b", branch, worktreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w", err)
	}

	return &Worktree{
		Path:   worktreePath,
		Branch: branch,
		Name:   name,
	}, nil
}
