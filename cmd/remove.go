package cmd

import (
	"fmt"
	"slices"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var (
	forceDelete bool
)

var removeCmd = &cobra.Command{
	Use:     "remove [branch]",
	Aliases: []string{"rm"},
	Short:   "Remove a worktree",
	Long:    `Remove a worktree. If no branch is specified, removes the current worktree. Cannot remove the main worktree.`,
	Args:    cobra.MaximumNArgs(1),
	ValidArgsFunction: pkg.RepoValidArgsFunction(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {
		// Return list of branch names
		var branches []string
		for _, wt := range repo.Worktrees {
			if !repo.IsMainWorktree(&wt) && !slices.Contains(args, wt.Branch) {
				branches = append(branches, wt.Branch)
			}
		}
		return branches, cobra.ShellCompDirectiveNoFileComp
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		var targetWorktree *pkg.Worktree

		// Determine which worktree to remove
		if len(args) == 0 {
			// No tree specified, use current worktree
			if repo.CurrentWorktree == nil {
				return fmt.Errorf("not currently in a worktree")
			}
			targetWorktree = repo.CurrentWorktree
		} else {
			// Tree specified, find it
			tree := args[0]
			targetWorktree = repo.FindWorktree(tree)
			if targetWorktree == nil {
				return fmt.Errorf("no worktree found '%s'", tree)
			}
		}

		// Remove the worktree
		err := repo.RemoveWorktree(targetWorktree, forceDelete)
		if err != nil {
			return err
		}

		fmt.Printf("Worktree removed: '%s'\n", targetWorktree.Name)

		// If we just removed the current worktree, cd to the main worktree
		if targetWorktree == repo.CurrentWorktree {
			pkg.ChangeDirectory(repo.MainWorktree.Path)
		}
		return nil
	}),
}

func init() {
	removeCmd.Flags().BoolVarP(&forceDelete, "force", "D", false, "Force delete the worktree (like git branch -D)")
	rootCmd.AddCommand(removeCmd)
}
