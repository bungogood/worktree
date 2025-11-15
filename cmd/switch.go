package cmd

import (
	"fmt"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch <branch>",
	Short: "Switch to a worktree",
	Long:  `Switch to an existing worktree by branch name.`,
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: pkg.RepoValidArgsFunction(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Return list of branch names
		var branches []string
		for _, wt := range repo.Worktrees {
			branches = append(branches, wt.Branch)
		}
		return branches, cobra.ShellCompDirectiveNoFileComp
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		branch := args[0]

		// Find the worktree
		worktree := repo.FindWorktree(branch)
		if worktree == nil {
			return fmt.Errorf("no worktree found '%s'", branch)
		}

		// Switch to the worktree
		pkg.ChangeDirectory(worktree.Path)
		return nil
	}),
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
