package pkg

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllBranches_IncludesLocalSlashAndRemoteBranches(t *testing.T) {
	repoDir := t.TempDir()
	remoteDir := filepath.Join(t.TempDir(), "remote.git")

	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "tests@example.com")
	runGit(t, repoDir, "config", "user.name", "Tests")

	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("test\n"), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}
	runGit(t, repoDir, "add", "README.md")
	runGit(t, repoDir, "commit", "-m", "init")
	runGit(t, repoDir, "branch", "-M", "main")

	runGit(t, repoDir, "checkout", "-b", "feature/local")
	runGit(t, repoDir, "checkout", "main")

	runGit(t, repoDir, "init", "--bare", remoteDir)
	runGit(t, repoDir, "remote", "add", "origin", remoteDir)
	runGit(t, repoDir, "push", "-u", "origin", "main")

	runGit(t, repoDir, "checkout", "-b", "feature/remote")
	runGit(t, repoDir, "push", "-u", "origin", "feature/remote")
	runGit(t, repoDir, "checkout", "main")
	runGit(t, repoDir, "branch", "-D", "feature/remote")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to chdir to repo: %v", err)
	}

	r := &Repo{}
	branches, err := r.AllBranches("origin")
	if err != nil {
		t.Fatalf("AllBranches failed: %v", err)
	}

	assertContains(t, branches, "main")
	assertContains(t, branches, "feature/local")
	assertContains(t, branches, "feature/remote")

	for _, b := range branches {
		if strings.HasPrefix(b, "origin/") {
			t.Fatalf("expected stripped remote name, got %q", b)
		}
	}

	seen := map[string]bool{}
	for _, b := range branches {
		if seen[b] {
			t.Fatalf("expected unique branch list, duplicate %q in %v", b, branches)
		}
		seen[b] = true
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

func assertContains(t *testing.T, values []string, expected string) {
	t.Helper()
	for _, v := range values {
		if v == expected {
			return
		}
	}
	t.Fatalf("expected %q in %v", expected, values)
}
