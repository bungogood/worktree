package pkg

import (
	"reflect"
	"testing"
)

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
