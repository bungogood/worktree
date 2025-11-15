package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch [directory]",
	Short: "Switch to a directory",
	Long:  `Switch to a specified directory using the wrk wrapper.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetDir := args[0]

		// Check if directory exists
		if info, err := os.Stat(targetDir); err != nil || !info.IsDir() {
			fmt.Fprintf(os.Stderr, "Error: '%s' is not a valid directory\n", targetDir)
			os.Exit(1)
		}

		// Get absolute path
		absPath, err := filepath.Abs(targetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not resolve path: %v\n", err)
			os.Exit(1)
		}

		// Output the directory path with delimiter
		fmt.Printf("Switching to %s\n", absPath)
		pkg.ChangeDirectory(absPath)
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
