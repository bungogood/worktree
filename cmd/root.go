package cmd

import (
	"fmt"
	"os"

	"github.com/bungogood/worktree/cmd/commands"
	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Git worktree manager",
	Long:  `A CLI tool for managing git worktrees with automatic organization and navigation.`,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&pkg.GlobalFlags.Verbose, "verbose", "v", false, "Show all git commands being executed")

	// Register all commands
	rootCmd.AddCommand(commands.NewAddCmd())
	rootCmd.AddCommand(commands.NewNewCmd())
	rootCmd.AddCommand(commands.NewListCmd())
	rootCmd.AddCommand(commands.NewRemoveCmd())
	rootCmd.AddCommand(commands.NewSwitchCmd())
	rootCmd.AddCommand(commands.NewIgnoreCmd())
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
