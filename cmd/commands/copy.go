package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var (
	sourceWorktree string
	alwaysCopy     bool
	alwaysRemove   bool
)

var copyCmd = &cobra.Command{
	Use:   "copy <source-path> [dest-path]",
	Short: "Copy files from another worktree",
	Long: `Copy files or directories from another worktree to the current worktree. By default copies from the main worktree. If dest-path is not specified, uses source-path as the destination.

Use --always to add the path to the config so it's automatically copied to all new worktrees.
Use --always with no arguments to list all always-copy paths.
Use --always-rm to remove paths from the always-copy list.`,
	Args: cobra.MaximumNArgs(2),
	ValidArgsFunction: pkg.RepoCompletion(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {

		if alwaysRemove {
			if repo.Config == nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			// Filter out already specified args
			var completions []string
			for _, path := range repo.Config.Copy {
				if !slices.Contains(args, path) {
					completions = append(completions, path)
				}
			}

			return pkg.GlobFilterComplete(args, completions, toComplete), cobra.ShellCompDirectiveNoFileComp
		}

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

		return pkg.GlobFilterComplete(args, completions, toComplete), cobra.ShellCompDirectiveNoFileComp
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		// If --always flag is set with no args, list always-copy paths
		if alwaysCopy && len(args) == 0 {
			return repo.PrintAlwaysCopy()
		}

		// If --always flag is set with args, add to config
		if alwaysCopy {
			srcPath := args[0]
			if err := repo.AddAlwaysCopy(srcPath); err != nil {
				return err
			}
			fmt.Printf("Added '%s' to always-copy list\n", srcPath)
			return nil
		}

		// If --always-rm flag is set, remove from config
		if alwaysRemove {
			if len(args) == 0 {
				return fmt.Errorf("path required for --always-rm")
			}
			srcPath := args[0]
			if err := repo.RemoveAlwaysCopy(srcPath); err != nil {
				return err
			}
			fmt.Printf("Removed '%s' from always-copy list\n", srcPath)
			return nil
		}

		// Otherwise, perform the copy to current worktree
		if len(args) == 0 {
			return fmt.Errorf("source path required")
		}

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
	copyCmd.Flags().BoolVar(&alwaysCopy, "always", false, "Add path to config to automatically copy to new worktrees")
	copyCmd.Flags().BoolVar(&alwaysRemove, "always-rm", false, "Remove path from always-copy list")
	copyCmd.RegisterFlagCompletionFunc("from", pkg.RepoCompletion(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {
		return pkg.GlobFilterComplete(args, repo.WorktreeAliases(), toComplete), cobra.ShellCompDirectiveNoFileComp
	}))

	return copyCmd
}
