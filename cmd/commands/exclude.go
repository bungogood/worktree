package commands

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var (
	removeExclude bool
)

var excludeCmd = &cobra.Command{
	Use:   "exclude [file...]",
	Short: "Manage excluded untracked files",
	Long: `Manage files that should be excluded from git without modifying .gitignore. Uses .git/info/exclude.
	
With no arguments, lists all excluded files.
With file arguments, adds patterns to exclude.
Use --rm flag to remove exclusions instead.`,
	ValidArgsFunction: pkg.RepoCompletion(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {
		// If --rm flag is set, suggest excluded patterns
		if removeExclude {
			patterns, err := repo.ListExcludedPatterns()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			// Filter out already specified args
			var completions []string
			for _, pattern := range patterns {
				if !slices.Contains(args, pattern) {
					completions = append(completions, pattern)
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		}
		// Otherwise use default file completion
		return nil, cobra.ShellCompDirectiveDefault
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		// No args: list excluded patterns
		if len(args) == 0 {
			return repo.PrintExcludedPatterns()
		}

		// With args: either exclude or unexclude patterns
		var errors []string

		if removeExclude {
			// Remove exclusions
			for _, pattern := range args {
				err := repo.UnexcludePattern(pattern)
				if err != nil {
					errors = append(errors, fmt.Sprintf("  %s: %v", pattern, err))
				}
			}

			if len(errors) > 0 {
				return fmt.Errorf("failed to unexclude %d pattern(s):\n%s", len(errors), strings.Join(errors, "\n"))
			}
		} else {
			// Add exclusions
			for _, pattern := range args {
				err := repo.ExcludePattern(pattern)
				if err != nil {
					errors = append(errors, fmt.Sprintf("  %s: %v", pattern, err))
				}
			}

			if len(errors) > 0 {
				return fmt.Errorf("failed to exclude %d pattern(s):\n%s", len(errors), strings.Join(errors, "\n"))
			}
		}

		return nil
	}),
}

// NewExcludeCmd returns the exclude command
func NewExcludeCmd() *cobra.Command {
	excludeCmd.Flags().BoolVar(&removeExclude, "rm", false, "Remove patterns from exclude list")
	return excludeCmd
}
