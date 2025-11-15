package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "worktree",
	Short: "A tiny hello world Cobra app",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello, world!")
	},
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
