package cmd

import (
	"fmt"
	"os"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var (
	forceDelete bool
)

var removeCmd = &cobra.Command{
	Use:     "remove [branch]",
	Aliases: []string{"rm"},
	Short:   "Remove a worktree",
	Long:    `Remove a worktree. If no branch is specified, removes the current worktree. Cannot remove the main worktree.`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Load the repository
		repo, err := pkg.LoadRepo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var targetWorktree *pkg.Worktree

		// Determine which worktree to remove
		if len(args) == 0 {
			// No branch specified, use current worktree
			if repo.CurrentWorktree == nil {
				fmt.Fprintf(os.Stderr, "Error: Not currently in a worktree\n")
				os.Exit(1)
			}
			targetWorktree = repo.CurrentWorktree
		} else {
			// Branch specified, find it
			branch := args[0]
			targetWorktree = repo.FindWorktree(branch)
			if targetWorktree == nil {
				fmt.Fprintf(os.Stderr, "Error: No worktree found for branch '%s'\n", branch)
				os.Exit(1)
			}
		}

		// Protect the main worktree
		if repo.IsMainWorktree(targetWorktree) {
			fmt.Fprintf(os.Stderr, "Error: Cannot remove the main worktree (contains .git directory)\n")
			os.Exit(1)
		}

		// Build git worktree remove command
		removeArgs := []string{"worktree", "remove", targetWorktree.Path}
		if forceDelete {
			removeArgs = append(removeArgs, "--force")
		}

		// Remove the worktree
		fmt.Printf("Removing worktree for branch '%s'...\n", targetWorktree.Branch)
		_, err = repo.RunGitCommand(nil, removeArgs...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error removing worktree: %v\n", err)
			fmt.Fprintf(os.Stderr, "Tip: Use -D flag to force delete\n")
			os.Exit(1)
		}

		fmt.Printf("Worktree removed: %s\n", targetWorktree.Path)

		// If we just removed the current worktree, cd to the main worktree
		if targetWorktree == repo.CurrentWorktree {
			fmt.Println("Returning to main worktree...")
			pkg.ChangeDirectory(repo.MainWorktree.Path)
		}
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&forceDelete, "force", "D", false, "Force delete the worktree (like git branch -D)")
	rootCmd.AddCommand(removeCmd)
}
