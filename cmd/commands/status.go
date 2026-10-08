package commands

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/bungogood/wrk/pkg"
	"github.com/spf13/cobra"
)

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

// padVisible pads s with spaces to width, counting only visible characters
// so rows stay aligned when the display contains color codes.
func padVisible(s string, width int) string {
	if pad := width - len(ansiRe.ReplaceAllString(s, "")); pad > 0 {
		return s + strings.Repeat(" ", pad)
	}
	return s
}

var statusRemote string

var statusCmd = &cobra.Command{
	Use:               "status",
	Short:             "Show worktree status",
	Long:              `Shows every worktree with uncommitted changes, commits ahead of/behind the default branch, and last-commit age. Read-only: never fetches.`,
	Args:              cobra.NoArgs,
	ValidArgsFunction: cobra.NoFileCompletions,
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		if len(repo.Worktrees) == 0 {
			fmt.Println("No worktrees found.")
			return nil
		}

		// Compare against the default branch; fall back to per-row
		// unknowns when it cannot be determined.
		if cmd.Flags().Changed("remote") {
			if err := repo.RequireRemote(statusRemote); err != nil {
				return err
			}
		}
		base, baseErr := repo.DefaultBranchFor(repo.ResolveRemote(statusRemote))
		if baseErr != nil {
			base = ""
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
		probes := repo.ProbeWorktrees(present, base)
		byPath := make(map[string]pkg.WorktreeProbe, len(probes))
		for _, p := range probes {
			byPath[p.Worktree.Path] = p
		}

		for _, wt := range sorted {
			wt := wt
			if _, err := os.Stat(wt.Path); os.IsNotExist(err) {
				fmt.Printf("%s %-6s %7s %s\n", padVisible(repo.GetWorktreeDisplay(&wt), 28), "gone", "?", "?")
				continue
			}

			p := byPath[wt.Path]
			state, progress, age := "?", "?", "?"
			if p.DirtyErr == nil {
				state = "clean"
				if p.Dirty {
					state = "dirty"
				}
			}
			if p.Counted {
				progress = fmt.Sprintf("+%d/-%d", p.Ahead, p.Behind)
			}
			if p.HasDate {
				age = fmt.Sprintf("%s (%s)", pkg.HumanizeAge(now.Sub(p.LastCommit)), p.LastCommit.Format("2006-01-02"))
			}

			fmt.Printf("%s %-6s %7s %s\n", padVisible(repo.GetWorktreeDisplay(&wt), 28), state, progress, age)
		}
		return nil
	}),
}

// NewStatusCmd returns the status command
func NewStatusCmd() *cobra.Command {
	statusCmd.Flags().StringVarP(&statusRemote, "remote", "R", "", "Remote to resolve the default branch from (default: origin when configured, else first remote)")
	return statusCmd
}
