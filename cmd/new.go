package cmd

import (
	"fmt"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <branch> [name]",
	Short: "Create a new branch as a worktree",
	Long:  `Creates a new branch in a new worktree and navigates to it. Optionally specify a custom directory name.`,
	Args:  cobra.RangeArgs(1, 2),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		branch := args[0]
		name := branch
		if len(args) > 1 {
			name = args[1]
		}

		// Create the new branch
		worktree, err := repo.CreateNewBranch(branch, name)
		if err != nil {
			return err
		}

		fmt.Printf("Worktree created: '%s'\n", name)
		pkg.ChangeDirectory(worktree.Path)
		return nil
	}),
}

func init() {
	rootCmd.AddCommand(newCmd)
}
