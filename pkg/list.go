package pkg

import (
	"fmt"
	"sort"

	"github.com/fatih/color"
)

// WorktreeMarker represents the marker type for a worktree
type WorktreeMarker int

const (
	MarkerNone WorktreeMarker = iota
	MarkerMain
	MarkerCurrent
)

// GetWorktreeMarker returns the marker type for a worktree
func (r *Repo) GetWorktreeMarker(wt *Worktree) WorktreeMarker {
	if r.IsMainWorktree(wt) {
		return MarkerMain
	}
	if wt.Path == r.CurrentWorktree.Path {
		return MarkerCurrent
	}
	return MarkerNone
}

// MarkerGlyph returns the two-character marker prefix for a worktree.
func (r *Repo) MarkerGlyph(wt *Worktree) string {
	switch r.GetWorktreeMarker(wt) {
	case MarkerMain:
		return "> "
	case MarkerCurrent:
		return "* "
	default:
		return "  "
	}
}

// WorktreeLabel returns the bare "name [branch]" label without markers.
func WorktreeLabel(wt *Worktree) string {
	if wt.Branch != "" && wt.Branch != wt.Name {
		return fmt.Sprintf("%s [%s]", wt.Name, wt.Branch)
	}
	return wt.Name
}

// StatusRow formats a status/clean row in the canonical order:
// age, state, marker, label.
func StatusRow(age, state, marker, label string) string {
	return fmt.Sprintf("%-7s %-6s %s%s", age, state, marker, label)
}

// GetWorktreeDisplay returns the formatted display string for a worktree
func (r *Repo) GetWorktreeDisplay(wt *Worktree) string {
	marker := r.GetWorktreeMarker(wt)

	// Build the display name
	display := WorktreeLabel(wt)

	// Add marker and color
	var prefix string
	switch marker {
	case MarkerMain:
		prefix = "> "
		display = color.CyanString(display)
	case MarkerCurrent:
		prefix = "* "
		display = color.GreenString(display)
	default:
		prefix = "  "
	}

	return prefix + display
}

// SortedWorktrees returns worktrees sorted with main first, then current, then alphabetically
func (r *Repo) SortedWorktrees() []Worktree {
	sorted := make([]Worktree, len(r.Worktrees))
	copy(sorted, r.Worktrees)

	rank := func(wt *Worktree) int {
		if r.IsMainWorktree(wt) {
			return 0
		}
		if r.CurrentWorktree != nil && wt.Path == r.CurrentWorktree.Path {
			return 1
		}
		return 2
	}

	sort.Slice(sorted, func(i, j int) bool {
		ri := rank(&sorted[i])
		rj := rank(&sorted[j])
		if ri != rj {
			return ri < rj
		}

		// Otherwise sort alphabetically by name.
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}
