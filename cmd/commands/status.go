package commands

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/bungogood/worktree/pkg"
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
		now := time.Now()

		for _, wt := range repo.SortedWorktrees() {
			wt := wt
			state, progress, age := "?", "?", "?"

			if _, err := os.Stat(wt.Path); os.IsNotExist(err) {
				fmt.Printf("%-28s %-6s %7s %s\n", repo.GetWorktreeDisplay(&wt), "gone", "?", "?")
				continue
			}

			if dirty, err := repo.IsWorktreeDirty(&wt); err == nil {
				state = "clean"
				if dirty {
					state = "dirty"
				}
			}

			if baseErr == nil {
				if ahead, behind, err := repo.AheadBehind(&wt, base); err == nil {
					progress = fmt.Sprintf("+%d/-%d", ahead, behind)
				}
			}

			if last, err := repo.WorktreeLastActivity(&wt); err == nil {
				age = fmt.Sprintf("%s (%s)", pkg.HumanizeAge(now.Sub(last)), last.Format("2006-01-02"))
			}

			fmt.Printf("%-28s %-6s %7s %s\n", repo.GetWorktreeDisplay(&wt), state, progress, age)
		}
		return nil
	}),
}

// NewStatusCmd returns the status command
func NewStatusCmd() *cobra.Command {
	statusCmd.Flags().StringVarP(&statusRemote, "remote", "R", "", "Remote to resolve the default branch from (default: origin when configured, else first remote)")
	return statusCmd
}
