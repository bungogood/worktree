package commands

import (
	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch [branch]",
	Short: "Switch to a worktree",
	Long:  `Switch to an existing worktree by branch name. If no branch is specified, switches to the main worktree.`,
	Args:  cobra.MaximumNArgs(1),
	ValidArgsFunction: pkg.RepoCompletion(func(
		repo *pkg.Repo,
		cmd *cobra.Command,
		args []string,
		toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Remove current worktree from completions
		args = append(args, repo.CurrentWorktree.Name)
		args = append(args, repo.CurrentWorktree.Branch)

		return pkg.GlobFilterComplete(args, repo.WorktreeAliases(), toComplete), cobra.ShellCompDirectiveNoFileComp
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		var worktree *pkg.Worktree

		// If no args, switch to main worktree
		if len(args) == 0 {
			worktree = repo.MainWorktree
		} else {
			pattern := args[0]
			wt, err := repo.FindWorktree(pattern)
			if err != nil {
				return err
			}
			worktree = wt
		}

		// Switch to the worktree
		pkg.ChangeDirectory(worktree.Path)
		return nil
	}),
}

// NewSwitchCmd returns the switch command
func NewSwitchCmd() *cobra.Command {
	return switchCmd
}
