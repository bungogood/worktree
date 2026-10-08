package pkg

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const CD_DELIMITER = "wrk:cd:"

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

// ParseMaxAge parses an age threshold like "36h", "14d" or "4w".
// A bare number means days. Returns an error for empty, non-positive
// or malformed values.
func ParseMaxAge(s string) (time.Duration, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, fmt.Errorf("invalid age %q: expected like 36h, 14d or 4w", s)
	}

	multipliers := map[byte]time.Duration{
		'h': time.Hour,
		'd': 24 * time.Hour,
		'w': 7 * 24 * time.Hour,
	}

	multiplier := 24 * time.Hour
	number := s
	if last := s[len(s)-1]; last < '0' || last > '9' {
		m, ok := multipliers[last]
		if !ok {
			return 0, fmt.Errorf("invalid age %q: unknown unit %q, expected h, d or w", s, string(last))
		}
		multiplier = m
		number = s[:len(s)-1]
	}

	n, err := strconv.Atoi(number)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid age %q: expected like 36h, 14d or 4w", s)
	}

	return time.Duration(n) * multiplier, nil
}

// HumanizeAge renders a duration compactly ("45m", "3h", "5d", "2w3d").
func HumanizeAge(d time.Duration) string {
	days := int(d.Hours()) / 24
	if days >= 7 {
		if rest := days % 7; rest > 0 {
			return fmt.Sprintf("%dw%dd", days/7, rest)
		}
		return fmt.Sprintf("%dw", days/7)
	}
	if days >= 1 {
		return fmt.Sprintf("%dd", days)
	}
	if d >= time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
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
