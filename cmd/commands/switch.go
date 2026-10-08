package commands

import (
	"fmt"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var switchBranch bool

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
		if branch, _ := cmd.Flags().GetBool("branch"); branch {
			branches, err := repo.AllBranches(repo.DefaultRemote())
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return pkg.GlobFilterComplete(args, branches, toComplete), cobra.ShellCompDirectiveNoFileComp
		}

		args = append(args, repo.CurrentWorktree.Name)
		args = append(args, repo.CurrentWorktree.Branch)

		return pkg.GlobFilterComplete(args, repo.WorktreeAliases(), toComplete), cobra.ShellCompDirectiveNoFileComp
	}),
	RunE: pkg.RepoCommand(func(repo *pkg.Repo, cmd *cobra.Command, args []string) error {
		// Branch mode: stay in the current directory and just
		// git-switch the branch. No ChangeDirectory, so the wrk
		// wrapper does not cd.
		if switchBranch {
			var target string
			if len(args) > 0 {
				target = args[0]
			} else if repo.IsMainWorktree(repo.CurrentWorktree) {
				// In the main worktree, switch to the remote's
				// default branch (origin/HEAD, resolved offline).
				def, err := repo.DefaultBranch()
				if err != nil {
					return err
				}
				target = def
			} else {
				// In a linked worktree, switch to the branch
				// matching the directory name.
				target = repo.CurrentWorktree.Name
			}

			current, _ := repo.CurrentBranch(repo.CurrentWorktree)
			if current == target {
				fmt.Printf("Already on '%s'\n", target)
				return nil
			}

			if !repo.BranchExists(target) {
				return fmt.Errorf("branch '%s' does not exist", target)
			}

			if err := repo.SwitchBranch(repo.CurrentWorktree, target); err != nil {
				return err
			}
			fmt.Printf("Switched to '%s'\n", target)
			return nil
		}

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
	switchCmd.Flags().BoolVarP(&switchBranch, "branch", "b", false, "Switch branch in place without changing directory (defaults: main worktree switches to remote default branch, linked worktree switches to branch matching directory name)")
	return switchCmd
}
