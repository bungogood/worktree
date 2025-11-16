package commands

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var (
	forceDelete bool
)

var removeCmd = &cobra.Command{
	Use:     "remove [branch...]",
	Aliases: []string{"rm"},
	Short:   "Remove worktrees",
	Long:    `Remove one or more worktrees. If no branches are specified, removes the current worktree. Cannot remove the main worktree.`,
	ValidArgsFunction: pkg.RepoValidArgsFunction(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {
		// Return list of branch names (excluding already specified ones)
		var branches []string
		for _, wt := range repo.Worktrees {
			if !repo.IsMainWorktree(&wt) && !slices.Contains(args, wt.Branch) {
				branches = append(branches, wt.Branch)
			}
		}
		return branches, cobra.ShellCompDirectiveNoFileComp
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		var worktreesToRemove []*pkg.Worktree
		var removed []string
		var errors []string

		// Determine which worktrees to remove
		if len(args) == 0 {
			// No trees specified, use current worktree
			if repo.CurrentWorktree == nil {
				return fmt.Errorf("not currently in a worktree")
			}
			worktreesToRemove = append(worktreesToRemove, repo.CurrentWorktree)
		} else {
			// Trees specified, find them all
			for _, tree := range args {
				wt := repo.FindWorktree(tree)
				if wt != nil {
					worktreesToRemove = append(worktreesToRemove, wt)
				} else {
					errors = append(errors, fmt.Sprintf("  no worktree found: '%s'", tree))
				}
			}
		}

		// Remove all worktrees, collecting errors
		removedCurrent := false

		for _, wt := range worktreesToRemove {
			if err := repo.RemoveWorktree(wt, forceDelete); err != nil {
				errors = append(errors, fmt.Sprintf("  %s: %v", wt.Name, err))
			} else {
				removed = append(removed, wt.Name)
				if wt == repo.CurrentWorktree {
					removedCurrent = true
				}
			}
		}

		// Print results
		if len(removed) > 0 {
			fmt.Printf("Removed %d worktree(s): %s\n", len(removed), strings.Join(removed, ", "))
		}

		if len(errors) > 0 {
			return fmt.Errorf("failed to remove %d worktree(s):\n%s", len(errors), strings.Join(errors, "\n"))
		}

		// If we removed the current worktree, cd to the main worktree
		if removedCurrent {
			pkg.ChangeDirectory(repo.MainWorktree.Path)
		}

		return nil
	}),
}

// NewRemoveCmd returns the remove command
func NewRemoveCmd() *cobra.Command {
	removeCmd.Flags().BoolVarP(&forceDelete, "force", "D", false, "Force delete the worktree (like git branch -D)")
	return removeCmd
}
