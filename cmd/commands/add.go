package commands

import (
	"fmt"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <branch> [name]",
	Short: "Add an existing branch as a worktree",
	Long:  `Creates a new worktree for an existing local or remote branch and navigates to it. Optionally specify a custom directory name.`,
	Args:  cobra.RangeArgs(1, 2),
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
		for _, branch := range branches {
			if repo.FindWorktreeByBranch(branch) == nil {
				filtered = append(filtered, branch)
			}
		}
		return filtered, cobra.ShellCompDirectiveNoFileComp
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		branch := args[0]
		name := branch
		if len(args) > 1 {
			name = args[1]
		}

		// Try to add the existing branch
		worktree, err := repo.AddExistingBranch(branch, name)
		if err != nil {
			return err
		}

		fmt.Printf("Worktree created: '%s'\n", name)
		pkg.ChangeDirectory(worktree.Path)
		return nil
	}),
}

// NewAddCmd returns the add command
func NewAddCmd() *cobra.Command {
	return addCmd
}
