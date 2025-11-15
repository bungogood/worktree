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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		repo, err := pkg.LoadRepo()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Return list of branch names
		branches := make([]string, 0, len(repo.Worktrees))
		for _, wt := range repo.Worktrees {
			branches = append(branches, wt.Branch)
		}
		return branches, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		branch := args[0]

		// Find the worktree
		worktree := repo.FindWorktree(branch)
		if worktree == nil {
			return fmt.Errorf("no worktree found for branch '%s'", branch)
		}

		// Switch to the worktree
		fmt.Printf("Switching to worktree: %s\n", worktree.Branch)
		pkg.ChangeDirectory(worktree.Path)
		return nil
	}),
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
