package cmd

import (
	"fmt"
	"os"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <branch>",
	Short: "Add an existing branch as a worktree",
	Long:  `Creates a new worktree for an existing local or remote branch and navigates to it.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		branch := args[0]

		// Load the repository
		repo, err := pkg.LoadRepo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Check if worktree already exists
		if existing := repo.FindWorktree(branch); existing != nil {
			fmt.Printf("Worktree for branch '%s' already exists at: %s\n", branch, existing.Path)
			pkg.ChangeDirectory(existing.Path)
			return
		}

		// Ensure the .{repo}.worktrees directory exists
		if err := repo.EnsureWorktreesDir(); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating worktrees directory: %v\n", err)
			os.Exit(1)
		}

		// Get the path for the new worktree
		worktreePath := repo.GetWorktreePath(branch)

		// Check if branch exists locally
		_, err = repo.RunGitCommand(nil, "rev-parse", "--verify", branch)
		localExists := err == nil

		// If not local, check if it exists on remote
		if !localExists {
			_, err = repo.RunGitCommand(nil, "rev-parse", "--verify", fmt.Sprintf("origin/%s", branch))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Branch '%s' does not exist locally or on remote\n", branch)
				os.Exit(1)
			}
			// Branch exists on remote, create worktree with tracking
			fmt.Printf("Creating worktree for remote branch '%s'...\n", branch)
			_, err = repo.RunGitCommand(nil, "worktree", "add", "-b", branch, worktreePath, fmt.Sprintf("origin/%s", branch))
		} else {
			// Branch exists locally
			fmt.Printf("Creating worktree for branch '%s'...\n", branch)
			_, err = repo.RunGitCommand(nil, "worktree", "add", worktreePath, branch)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating worktree: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Worktree created at: %s\n", worktreePath)
		pkg.ChangeDirectory(worktreePath)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
