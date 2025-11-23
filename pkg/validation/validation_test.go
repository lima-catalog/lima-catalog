// Unit Tests: Input Validation (validation package)
//
// High-level overview of what's being tested:
// - GitHub token format validation (empty, too short, classic ghp_ format, hex format)
// - Rejecting tokens with invalid characters
// - Template ID validation (owner/repo/path structure)
// - File extension validation (.yaml/.yml)
// - Path sanitization and directory traversal prevention
// - Absolute path rejection
// - Context lines validation (0-100 range)
// - Max length validation for configuration parameters
// - Max files count validation (0-1000 range)
// - Repository identifier validation (owner/repo format with allowed characters)
// - Parsing repository IDs (owner/repo splitting)
// - Handling empty and malformed inputs

package validation

import (
	"strings"
	"testing"
)

func TestValidateGitHubToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Too short token",
			token:   "short",
			wantErr: true,
		},
		{
			name:    "Valid classic token format",
			token:   "ghp_1234567890123456789012345678901234567890",
			wantErr: false,
		},
		{
			name:    "Valid hex token (40 chars)",
			token:   "0123456789abcdef0123456789abcdef01234567",
			wantErr: false,
		},
		{
			name:    "Invalid characters",
			token:   "ghp_!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitHubToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGitHubToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTemplateID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "Valid template ID",
			id:      "owner/repo/template.yaml",
			wantErr: false,
		},
		{
			name:    "Valid nested template ID",
			id:      "owner/repo/path/to/template.yaml",
			wantErr: false,
		},
		{
			name:    "Valid .yml extension",
			id:      "owner/repo/template.yml",
			wantErr: false,
		},
		{
			name:    "Empty ID",
			id:      "",
			wantErr: true,
		},
		{
			name:    "Missing repo",
			id:      "owner/template.yaml",
			wantErr: true,
		},
		{
			name:    "Missing path",
			id:      "owner/repo/",
			wantErr: true,
		},
		{
			name:    "Wrong extension",
			id:      "owner/repo/template.txt",
			wantErr: true,
		},
		{
			name:    "Empty owner",
			id:      "/repo/template.yaml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplateID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplateID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		want     string
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name:    "Simple path",
			path:    "data/templates.jsonl",
			want:    "data/templates.jsonl",
			wantErr: false,
		},
		{
			name:    "Path with dots",
			path:    "data/./templates.jsonl",
			want:    "data/templates.jsonl",
			wantErr: false,
		},
		{
			name:    "Directory traversal attempt",
			path:    "../etc/passwd",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Directory traversal in middle",
			path:    "data/../../../etc/passwd",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Absolute path",
			path:    "/etc/passwd",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty path",
			path:    "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateContextLines(t *testing.T) {
	tests := []struct {
		name    string
		lines   int
		wantErr bool
	}{
		{
			name:    "Valid zero",
			lines:   0,
			wantErr: false,
		},
		{
			name:    "Valid positive",
			lines:   5,
			wantErr: false,
		},
		{
			name:    "Valid max",
			lines:   100,
			wantErr: false,
		},
		{
			name:    "Negative",
			lines:   -1,
			wantErr: true,
		},
		{
			name:    "Too large",
			lines:   101,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContextLines(tt.lines)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContextLines() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMaxLength(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		param   string
		wantErr bool
	}{
		{
			name:    "Valid zero",
			length:  0,
			param:   "MaxReadmeLength",
			wantErr: false,
		},
		{
			name:    "Valid positive",
			length:  5000,
			param:   "MaxReadmeLength",
			wantErr: false,
		},
		{
			name:    "Negative",
			length:  -1,
			param:   "MaxReadmeLength",
			wantErr: true,
		},
		{
			name:    "Too large",
			length:  20 * 1024 * 1024,
			param:   "MaxReadmeLength",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxLength(tt.length, tt.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMaxLength() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.param) {
				t.Errorf("Error should mention parameter name %s, got: %v", tt.param, err)
			}
		})
	}
}

func TestValidateMaxFiles(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		wantErr bool
	}{
		{
			name:    "Valid zero",
			count:   0,
			wantErr: false,
		},
		{
			name:    "Valid positive",
			count:   10,
			wantErr: false,
		},
		{
			name:    "Valid max",
			count:   1000,
			wantErr: false,
		},
		{
			name:    "Negative",
			count:   -1,
			wantErr: true,
		},
		{
			name:    "Too large",
			count:   1001,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxFiles(tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMaxFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRepoIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		owner   string
		repo    string
		wantErr bool
	}{
		{
			name:    "Valid simple names",
			owner:   "user",
			repo:    "repo",
			wantErr: false,
		},
		{
			name:    "Valid with hyphens",
			owner:   "my-org",
			repo:    "my-repo",
			wantErr: false,
		},
		{
			name:    "Valid with underscores",
			owner:   "my_org",
			repo:    "my_repo",
			wantErr: false,
		},
		{
			name:    "Valid with dots",
			owner:   "my.org",
			repo:    "my.repo",
			wantErr: false,
		},
		{
			name:    "Empty owner",
			owner:   "",
			repo:    "repo",
			wantErr: true,
		},
		{
			name:    "Empty repo",
			owner:   "owner",
			repo:    "",
			wantErr: true,
		},
		{
			name:    "Invalid characters in owner",
			owner:   "my org",
			repo:    "repo",
			wantErr: true,
		},
		{
			name:    "Invalid characters in repo",
			owner:   "owner",
			repo:    "my repo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepoIdentifier(tt.owner, tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRepoIdentifier() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseRepoID(t *testing.T) {
	tests := []struct {
		name          string
		repoFullName  string
		wantOwner     string
		wantRepo      string
		wantErr       bool
	}{
		{
			name:         "Valid repo ID",
			repoFullName: "owner/repo",
			wantOwner:    "owner",
			wantRepo:     "repo",
			wantErr:      false,
		},
		{
			name:         "Valid with hyphens",
			repoFullName: "my-org/my-repo",
			wantOwner:    "my-org",
			wantRepo:     "my-repo",
			wantErr:      false,
		},
		{
			name:         "Valid with underscores",
			repoFullName: "my_org/my_repo",
			wantOwner:    "my_org",
			wantRepo:     "my_repo",
			wantErr:      false,
		},
		{
			name:         "Empty string",
			repoFullName: "",
			wantOwner:    "",
			wantRepo:     "",
			wantErr:      true,
		},
		{
			name:         "Missing slash",
			repoFullName: "ownerrepo",
			wantOwner:    "",
			wantRepo:     "",
			wantErr:      true,
		},
		{
			name:         "Too many slashes",
			repoFullName: "owner/repo/extra",
			wantOwner:    "",
			wantRepo:     "",
			wantErr:      true,
		},
		{
			name:         "Empty owner",
			repoFullName: "/repo",
			wantOwner:    "",
			wantRepo:     "",
			wantErr:      true,
		},
		{
			name:         "Empty repo",
			repoFullName: "owner/",
			wantOwner:    "",
			wantRepo:     "",
			wantErr:      true,
		},
		{
			name:         "Only slash",
			repoFullName: "/",
			wantOwner:    "",
			wantRepo:     "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepoID(tt.repoFullName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepoID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if owner != tt.wantOwner {
				t.Errorf("ParseRepoID() owner = %v, want %v", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("ParseRepoID() repo = %v, want %v", repo, tt.wantRepo)
			}
		})
	}
}
