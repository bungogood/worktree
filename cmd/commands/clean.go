package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var (
	cleanOlder  string
	cleanForce  bool
	cleanDryRun bool
)

type staleWorktree struct {
	worktree   *pkg.Worktree
	lastCommit time.Time
	dirty      bool
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove stale worktrees",
	Long:  `Removes worktrees whose branch has seen no commits for longer than the threshold. Never touches the main or current worktree, and skips worktrees with uncommitted changes unless --force is given. Branches are kept unless deleteBranchWithWorktree is set in the config.`,
	Args:  cobra.NoArgs,
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		maxAge, err := pkg.ParseMaxAge(cleanOlder)
		if err != nil {
			return err
		}
		cutoff := time.Now().Add(-maxAge)

		var stale []staleWorktree
		for i := range repo.Worktrees {
			wt := &repo.Worktrees[i]
			if repo.IsMainWorktree(wt) || wt.Path == repo.CurrentWorktree.Path {
				continue
			}

			last, err := repo.WorktreeLastActivity(wt)
			if err != nil {
				fmt.Printf("Skipped '%s': %v\n", wt.Name, err)
				continue
			}
			if last.After(cutoff) {
				continue
			}

			dirty, err := repo.IsWorktreeDirty(wt)
			if err != nil {
				return err
			}
			stale = append(stale, staleWorktree{worktree: wt, lastCommit: last, dirty: dirty})
		}

		if len(stale) == 0 {
			fmt.Printf("No worktrees older than %s\n", cleanOlder)
			return nil
		}

		// Oldest first
		sort.Slice(stale, func(i, j int) bool { return stale[i].lastCommit.Before(stale[j].lastCommit) })

		now := time.Now()
		fmt.Printf("Stale worktrees (older than %s):\n", cleanOlder)
		for _, s := range stale {
			marker := ""
			if s.dirty {
				marker = " (uncommitted changes)"
			}
			fmt.Printf("  %s  last commit %s (%s ago)%s\n",
				strings.TrimSpace(repo.GetWorktreeDisplay(s.worktree)),
				s.lastCommit.Format("2006-01-02"),
				pkg.HumanizeAge(now.Sub(s.lastCommit)),
				marker)
		}

		// Partition into removable and skipped (dirty without --force)
		var removable []*pkg.Worktree
		var skipped []string
		for _, s := range stale {
			if s.dirty && !cleanForce {
				skipped = append(skipped, s.worktree.Name)
				continue
			}
			removable = append(removable, s.worktree)
		}

		if cleanDryRun {
			names := make([]string, 0, len(removable))
			for _, wt := range removable {
				names = append(names, wt.Name)
			}
			fmt.Printf("Would remove %d worktree(s): %s\n", len(names), strings.Join(names, ", "))
			if len(skipped) > 0 {
				fmt.Printf("Would skip %d worktree(s) with uncommitted changes (use -f): %s\n", len(skipped), strings.Join(skipped, ", "))
			}
			return nil
		}

		var removed []string
		var errors []string
		for _, wt := range removable {
			// Remove the worktree directory (the branch goes too only
			// if deleteBranchWithWorktree is set in the config).
			if err := repo.RemoveWorktree(wt, false); err != nil {
				errors = append(errors, fmt.Sprintf("  %s: %v", wt.Name, err))
			} else {
				removed = append(removed, wt.Name)
			}
		}

		if len(removed) > 0 {
			fmt.Printf("Removed %d worktree(s): %s\n", len(removed), strings.Join(removed, ", "))
		}
		if len(skipped) > 0 {
			fmt.Printf("Skipped %d worktree(s) with uncommitted changes (use -f): %s\n", len(skipped), strings.Join(skipped, ", "))
		}
		if len(errors) > 0 {
			return fmt.Errorf("failed to remove %d worktree(s):\n%s", len(errors), strings.Join(errors, "\n"))
		}

		return nil
	}),
}

// NewCleanCmd returns the clean command
func NewCleanCmd() *cobra.Command {
	cleanCmd.Flags().StringVarP(&cleanOlder, "older", "o", "2w", "Remove worktrees with no commits for longer than this (e.g. 36h, 14d, 4w, or bare days)")
	cleanCmd.Flags().BoolVarP(&cleanForce, "force", "f", false, "Also remove worktrees with uncommitted changes")
	cleanCmd.Flags().BoolVarP(&cleanDryRun, "dry-run", "n", false, "List stale worktrees without removing them")
	return cleanCmd
}
