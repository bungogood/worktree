package commands

import (
	"fmt"
	"strings"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var ignoreCmd = &cobra.Command{
	Use:   "ignore",
	Short: "Manage ignored file changes",
	Long:  `Manage files whose changes should be ignored without modifying .gitignore. Uses git update-index --skip-worktree.`,
}

var ignoreAddCmd = &cobra.Command{
	Use:   "add <file>...",
	Short: "Ignore changes to files",
	Long:  `Mark files to have their changes ignored. Git will not show these files as modified.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		var errors []string

		for _, file := range args {
			if err := repo.IgnoreFile(file); err != nil {
				errors = append(errors, fmt.Sprintf("  %s: %v", file, err))
			}
		}

		if len(errors) > 0 {
			return fmt.Errorf("failed to ignore %d file(s):\n%s", len(errors), strings.Join(errors, "\n"))
		}

		return nil
	}),
}

var ignoreRemoveCmd = &cobra.Command{
	Use:     "remove <file>...",
	Aliases: []string{"rm"},
	Short:   "Stop ignoring changes to files",
	Long:    `Unmark files so their changes will be tracked again.`,
	Args:    cobra.MinimumNArgs(1),
	ValidArgsFunction: pkg.RepoValidArgsFunction(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {
		ignored, err := repo.ListIgnoredFiles()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Filter out already specified args
		var completions []string
		for _, file := range ignored {
			if !strings.HasPrefix(file, toComplete) {
				continue
			}
			skip := false
			for _, arg := range args {
				if arg == file {
					skip = true
					break
				}
			}
			if !skip {
				completions = append(completions, file)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		var errors []string

		for _, file := range args {
			if err := repo.UnignoreFile(file); err != nil {
				errors = append(errors, fmt.Sprintf("  %s: %v", file, err))
			}
		}

		if len(errors) > 0 {
			return fmt.Errorf("failed to unignore %d file(s):\n%s", len(errors), strings.Join(errors, "\n"))
		}

		return nil
	}),
}

var ignoreListCmd = &cobra.Command{
	Use:               "list",
	Aliases:           []string{"ls"},
	Short:             "List files with ignored changes",
	Long:              `Show all files that are currently marked to have their changes ignored.`,
	Args:              cobra.NoArgs,
	ValidArgsFunction: cobra.NoFileCompletions,
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
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
	}),
}

// NewIgnoreCmd returns the ignore command with its subcommands
func NewIgnoreCmd() *cobra.Command {
	ignoreCmd.AddCommand(ignoreAddCmd)
	ignoreCmd.AddCommand(ignoreRemoveCmd)
	ignoreCmd.AddCommand(ignoreListCmd)
	return ignoreCmd
}
