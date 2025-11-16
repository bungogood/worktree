package commands

import (
	"fmt"
	"strings"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var (
	removeSkip bool
	localSkip  bool
)

var skipCmd = &cobra.Command{
	Use:   "skip [file...]",
	Short: "Manage skipped file changes",
	Long: `Manage files whose changes should be skipped without modifying .gitignore. Uses git update-index --skip-worktree.
	
With no arguments, lists all skipped files.
With file arguments, marks files to have their changes skipped.
Use --rm flag to unskip files instead.`,
	ValidArgsFunction: pkg.RepoCompletion(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {
		// If --rm flag is set, suggest skipped files
		if removeSkip {
			skipped, err := repo.ListSkippedFiles()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			return pkg.GlobFilterComplete(args, skipped, toComplete), cobra.ShellCompDirectiveNoFileComp
		}
		// Otherwise use default file completion
		return nil, cobra.ShellCompDirectiveDefault
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		// No args: list skipped files
		if len(args) == 0 {
			return repo.PrintSkippedFiles()
		}

		// With args: either skip or unskip files
		var errors []string

		if removeSkip {
			// Unskip files
			for _, file := range args {
				var err error
				if localSkip {
					err = repo.LocalUnskipFile(file)
				} else {
					err = repo.UnskipFile(file)
				}
				if err != nil {
					errors = append(errors, fmt.Sprintf("  %s: %v", file, err))
				}
			}

			if len(errors) > 0 {
				return fmt.Errorf("failed to unskip %d file(s):\n%s", len(errors), strings.Join(errors, "\n"))
			}
		} else {
			// Skip files
			for _, file := range args {
				var err error
				if localSkip {
					err = repo.LocalSkipFile(file)
				} else {
					err = repo.SkipFile(file)
				}
				if err != nil {
					errors = append(errors, fmt.Sprintf("  %s: %v", file, err))
				}
			}

			if len(errors) > 0 {
				return fmt.Errorf("failed to skip %d file(s):\n%s", len(errors), strings.Join(errors, "\n"))
			}
		}

		return nil
	}),
}

// NewSkipCmd returns the skip command
func NewSkipCmd() *cobra.Command {
	skipCmd.Flags().BoolVar(&removeSkip, "rm", false, "Remove files from skip list")
	skipCmd.Flags().BoolVar(&localSkip, "local", false, "Only affect current worktree (does not work in main worktree)")
	return skipCmd
}
