package pkg

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsWorktreeDirty_States(t *testing.T) {
	repoDir := t.TempDir()

	runGit(t, repoDir, "init", "-b", "main")
	runGit(t, repoDir, "config", "user.email", "tests@example.com")
	runGit(t, repoDir, "config", "user.name", "Tests")

	if err := os.WriteFile(filepath.Join(repoDir, "file.txt"), []byte("test\n"), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}
	runGit(t, repoDir, "add", "file.txt")
	runGit(t, repoDir, "commit", "-m", "init")

	r := &Repo{}
	wt := &Worktree{Path: repoDir, Name: "repo", Branch: "main"}

	check := func(want bool, what string) {
		t.Helper()
		got, err := r.IsWorktreeDirty(wt)
		if err != nil {
			t.Fatalf("IsWorktreeDirty (%s) failed: %v", what, err)
		}
		if got != want {
			t.Fatalf("IsWorktreeDirty (%s) = %v, want %v", what, got, want)
		}
	}

	check(false, "clean")
	runGit(t, repoDir, "checkout", "-q", "-b", "scratch")

	if err := os.WriteFile(filepath.Join(repoDir, "file.txt"), []byte("changed\n"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}
	check(true, "modified tracked file")

	runGit(t, repoDir, "checkout", "-q", "--", "file.txt")
	if err := os.WriteFile(filepath.Join(repoDir, "staged.txt"), []byte("staged\n"), 0644); err != nil {
		t.Fatalf("failed to write staged file: %v", err)
	}
	runGit(t, repoDir, "add", "staged.txt")
	check(true, "staged-only change")

	runGit(t, repoDir, "reset", "-q", "HEAD", "staged.txt")
	if err := os.Remove(filepath.Join(repoDir, "staged.txt")); err != nil {
		t.Fatalf("failed to remove staged file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "untracked.txt"), []byte("new\n"), 0644); err != nil {
		t.Fatalf("failed to write untracked file: %v", err)
	}
	check(true, "untracked-only change")
}

func TestAheadBehind_CountsAgainstBase(t *testing.T) {
	repoDir := t.TempDir()

	runGit(t, repoDir, "init", "-b", "main")
	runGit(t, repoDir, "config", "user.email", "tests@example.com")
	runGit(t, repoDir, "config", "user.name", "Tests")

	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("test\n"), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}
	runGit(t, repoDir, "add", "README.md")
	runGit(t, repoDir, "commit", "-m", "init")

	runGit(t, repoDir, "checkout", "-b", "feature")
	for i := range 2 {
		if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte(strings.Repeat("x", i+1)+"\n"), 0644); err != nil {
			t.Fatalf("failed to write seed file: %v", err)
		}
		runGit(t, repoDir, "add", "README.md")
		runGit(t, repoDir, "commit", "-m", "feature change")
	}
	runGit(t, repoDir, "checkout", "main")
	if err := os.WriteFile(filepath.Join(repoDir, "other.txt"), []byte("other\n"), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}
	runGit(t, repoDir, "add", "other.txt")
	runGit(t, repoDir, "commit", "-m", "main change")

	r := &Repo{}
	wt := &Worktree{Path: repoDir, Name: "repo", Branch: "feature"}
	runGit(t, repoDir, "checkout", "feature")
	ahead, behind, err := r.AheadBehind(wt, "main")
	if err != nil {
		t.Fatalf("AheadBehind failed: %v", err)
	}
	if ahead != 2 || behind != 1 {
		t.Fatalf("AheadBehind = +%d/-%d, want +2/-1", ahead, behind)
	}
}

func TestAllBranches_SkipsRemoteHead(t *testing.T) {
	repoDir := t.TempDir()
	remoteDir := filepath.Join(t.TempDir(), "remote.git")

	runGit(t, repoDir, "init", "-b", "main")
	runGit(t, repoDir, "config", "user.email", "tests@example.com")
	runGit(t, repoDir, "config", "user.name", "Tests")

	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("test\n"), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}
	runGit(t, repoDir, "add", "README.md")
	runGit(t, repoDir, "commit", "-m", "init")

	runGit(t, repoDir, "init", "--bare", remoteDir)
	runGit(t, repoDir, "remote", "add", "origin", remoteDir)
	runGit(t, repoDir, "push", "-u", "origin", "main")
	// Simulate a clone's origin/HEAD pointer
	runGit(t, repoDir, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")

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
	for _, b := range branches {
		// The origin/HEAD symref shortens to a bare "origin" — neither
		// form is a branch and neither may leak into completions.
		if b == "HEAD" || b == "origin" {
			t.Fatalf("expected no HEAD pseudo-branch, got %v", branches)
		}
	}
}

func TestResolveRemote_PrefersFlagThenDefault(t *testing.T) {
	repoDir := t.TempDir()
	remoteDir := filepath.Join(t.TempDir(), "remote.git")

	runGit(t, repoDir, "init", "-b", "main")
	runGit(t, repoDir, "init", "--bare", remoteDir)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to chdir to repo: %v", err)
	}

	r := &Repo{}

	// No remotes: auto resolves to "", explicit flag wins, unknown is rejected
	if got := r.ResolveRemote(""); got != "" {
		t.Fatalf("ResolveRemote(\"\") = %q, want \"\"", got)
	}
	if got := r.ResolveRemote("upstream"); got != "upstream" {
		t.Fatalf("ResolveRemote(\"upstream\") = %q, want \"upstream\"", got)
	}
	if err := r.RequireRemote("origin"); err == nil {
		t.Fatalf("RequireRemote(\"origin\") succeeded with no remotes, want error")
	}

	runGit(t, repoDir, "remote", "add", "upstream", remoteDir)

	// Single non-origin remote becomes the default
	if got := r.ResolveRemote(""); got != "upstream" {
		t.Fatalf("ResolveRemote(\"\") = %q, want \"upstream\"", got)
	}
	if err := r.RequireRemote("upstream"); err != nil {
		t.Fatalf("RequireRemote(\"upstream\") failed: %v", err)
	}
	if err := r.RequireRemote("nope"); err == nil {
		t.Fatalf("RequireRemote(\"nope\") succeeded, want error")
	}
}

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
