package commands

import (
	"fmt"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:               "list",
	Aliases:           []string{"ls"},
	Short:             "List all worktrees",
	Long:              `Display all worktrees in the repository with their branches and paths.`,
	Args:              cobra.NoArgs,
	ValidArgsFunction: cobra.NoFileCompletions,
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		if len(repo.Worktrees) == 0 {
			fmt.Println("No worktrees found.")
			return nil
		}

		// Get sorted worktrees (main first, current second, then alphabetically)
		worktrees := repo.SortedWorktrees()

		// Display each worktree
		for _, wt := range worktrees {
			display := repo.GetWorktreeDisplay(&wt)
			fmt.Println(display)
		}
		return nil
	}),
}

// NewListCmd returns the list command
func NewListCmd() *cobra.Command {
	return listCmd
}
