package cmd

import (
	"fmt"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all worktrees",
	Long:    `Display all worktrees in the repository with their branches and paths.`,
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		if len(repo.Worktrees) == 0 {
			fmt.Println("No worktrees found.")
			return nil
		}

		// Display each worktree
		for _, wt := range repo.Worktrees {
			marker := "  "
			if repo.IsMainWorktree(&wt) {
				marker = "* "
			} else if repo.CurrentWorktree != nil && wt.Path == repo.CurrentWorktree.Path {
				marker = "> "
			}

			fmt.Printf("%s%-20s %s\n", marker, wt.Branch, wt.Path)
		}
		return nil
	}),
}

func init() {
	rootCmd.AddCommand(listCmd)
}
