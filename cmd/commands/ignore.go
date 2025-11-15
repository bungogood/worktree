package commands

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var (
	removeIgnore bool
)

var ignoreCmd = &cobra.Command{
	Use:   "ignore [file...]",
	Short: "Manage ignored file changes",
	Long: `Manage files whose changes should be ignored without modifying .gitignore. Uses git update-index --skip-worktree.
	
With no arguments, lists all ignored files.
With file arguments, marks files to have their changes ignored.
Use --rm flag to unignore files instead.`,
	ValidArgsFunction: pkg.RepoValidArgsFunction(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {
		// If --rm flag is set, suggest ignored files
		if removeIgnore {
			ignored, err := repo.ListIgnoredFiles()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			// Filter out already specified args
			var completions []string
			for _, file := range ignored {
				if !slices.Contains(args, file) {
					completions = append(completions, file)
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		}
		// Otherwise use default file completion
		return nil, cobra.ShellCompDirectiveDefault
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		// No args: list ignored files
		if len(args) == 0 {
			ignored, err := repo.ListIgnoredFiles()
			if err != nil {
				return err
			}

			if len(ignored) == 0 {
				fmt.Println("No files with ignored changes.")
				return nil
			}

			for _, file := range ignored {
				fmt.Println(file)
			}

			return nil
		}

		// With args: either ignore or unignore files
		var errors []string

		if removeIgnore {
			// Unignore files
			for _, file := range args {
				if err := repo.UnignoreFile(file); err != nil {
					errors = append(errors, fmt.Sprintf("  %s: %v", file, err))
				}
			}

			if len(errors) > 0 {
				return fmt.Errorf("failed to unignore %d file(s):\n%s", len(errors), strings.Join(errors, "\n"))
			}
		} else {
			// Ignore files
			for _, file := range args {
				if err := repo.IgnoreFile(file); err != nil {
					errors = append(errors, fmt.Sprintf("  %s: %v", file, err))
				}
			}

			if len(errors) > 0 {
				return fmt.Errorf("failed to ignore %d file(s):\n%s", len(errors), strings.Join(errors, "\n"))
			}
		}

		return nil
	}),
}

// NewIgnoreCmd returns the ignore command
func NewIgnoreCmd() *cobra.Command {
	ignoreCmd.Flags().BoolVar(&removeIgnore, "rm", false, "Remove files from ignore list")
	return ignoreCmd
}
