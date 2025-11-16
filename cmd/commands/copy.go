package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var (
	sourceWorktree string
)

var copyCmd = &cobra.Command{
	Use:   "copy <source-path> [dest-path]",
	Short: "Copy files from another worktree",
	Long:  `Copy files or directories from another worktree to the current worktree. By default copies from the main worktree. If dest-path is not specified, uses source-path as the destination.`,
	Args:  cobra.RangeArgs(1, 2),
	ValidArgsFunction: pkg.RepoCompletion(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {

		// Only provide completion for the first argument (source path)
		if len(args) == 1 {
			// For second argument, use default file completion in current directory
			return nil, cobra.ShellCompDirectiveDefault
		} else if len(args) > 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Determine source worktree
		sourceWt, err := repo.GetSourceWorktree(sourceWorktree)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Build the path to complete in the source worktree
		completePath := filepath.Join(sourceWt.Path, toComplete)

		// Get directory to list
		var dirToList string
		if toComplete == "" || toComplete[len(toComplete)-1] == '/' {
			dirToList = completePath
		} else {
			dirToList = filepath.Dir(completePath)
		}

		// Read directory entries
		entries, err := os.ReadDir(dirToList)
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}

		// Build completion list
		var completions []string
		baseDir := filepath.Dir(toComplete)
		if baseDir == "." {
			baseDir = ""
		} else {
			baseDir = baseDir + "/"
		}

		for _, entry := range entries {
			name := entry.Name()

			if entry.IsDir() {
				completions = append(completions, baseDir+name+"/")
			} else {
				completions = append(completions, baseDir+name)
			}
		}

		return completions, cobra.ShellCompDirectiveNoSpace
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		if repo.CurrentWorktree == nil {
			return fmt.Errorf("not currently in a worktree")
		}

		srcPath := args[0]
		dstPath := srcPath
		if len(args) == 2 {
			dstPath = args[1]
		}

		// Determine source worktree
		sourceWt, err := repo.GetSourceWorktree(sourceWorktree)
		if err != nil {
			return err
		}

		// Perform the copy
		return repo.CopyFromWorktree(sourceWt, repo.CurrentWorktree, srcPath, dstPath)
	}),
}

// NewCopyCmd returns the copy command
func NewCopyCmd() *cobra.Command {
	copyCmd.Flags().StringVarP(&sourceWorktree, "from", "f", "", "Source worktree (defaults to main worktree)")
	copyCmd.RegisterFlagCompletionFunc("from", pkg.RepoCompletion(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {
		// Return list of worktrees excluding current
		var trees []string
		for _, wt := range repo.Worktrees {
			if repo.CurrentWorktree == nil || wt.Name != repo.CurrentWorktree.Name {
				trees = append(trees, wt.Name)
				trees = append(trees, wt.Branch)
			}
		}
		return trees, cobra.ShellCompDirectiveNoFileComp
	}))

	return copyCmd
}
