package cmd

import (
	"fmt"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch [branch]",
	Short: "Switch to a worktree",
	Long:  `Switch to an existing worktree by branch name. If no branch is specified, switches to the main worktree.`,
	Args:  cobra.MaximumNArgs(1),
	ValidArgsFunction: pkg.RepoValidArgsFunction(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Return list of worktrees
		var trees []string
		for _, wt := range repo.Worktrees {
			trees = append(trees, wt.Name)
			trees = append(trees, wt.Branch)
		}
		return trees, cobra.ShellCompDirectiveNoFileComp
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		var worktree *pkg.Worktree

		// If no args, switch to main worktree
		if len(args) == 0 {
			worktree = repo.MainWorktree
		} else {
			tree := args[0]
			worktree = repo.FindWorktree(tree)
			if worktree == nil {
				return fmt.Errorf("no worktree found '%s'", tree)
			}
		}

		// Switch to the worktree
		pkg.ChangeDirectory(worktree.Path)
		return nil
	}),
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
