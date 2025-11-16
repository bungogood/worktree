package pkg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Copy                     []string `yaml:"copy"`
	Commands                 []string `yaml:"commands"`
	DeleteBranchWithWorktree bool     `yaml:"deleteBranchWithWorktree"`
}

// ConfigPath returns the path to the config file
func (r *Repo) ConfigPath() string {
	return filepath.Join(r.WorktreesDir, ".config.yml")
}

// LoadConfig loads the config file if it exists
func (r *Repo) LoadConfig() (*Config, error) {
	configPath := r.ConfigPath()

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return nil, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the config file
func (r *Repo) SaveConfig() error {
	configPath := r.ConfigPath()

	// Ensure worktrees directory exists
	if err := os.MkdirAll(r.WorktreesDir, 0755); err != nil {
		return fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	data, err := yaml.Marshal(r.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// AddAlwaysCopy adds a path to the copy list
func (r *Repo) AddAlwaysCopy(path string) error {
	if r.Config == nil {
		r.Config = &Config{}
	}

	// Check if path already exists
	for _, existing := range r.Config.Copy {
		if existing == path {
			return fmt.Errorf("path already in copy list")
		}
	}

	r.Config.Copy = append(r.Config.Copy, path)
	return r.SaveConfig()
}

// RemoveAlwaysCopy removes a path from the copy list
func (r *Repo) RemoveAlwaysCopy(path string) error {
	if r.Config == nil || len(r.Config.Copy) == 0 {
		return fmt.Errorf("path not found in copy list")
	}

	// Find and remove the path
	found := false
	var newCopy []string
	for _, existing := range r.Config.Copy {
		if existing == path {
			found = true
		} else {
			newCopy = append(newCopy, existing)
		}
	}

	if !found {
		return fmt.Errorf("path not found in copy list")
	}

	r.Config.Copy = newCopy
	return r.SaveConfig()
}

// ApplyAlwaysCopy applies all always-copy paths to a worktree
func (r *Repo) ApplyAlwaysCopy(destWt *Worktree) error {
	if r.Config == nil || len(r.Config.Copy) == 0 {
		return nil
	}

	var errors []string
	for _, path := range r.Config.Copy {
		if err := r.CopyFromWorktree(r.MainWorktree, destWt, path, path); err != nil {
			errors = append(errors, fmt.Sprintf("  %s: %v", path, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to copy %d file(s):\n%s", len(errors), strings.Join(errors, "\n"))
	}

	return nil
}

// PrintAlwaysCopy displays all always-copy paths
func (r *Repo) PrintAlwaysCopy() error {
	if r.Config == nil || len(r.Config.Copy) == 0 {
		fmt.Println("No always-copy paths configured")
		return nil
	}

	for _, path := range r.Config.Copy {
		fmt.Printf("%s\n", path)
	}

	return nil
}

// RunPostCreateCommands runs all configured post-create commands in a worktree
func (r *Repo) RunPostCreateCommands(wt *Worktree) error {
	if r.Config == nil || len(r.Config.Commands) == 0 {
		return nil
	}

	fmt.Printf("Running post-create commands...\n")
	for i, cmdStr := range r.Config.Commands {
		fmt.Printf("  [%d/%d] %s\n", i+1, len(r.Config.Commands), cmdStr)

		cmd := exec.Command("bash", "-c", cmdStr)
		cmd.Dir = wt.Path
		cmd.Stdin = os.Stdin

		// Create pipes for stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdout pipe: %w", err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to create stderr pipe: %w", err)
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("command %d failed to start: %w", i+1, err)
		}

		// Function to prefix and print lines
		printPrefixed := func(reader io.Reader, output *os.File) {
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				fmt.Fprintf(output, "> %s\n", scanner.Text())
			}
		}

		// Process stdout and stderr concurrently
		done := make(chan bool, 2)
		go func() {
			printPrefixed(stdout, os.Stdout)
			done <- true
		}()
		go func() {
			printPrefixed(stderr, os.Stderr)
			done <- true
		}()

		// Wait for both goroutines to finish
		<-done
		<-done

		// Wait for the command to complete
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("command %d failed: %w", i+1, err)
		}
	}

	return nil
}
