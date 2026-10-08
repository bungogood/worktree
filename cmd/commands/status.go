package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/bungogood/wrk/pkg"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:               "status",
	Short:             "Show worktree status",
	Long:              `Shows every worktree with uncommitted changes and last-commit age. Age is the last commit on the branch (HEAD when detached); the dirty flag guards uncommitted work. Read-only: never fetches.`,
	Args:              cobra.NoArgs,
	ValidArgsFunction: cobra.NoFileCompletions,
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		if len(repo.Worktrees) == 0 {
			fmt.Println("No worktrees found.")
			return nil
		}

		// One-time per repo: enable git's read caches so status-style
		// sweeps stay fast on huge trees. Never fails the command.
		if tuned, err := repo.EnsureFastReads(); err == nil && tuned {
			fmt.Fprintln(os.Stderr, "Enabled git read caches (core.fsmonitor, core.untrackedCache) for faster status; opt out: git config core.fsmonitor false")
		}

		now := time.Now()

		// Gather state for all present worktrees in one concurrent sweep;
		// missing directories are reported as gone without any git calls.
		sorted := repo.SortedWorktrees()
		var present []pkg.Worktree
		for i := range sorted {
			if _, err := os.Stat(sorted[i].Path); os.IsNotExist(err) {
				continue
			}
			present = append(present, sorted[i])
		}
		probes := repo.ProbeWorktrees(present, "", nil)
		byPath := make(map[string]pkg.WorktreeProbe, len(probes))
		for _, p := range probes {
			byPath[p.Worktree.Path] = p
		}

		for i := range sorted {
			wt := &sorted[i]
			if _, err := os.Stat(wt.Path); os.IsNotExist(err) {
				fmt.Println(pkg.StatusRow("?", "gone", repo.MarkerGlyph(wt), pkg.WorktreeLabel(wt)))
				continue
			}

			p := byPath[wt.Path]
			state, age := "?", "?"
			if p.DirtyErr == nil {
				state = "clean"
				if p.Dirty {
					state = "dirty"
				}
			}
			if p.HasDate {
				age = pkg.HumanizeAge(now.Sub(p.LastCommit))
			}

			fmt.Println(pkg.StatusRow(age, state, repo.MarkerGlyph(wt), pkg.WorktreeLabel(wt)))
		}
		return nil
	}),
}

// NewStatusCmd returns the status command
func NewStatusCmd() *cobra.Command {
	return statusCmd
}
