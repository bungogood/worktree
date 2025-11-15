package cmd

import (
	"fmt"
	"os"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all worktrees",
	Long:    `Display all worktrees in the repository with their branches and paths.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load the repository
		repo, err := pkg.LoadRepo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(repo.Worktrees) == 0 {
			fmt.Println("No worktrees found.")
			return
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
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
