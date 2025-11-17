package cmd

import (
	"github.com/bungogood/worktree/cmd/commands"
	"github.com/bungogood/worktree/pkg"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "wrk",
	Short: "Git worktree manager",
	Long:  `A CLI tool for managing git worktrees with automatic organisation and navigation.`,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set color mode based on flag after parsing
		color.NoColor = pkg.GlobalFlags.NoColor || !pkg.IsTerminal()
	},
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&pkg.GlobalFlags.Verbose, "verbose", "v", false, "Show all git commands being executed")
	RootCmd.PersistentFlags().BoolVarP(&pkg.GlobalFlags.NoColor, "no-color", "", false, "Disable colored output")

	// Register all commands
	RootCmd.AddCommand(commands.NewAddCmd())
	RootCmd.AddCommand(commands.NewNewCmd())
	RootCmd.AddCommand(commands.NewListCmd())
	RootCmd.AddCommand(commands.NewRemoveCmd())
	RootCmd.AddCommand(commands.NewSwitchCmd())
	RootCmd.AddCommand(commands.NewSkipCmd())
	RootCmd.AddCommand(commands.NewExcludeCmd())
	RootCmd.AddCommand(commands.NewCopyCmd())
}
