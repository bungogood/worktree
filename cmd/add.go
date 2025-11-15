package cmd

import (
	"fmt"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <branch>",
	Short: "Add an existing branch as a worktree",
	Long:  `Creates a new worktree for an existing local or remote branch and navigates to it.`,
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: pkg.RepoValidArgsFunction(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		branches, err := repo.AllBranches()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		// Filter out all branches that already have worktrees
		var filtered []string
		for _, br := range branches {
			if repo.FindWorktree(br) == nil {
				filtered = append(filtered, br)
			}
		}
		return filtered, cobra.ShellCompDirectiveNoFileComp
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		branch := args[0]

		// Try to add the existing branch
		worktreePath, err := repo.AddExistingBranch(branch)
		if err != nil {
			// Check if it's the "already exists" error
			if err.Error() == "worktree already exists" {
				existing := repo.FindWorktreeByBranch(branch)
				fmt.Printf("Worktree for branch '%s' already exists at: %s\n", branch, existing.Path)
				pkg.ChangeDirectory(existing.Path)
				return nil
			}
			return err
		}

		fmt.Printf("Worktree created: %s\n", branch)
		pkg.ChangeDirectory(worktreePath)
		return nil
	}),
}

func init() {
	rootCmd.AddCommand(addCmd)
}
