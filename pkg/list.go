package pkg

import (
	"fmt"
	"sort"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorCyan   = "\033[36m"
	ColorYellow = "\033[33m"
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
	if r.CurrentWorktree != nil && wt.Path == r.CurrentWorktree.Path {
		return MarkerCurrent
	}
	return MarkerNone
}

// GetWorktreeDisplay returns the formatted display string for a worktree
func (r *Repo) GetWorktreeDisplay(wt *Worktree, useColor bool) string {
	marker := r.GetWorktreeMarker(wt)

	// Build the display name
	display := wt.Name
	if wt.Branch != wt.Name {
		display = fmt.Sprintf("%s [%s]", wt.Name, wt.Branch)
	}

	// Add marker and color
	var prefix string
	switch marker {
	case MarkerMain:
		prefix = "> "
		if useColor {
			display = ColorCyan + display + ColorReset
		}
	case MarkerCurrent:
		prefix = "* "
		if useColor {
			display = ColorGreen + display + ColorReset
		}
	default:
		prefix = "  "
	}

	return prefix + display
}

// SortedWorktrees returns worktrees sorted with main first, then current, then alphabetically
func (r *Repo) SortedWorktrees() []Worktree {
	sorted := make([]Worktree, len(r.Worktrees))
	copy(sorted, r.Worktrees)

	sort.Slice(sorted, func(i, j int) bool {
		// Main worktree always comes first
		if r.IsMainWorktree(&sorted[i]) {
			return true
		}
		if r.IsMainWorktree(&sorted[j]) {
			return false
		}

		// Current worktree comes second
		if r.CurrentWorktree != nil {
			if sorted[i].Path == r.CurrentWorktree.Path {
				return true
			}
			if sorted[j].Path == r.CurrentWorktree.Path {
				return false
			}
		}

		// Otherwise sort alphabetically by name
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}
