package cmd

import (
	"github.com/bungogood/worktree/cmd/commands"
	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Git worktree manager",
	Long:  `A CLI tool for managing git worktrees with automatic organization and navigation.`,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&pkg.GlobalFlags.Verbose, "verbose", "v", false, "Show all git commands being executed")

	// Register all commands
	RootCmd.AddCommand(commands.NewAddCmd())
	RootCmd.AddCommand(commands.NewNewCmd())
	RootCmd.AddCommand(commands.NewListCmd())
	RootCmd.AddCommand(commands.NewRemoveCmd())
	RootCmd.AddCommand(commands.NewSwitchCmd())
	RootCmd.AddCommand(commands.NewSkipCmd())
}
