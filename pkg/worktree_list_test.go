package pkg

import (
	"strings"
	"testing"
)

func TestFindWorktree(t *testing.T) {
	r := &Repo{
		Worktrees: []Worktree{
			{Name: "main", Branch: "main", Path: "/repo"},
			{Name: "feature-auth", Branch: "feature/auth", Path: "/wt/feature-auth"},
			{Name: "feature-api", Branch: "feature/api", Path: "/wt/feature-api"},
		},
	}

	wt, err := r.FindWorktree("feature-auth")
	if err != nil {
		t.Fatalf("expected exact name match, got error: %v", err)
	}
	if wt.Name != "feature-auth" {
		t.Fatalf("expected feature-auth, got %q", wt.Name)
	}

	wt, err = r.FindWorktree("feature/auth")
	if err != nil {
		t.Fatalf("expected exact branch match, got error: %v", err)
	}
	if wt.Branch != "feature/auth" {
		t.Fatalf("expected feature/auth, got %q", wt.Branch)
	}

	_, err = r.FindWorktree("does-not-exist")
	if err == nil || !strings.Contains(err.Error(), "no worktree found") {
		t.Fatalf("expected no match error, got %v", err)
	}

	_, err = r.FindWorktree("feature/*")
	if err == nil {
		t.Fatalf("expected ambiguous match error")
	}
	if !strings.Contains(err.Error(), "matches multiple worktrees") {
		t.Fatalf("expected ambiguous error, got %v", err)
	}
	if !strings.Contains(err.Error(), "feature-auth") && !strings.Contains(err.Error(), "feature/api") {
		t.Fatalf("expected alias names in error, got %v", err)
	}
}

func TestSortedWorktrees_MainThenCurrentThenAlphabetical(t *testing.T) {
	r := &Repo{
		Worktrees: []Worktree{
			{Name: "zzz", Branch: "zzz", Path: "/wt/zzz"},
			{Name: "main", Branch: "main", Path: "/repo"},
			{Name: "bbb", Branch: "bbb", Path: "/wt/bbb"},
			{Name: "aaa", Branch: "aaa", Path: "/wt/aaa"},
		},
	}
	r.MainWorktree = &r.Worktrees[1]
	r.CurrentWorktree = &r.Worktrees[2]

	sorted := r.SortedWorktrees()

	got := []string{sorted[0].Name, sorted[1].Name, sorted[2].Name, sorted[3].Name}
	want := []string{"main", "bbb", "aaa", "zzz"}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected sort order: got %v want %v", got, want)
		}
	}
}
