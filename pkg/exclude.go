package pkg

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (r *Repo) excludePath() string {
	return filepath.Join(r.MainWorktree.Path, ".git", "info", "exclude")
}

// ExcludePattern adds a pattern to .git/info/exclude
func (r *Repo) ExcludePattern(pattern string) error {
	excludePath := r.excludePath()

	// Check if exclude file exists
	if _, err := os.Stat(excludePath); os.IsNotExist(err) {
		return fmt.Errorf("exclude file does not exist: %s", excludePath)
	}

	// Read all lines (including comments and empty lines)
	content, err := os.ReadFile(excludePath)
	if err != nil {
		return fmt.Errorf("failed to read exclude file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Check if pattern already exists
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == pattern {
			return fmt.Errorf("pattern already excluded")
		}
	}

	// Append pattern to the end
	// If file doesn't end with newline, add one before the pattern
	newContent := string(content)
	if len(content) > 0 && content[len(content)-1] != '\n' {
		newContent += "\n"
	}
	newContent += pattern + "\n"

	// Write back the file (preserving permissions)
	f, err := os.OpenFile(excludePath, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return fmt.Errorf("failed to open exclude file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(newContent); err != nil {
		return fmt.Errorf("failed to write exclude file: %w", err)
	}

	return nil
}

// UnexcludePattern removes a pattern from .git/info/exclude
func (r *Repo) UnexcludePattern(pattern string) error {
	excludePath := r.excludePath()

	// Read all lines (including comments and empty lines)
	content, err := os.ReadFile(excludePath)
	if err != nil {
		return fmt.Errorf("failed to read exclude file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Filter out lines that match the pattern (preserve comments and empty lines)
	found := false
	var newLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == pattern {
			found = true
			// Skip this line (don't add it to newLines)
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		return fmt.Errorf("pattern not found in exclude list")
	}

	// Write back the filtered content (preserving permissions)
	newContent := strings.Join(newLines, "\n")
	f, err := os.OpenFile(excludePath, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return fmt.Errorf("failed to open exclude file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(newContent); err != nil {
		return fmt.Errorf("failed to write exclude file: %w", err)
	}

	return nil
}

// ListExcludedPatterns returns all exclude patterns
func (r *Repo) ListExcludedPatterns() ([]string, error) {
	f, err := os.Open(r.excludePath())
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to open exclude file: %w", err)
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read exclude file: %w", err)
	}

	return patterns, nil
}

// PrintExcludedPatterns displays all excluded patterns
func (r *Repo) PrintExcludedPatterns() error {
	patterns, err := r.ListExcludedPatterns()
	if err != nil {
		return err
	}

	if len(patterns) == 0 {
		fmt.Println("No excluded patterns")
		return nil
	}

	for _, pattern := range patterns {
		fmt.Printf("%s\n", pattern)
	}

	return nil
}
