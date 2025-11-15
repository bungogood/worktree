package cmd

import (
	"fmt"
	"os"

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
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
