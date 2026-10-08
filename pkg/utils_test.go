package pkg

import (
	"reflect"
	"testing"
	"time"
)

func TestParseMaxAge(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{name: "hours", input: "36h", expected: 36 * time.Hour},
		{name: "days", input: "14d", expected: 14 * 24 * time.Hour},
		{name: "weeks", input: "2w", expected: 2 * 7 * 24 * time.Hour},
		{name: "bare number means days", input: "14", expected: 14 * 24 * time.Hour},
		{name: "uppercase unit", input: "2W", expected: 2 * 7 * 24 * time.Hour},
		{name: "empty", input: "", wantErr: true},
		{name: "zero", input: "0d", wantErr: true},
		{name: "unknown unit", input: "3m", wantErr: true},
		{name: "not a number", input: "old", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMaxAge(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseMaxAge(%q) = %v, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseMaxAge(%q) failed: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Fatalf("ParseMaxAge(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestHumanizeAge(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{name: "minutes", input: 45 * time.Minute, expected: "45m"},
		{name: "hours", input: 3 * time.Hour, expected: "3h"},
		{name: "days", input: 5 * 24 * time.Hour, expected: "5d"},
		{name: "exact weeks", input: 14 * 24 * time.Hour, expected: "2w"},
		{name: "weeks and days", input: 17 * 24 * time.Hour, expected: "2w3d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HumanizeAge(tt.input); got != tt.expected {
				t.Fatalf("HumanizeAge(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
func TestGlobFilter(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		candidates []string
		expected   []string
	}{
		{
			name:       "matches simple wildcard",
			pattern:    "feature-*",
			candidates: []string{"feature-1", "feature-abc", "bugfix-1"},
			expected:   []string{"feature-1", "feature-abc"},
		},
		{
			name:       "returns empty on no matches",
			pattern:    "release-*",
			candidates: []string{"feature-1", "bugfix-1"},
			expected:   nil,
		},
		{
			name:       "supports character range",
			pattern:    "feature-[0-9]*",
			candidates: []string{"feature-1", "feature-x", "feature-42"},
			expected:   []string{"feature-1", "feature-42"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GlobFilter(tt.pattern, tt.candidates)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("GlobFilter(%q) = %v, want %v", tt.pattern, got, tt.expected)
			}
		})
	}
}

func TestGlobFilterComplete(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		completions []string
		toComplete  string
		expected    []string
	}{
		{
			name:        "filters already selected values",
			args:        []string{"feature-2"},
			completions: []string{"feature-1", "feature-2", "feature-3"},
			toComplete:  "feature-",
			expected:    []string{"feature-1", "feature-3"},
		},
		{
			name:        "matches slash names",
			args:        []string{},
			completions: []string{"feature/auth", "feature/api", "bugfix/one"},
			toComplete:  "feature/",
			expected:    []string{"feature/auth", "feature/api"},
		},
		{
			name:        "accepts glob-like input",
			args:        []string{},
			completions: []string{"JIRA-123-test", "JIRA-456-other", "nope"},
			toComplete:  "JIRA-*",
			expected:    []string{"JIRA-123-test", "JIRA-456-other"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GlobFilterComplete(tt.args, tt.completions, tt.toComplete)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("GlobFilterComplete(%q) = %v, want %v", tt.toComplete, got, tt.expected)
			}
		})
	}
}
