package cmd

import (
	"fmt"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <branch>",
	Short: "Create a new branch as a worktree",
	Long:  `Creates a new branch in a new worktree and navigates to it.`,
	Args:  cobra.ExactArgs(1),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		branch := args[0]

		// Create the new branch
		worktreePath, err := repo.CreateNewBranch(branch)
		if err != nil {
			return err
		}

		fmt.Printf("Worktree created: %s\n", branch)
		pkg.ChangeDirectory(worktreePath)
		return nil
	}),
}

func init() {
	rootCmd.AddCommand(newCmd)
}
