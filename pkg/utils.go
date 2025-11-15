package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

const CD_DELIMITER = "__WORKTREE_CD__"

type GFlags struct {
	Verbose bool
}

var GlobalFlags GFlags

// ChangeDirectory outputs the directory change command for the wrk wrapper
func ChangeDirectory(path string) {
	fmt.Printf("%s%s\n", CD_DELIMITER, path)
}

// RepoCommand wraps a command function that needs a loaded repository
// Returns a RunE function that can be used directly in cobra commands
func RepoCommand(fn func(*Repo, *cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		repo, err := LoadRepo()
		if err != nil {
			return err
		}
		return fn(repo, cmd, args)
	}
}

func RepoValidArgsFunction(fn func(*Repo, *cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		repo, err := LoadRepo()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return fn(repo, cmd, args, toComplete)
	}
}

func RunCommand(name string, args ...string) ([]byte, error) {
	if GlobalFlags.Verbose {
		fmt.Fprintf(os.Stderr, "Running: %s %s\n", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return output, err
}

// IsTerminal checks if stdout is a terminal
func IsTerminal() bool {
	// Check if TERM is set (more reliable when output is captured)
	term := os.Getenv("TERM")
	if term != "" && term != "dumb" {
		return true
	}

	// Fallback to checking if stdout is a character device
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}

	return false
}
