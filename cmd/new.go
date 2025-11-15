package cmd

import (
	"fmt"
	"os"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <branch>",
	Short: "Create a new branch as a worktree",
	Long:  `Creates a new branch in a new worktree and navigates to it.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		branch := args[0]

		// Load the repository
		repo, err := pkg.LoadRepo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Check if branch already exists
		_, err = repo.RunGitCommand(nil, "rev-parse", "--verify", branch)
		if err == nil {
			fmt.Fprintf(os.Stderr, "Error: Branch '%s' already exists. Use 'add' command instead.\n", branch)
			os.Exit(1)
		}

		// Check if worktree already exists (shouldn't happen, but be safe)
		if existing := repo.FindWorktree(branch); existing != nil {
			fmt.Fprintf(os.Stderr, "Error: Worktree for branch '%s' already exists at: %s\n", branch, existing.Path)
			os.Exit(1)
		}

		// Ensure the .{repo}.worktrees directory exists
		if err := repo.EnsureWorktreesDir(); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating worktrees directory: %v\n", err)
			os.Exit(1)
		}

		// Get the path for the new worktree
		worktreePath := repo.GetWorktreePath(branch)

		// Create the new worktree with a new branch
		fmt.Printf("Creating new branch '%s' in worktree...\n", branch)
		_, err = repo.RunGitCommand(nil, "worktree", "add", "-b", branch, worktreePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating worktree: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Worktree created at: %s\n", worktreePath)
		pkg.ChangeDirectory(worktreePath)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
