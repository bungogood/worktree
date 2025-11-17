package pkg

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const CD_DELIMITER = "__WORKTREE_CD__"

type GFlags struct {
	Verbose bool
	NoColor bool
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

func RepoCompletion(fn func(*Repo, *cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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

func CopyPath(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// It's a file
		return copyFileWithMode(path, targetPath, info.Mode())
	})
}

func copyFileWithMode(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func GlobFilter(pattern string, candidates []string) []string {
	var matches []string
	for _, candidate := range candidates {
		if matched, _ := filepath.Match(pattern, candidate); matched {
			matches = append(matches, candidate)
		}
	}
	return matches
}

func GlobFilterComplete(args []string, completions []string,
	toComplete string) []string {
	pattern := toComplete + "*"
	var matches []string
	for _, candidate := range completions {
		if slices.Contains(args, candidate) {
			continue
		}

		if matched, _ := filepath.Match(pattern, candidate); matched {
			matches = append(matches, candidate)
		} else if matched, _ := filepath.Match(pattern+"/", candidate); matched {
			matches = append(matches, candidate)
		}
	}
	return matches
}
