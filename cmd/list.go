package cmd

import (
	"fmt"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var (
	noColor bool
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

		// Determine if we should use color (check if stdout is a terminal)
		useColor := !noColor && pkg.IsTerminal()

		// Get sorted worktrees (main first, current second, then alphabetically)
		worktrees := repo.SortedWorktrees()

		// Display each worktree
		for _, wt := range worktrees {
			display := repo.GetWorktreeDisplay(&wt, useColor)
			fmt.Println(display)
		}
		return nil
	}),
}

func init() {
	listCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.AddCommand(listCmd)
}
