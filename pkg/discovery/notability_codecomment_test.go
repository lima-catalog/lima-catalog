package discovery

import (
	"testing"
)

func TestIsCodeComment(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		// Shell variables
		{
			name:     "Shell variable $VAR",
			line:     "export PATH=$PATH:/usr/local/bin",
			expected: true,
		},
		{
			name:     "Shell variable ${VAR}",
			line:     "echo ${HOME}/config",
			expected: true,
		},
		{
			name:     "Command substitution $(cmd)",
			line:     "VERSION=$(cat version.txt)",
			expected: true,
		},
		{
			name:     "Dollar sign in URL",
			line:     "Visit http://example.com?price=$10",
			expected: false, // URLs with $ are allowed
		},

		// Backticks
		{
			name:     "Backtick command substitution",
			line:     "FILES=`ls -l`",
			expected: true,
		},
		{
			name:     "Backticks in markdown code",
			line:     "Use the `lima` command",
			expected: true, // Currently filtered - could be relaxed
		},

		// Pipes
		{
			name:     "Shell pipe",
			line:     "cat file.txt | grep pattern",
			expected: true,
		},
		{
			name:     "Markdown table",
			line:     "| Column 1 | Column 2 | Column 3 |",
			expected: false,
		},
		{
			name:     "Single pipe (redirect)",
			line:     "command > output.txt | tee log.txt",
			expected: true,
		},

		// Command chaining
		{
			name:     "AND operator",
			line:     "make && make install",
			expected: true,
		},
		{
			name:     "OR operator",
			line:     "test -f file || echo missing",
			expected: true,
		},

		// Redirects
		{
			name:     "Output redirect",
			line:     "echo hello > file.txt",
			expected: true,
		},
		{
			name:     "Append redirect",
			line:     "echo world >> file.txt",
			expected: true,
		},
		{
			name:     "Error redirect",
			line:     "command 2>&1",
			expected: true,
		},
		{
			name:     "Input redirect",
			line:     "sort < input.txt",
			expected: true,
		},
		{
			name:     "Greater than in text",
			line:     "Version 2.0 is greater than 1.0",
			expected: false,
		},
		{
			name:     "HTML-like tags",
			line:     "Use <tag> elements",
			expected: false,
		},

		// Shell keywords
		{
			name:     "export command",
			line:     "export PATH=/usr/local/bin",
			expected: true,
		},
		{
			name:     "export with assignment",
			line:     "export VAR=value",
			expected: true,
		},
		{
			name:     "source command",
			line:     "source ~/.bashrc",
			expected: true,
		},
		{
			name:     "cd command",
			line:     "cd /etc/lima",
			expected: true,
		},
		{
			name:     "mkdir command",
			line:     "mkdir -p /tmp/test",
			expected: true,
		},
		{
			name:     "chmod command",
			line:     "chmod +x script.sh",
			expected: true,
		},

		// Variable assignment
		{
			name:     "Variable assignment",
			line:     "LIMA_VERSION=1.0.0",
			expected: true,
		},
		{
			name:     "Assignment with underscore",
			line:     "MY_VAR=value",
			expected: true,
		},
		{
			name:     "URL with equals",
			line:     "Visit https://example.com/download?version=1.0",
			expected: false,
		},
		{
			name:     "Comparison operator",
			line:     "Check if value = expected",
			expected: false, // Has spaces, not assignment
		},

		// File paths
		{
			name:     "Absolute path /etc",
			line:     "/etc/lima/config.yaml",
			expected: true,
		},
		{
			name:     "Absolute path /usr",
			line:     "/usr/local/bin/lima",
			expected: true,
		},
		{
			name:     "Home directory tilde",
			line:     "~/Documents/lima.yaml",
			expected: true,
		},
		{
			name:     "File path in sentence",
			line:     "The configuration is stored in /etc/lima/",
			expected: false, // Documentation that mentions a path, not code
		},

		// Normal documentation comments
		{
			name:     "Plain comment",
			line:     "This is a documentation comment",
			expected: false,
		},
		{
			name:     "Comment with punctuation",
			line:     "Example: lima start default.yaml",
			expected: false,
		},
		{
			name:     "Comment with numbers",
			line:     "Lima version 1.0 was released in 2023",
			expected: false,
		},
		{
			name:     "Empty line",
			line:     "",
			expected: false,
		},
		{
			name:     "URL without special chars",
			line:     "See https://lima-vm.io for more info",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCodeComment(tt.line)
			if result != tt.expected {
				t.Errorf("IsCodeComment(%q) = %v, want %v", tt.line, result, tt.expected)
			}
		})
	}
}

func TestIsCodeComment_EdgeCases(t *testing.T) {
	// Test multiple indicators in one line
	t.Run("Multiple indicators", func(t *testing.T) {
		line := "export PATH=$PATH:/usr/local/bin && source ~/.bashrc"
		if !IsCodeComment(line) {
			t.Errorf("IsCodeComment should detect line with multiple code indicators: %q", line)
		}
	})

	// Test that normal sentences with special words don't get filtered
	t.Run("Normal sentence with 'and'", func(t *testing.T) {
		line := "Lima supports both ARM and x86 architectures"
		if IsCodeComment(line) {
			t.Errorf("IsCodeComment should not filter normal sentence: %q", line)
		}
	})

	// Test markdown code reference
	t.Run("Markdown inline code", func(t *testing.T) {
		line := "Run the `lima` command to start"
		// Currently this is filtered due to backticks
		// Could be relaxed if we want to allow markdown
		result := IsCodeComment(line)
		if !result {
			t.Logf("Note: Markdown inline code is currently allowed: %q", line)
		}
	})
}
