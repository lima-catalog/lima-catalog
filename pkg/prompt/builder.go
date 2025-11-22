package prompt

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	ghclient "github.com/lima-catalog/lima-catalog/pkg/github"
	"github.com/lima-catalog/lima-catalog/pkg/types"
	"github.com/lima-catalog/lima-catalog/pkg/validation"
	"github.com/google/go-github/v57/github"
)

//go:embed default_template.tmpl
var defaultTemplate string

// Builder handles LLM prompt generation for Lima templates
type Builder struct {
	githubClient *ghclient.Client
	config       *PromptConfig
	ctx          context.Context
}

// NewBuilder creates a new prompt builder
func NewBuilder(ctx context.Context, githubToken string, config *PromptConfig) (*Builder, error) {
	// Validate GitHub token
	if err := validation.ValidateGitHubToken(githubToken); err != nil {
		return nil, fmt.Errorf("invalid GitHub token: %w", err)
	}

	// Use default config if not provided
	if config == nil {
		config = DefaultPromptConfig()
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Builder{
		githubClient: ghclient.NewClient(ctx, githubToken),
		config:       config,
		ctx:          ctx,
	}, nil
}

// BuildPrompt generates an LLM prompt for a given template
func (b *Builder) BuildPrompt(owner, repo, templatePath string) (string, error) {
	// Validate inputs
	if err := validation.ValidateRepoIdentifier(owner, repo); err != nil {
		return "", fmt.Errorf("invalid repository: %w", err)
	}

	// Sanitize template path
	sanitizedPath, err := validation.SanitizePath(templatePath)
	if err != nil {
		return "", fmt.Errorf("invalid template path: %w", err)
	}

	// Build the context
	ctx, err := b.GatherContext(owner, repo, sanitizedPath)
	if err != nil {
		return "", fmt.Errorf("failed to gather context: %w", err)
	}

	// Format the prompt
	prompt := b.FormatPrompt(ctx)
	return prompt, nil
}

// GatherContext collects all context needed for prompt generation
func (b *Builder) GatherContext(owner, repo, templatePath string) (*TemplateContext, error) {
	ctx := &TemplateContext{}

	// Fetch repository information
	repoInfo, err := b.githubClient.GetRepository(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}
	ctx.Repository = convertGitHubRepo(repoInfo, owner, repo)

	// Fetch organization information
	orgInfo, err := b.githubClient.GetUser(owner)
	if err != nil {
		// Non-fatal - continue without org info
		fmt.Fprintf(os.Stderr, "Warning: failed to fetch org info: %v\n", err)
	} else {
		ctx.Organization = convertGitHubUser(orgInfo)
	}

	// Fetch template content
	templateContent, err := b.fetchTemplateContent(owner, repo, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template content: %w", err)
	}
	ctx.TemplateContent = templateContent

	// Extract comments from YAML
	if b.config.IncludeComments {
		ctx.Comments = extractYAMLComments(templateContent)
	}

	// Create minimal template info for context
	ctx.Template = &types.Template{
		ID:   fmt.Sprintf("%s/%s/%s", owner, repo, templatePath),
		Repo: fmt.Sprintf("%s/%s", owner, repo),
		Path: templatePath,
	}

	// Fetch README if enabled
	if b.config.IncludeReadme {
		readme, err := b.fetchReadme(owner, repo)
		if err != nil {
			// Non-fatal - continue without README
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch README: %v\n", err)
		} else {
			ctx.ReadmeContent = readme
			if b.config.MaxReadmeLength > 0 && len(readme) > b.config.MaxReadmeLength {
				ctx.ReadmeContent = readme[:b.config.MaxReadmeLength] + "\n... [README truncated]"
			}
		}
	}

	// Find references to template file if enabled
	if b.config.IncludeReferences {
		refs, err := b.findTemplateReferences(owner, repo, templatePath)
		if err != nil {
			// Non-fatal - continue without references
			fmt.Fprintf(os.Stderr, "Warning: failed to find template references: %v\n", err)
		} else {
			ctx.References = refs
		}
	}

	return ctx, nil
}

// FormatPrompt formats the context into an LLM prompt using templates
func (b *Builder) FormatPrompt(ctx *TemplateContext) string {
	// Load template (custom or default)
	var templateContent string
	var err error

	if b.config.TemplatePath != "" {
		// Load custom template from file
		templateBytes, err := os.ReadFile(b.config.TemplatePath)
		if err != nil {
			// Fall back to default template if custom template fails to load
			fmt.Fprintf(os.Stderr, "Warning: failed to load custom template from %s: %v (using default)\n", b.config.TemplatePath, err)
			templateContent = defaultTemplate
		} else {
			templateContent = string(templateBytes)
		}
	} else {
		// Use embedded default template
		templateContent = defaultTemplate
	}

	// Create template with helper functions
	tmpl, err := template.New("prompt").Funcs(template.FuncMap{
		"join": strings.Join,
		"sub": func(a, b int) int {
			return a - b
		},
	}).Parse(templateContent)
	if err != nil {
		// Fall back to a minimal error message if template parsing fails
		return fmt.Sprintf("ERROR: Failed to parse prompt template: %v\n\nPlease check your template syntax or remove the custom template to use the default.", err)
	}

	// Prepare template data
	data := &TemplateData{
		TemplateContext: ctx,
		Config:          b.config,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("ERROR: Failed to execute prompt template: %v\n\nPlease check your template syntax or remove the custom template to use the default.", err)
	}

	return buf.String()
}


// fetchTemplateContent downloads the template YAML content
func (b *Builder) fetchTemplateContent(owner, repo, path string) (string, error) {
	content, err := b.githubClient.GetRepositoryContent(owner, repo, path)
	if err != nil {
		return "", err
	}

	// Decode base64 content
	decodedContent, err := content.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to decode content: %w", err)
	}

	return decodedContent, nil
}

// fetchReadme fetches the README file from the repository
func (b *Builder) fetchReadme(owner, repo string) (string, error) {
	// Try common README file names
	readmeNames := []string{"README.md", "README.MD", "Readme.md", "readme.md", "README", "README.txt"}

	for _, name := range readmeNames {
		content, err := b.githubClient.GetRepositoryContent(owner, repo, name)
		if err == nil {
			decoded, err := content.GetContent()
			if err == nil {
				return decoded, nil
			}
		}
	}

	return "", fmt.Errorf("no README found")
}

// findTemplateReferences searches for references to the template file in the repository
func (b *Builder) findTemplateReferences(owner, repo, templatePath string) ([]TemplateReference, error) {
	// Create a temporary directory for shallow clone
	tmpDir, err := os.MkdirTemp("", "lima-template-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Shallow clone the repository
	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	cmd := exec.CommandContext(b.ctx, "git", "clone", "--depth", "1", cloneURL, tmpDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to clone repo: %w (output: %s)", err, string(output))
	}

	// Extract just the filename from the path
	templateFilename := filepath.Base(templatePath)

	// Use grep to find references
	cmd = exec.CommandContext(b.ctx, "grep",
		"-r",                                    // recursive
		"-n",                                    // line numbers
		"-B", fmt.Sprintf("%d", b.config.ContextLines), // before context
		"-A", fmt.Sprintf("%d", b.config.ContextLines), // after context
		"-F",                 // fixed string (not regex)
		templateFilename,     // search pattern
		tmpDir,               // search path
		"--exclude-dir=.git", // exclude git directory
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// grep returns exit code 1 if no matches found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []TemplateReference{}, nil
		}
		return nil, fmt.Errorf("grep failed: %w", err)
	}

	// Parse grep output
	refs := parseGrepOutput(string(output), tmpDir)
	return refs, nil
}

// parseGrepOutput parses grep output with context into TemplateReference structs
func parseGrepOutput(output, basePath string) []TemplateReference {
	var refs []TemplateReference
	var currentRef *TemplateReference
	var beforeContext []string

	// Match line with line number: "path/file.txt:42:content"
	linePattern := regexp.MustCompile(`^([^:]+):(\d+):(.*)$`)
	// Match separator line: "--"
	separatorPattern := regexp.MustCompile(`^--$`)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// Check for separator (indicates new match group)
		if separatorPattern.MatchString(line) {
			// Save current reference if exists
			if currentRef != nil {
				refs = append(refs, *currentRef)
				currentRef = nil
				beforeContext = nil
			}
			continue
		}

		// Parse line with line number
		matches := linePattern.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}

		filePath := strings.TrimPrefix(matches[1], basePath+"/")
		_ = matches[2] // lineNum - not used in this simplified parsing
		content := matches[3]

		// Check if this is the match line (contains ">>>") or context
		if currentRef == nil {
			// We're before the match - accumulate context
			beforeContext = append(beforeContext, content)
		} else {
			// We're after the match - add to after context
			currentRef.AfterContext = append(currentRef.AfterContext, content)
		}

		// If we see a new file, it's a new match
		if currentRef != nil && currentRef.FilePath != filePath {
			refs = append(refs, *currentRef)
			currentRef = nil
			beforeContext = []string{content}
		}
	}

	// Save last reference
	if currentRef != nil {
		refs = append(refs, *currentRef)
	}

	// Now parse again to identify the actual match lines
	// The match line is the middle line in each group
	return identifyMatchLines(refs, output)
}

// identifyMatchLines identifies which lines are the actual matches vs context
func identifyMatchLines(refs []TemplateReference, output string) []TemplateReference {
	// For simplicity, we'll re-parse the output more carefully
	var result []TemplateReference

	lines := strings.Split(output, "\n")
	linePattern := regexp.MustCompile(`^([^:]+):(\d+):(.*)$`)

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Skip separators
		if line == "--" {
			i++
			continue
		}

		// Parse this match group
		matches := linePattern.FindStringSubmatch(line)
		if len(matches) != 4 {
			i++
			continue
		}

		// Start of a new match group
		var group []string
		var lineNumbers []string
		filePath := matches[1]

		// Collect all lines in this group until separator or end
		for i < len(lines) && lines[i] != "--" {
			lineMatches := linePattern.FindStringSubmatch(lines[i])
			if len(lineMatches) == 4 {
				group = append(group, lineMatches[3])
				lineNumbers = append(lineNumbers, lineMatches[2])
			}
			i++
		}

		if len(group) == 0 {
			continue
		}

		// The middle line is typically the match line
		// But we don't know for sure from grep output alone
		// So we'll just mark the first line as the match for now
		ref := TemplateReference{
			FilePath:      strings.TrimPrefix(filePath, "/tmp/lima-template-"),
			LineNumber:    0, // Will be parsed from lineNumbers[0]
			BeforeContext: []string{},
			MatchLine:     group[0],
			AfterContext:  group[1:],
		}

		// Parse line number
		_, _ = fmt.Sscanf(lineNumbers[0], "%d", &ref.LineNumber)

		result = append(result, ref)
	}

	return result
}

// extractYAMLComments extracts comments from YAML content
func extractYAMLComments(content string) []string {
	var comments []string
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if comment != "" && !isYAMLSectionDivider(comment) {
				comments = append(comments, comment)
			}
		}
	}

	return comments
}

// isYAMLSectionDivider checks if a comment is just a section divider
func isYAMLSectionDivider(comment string) bool {
	// Skip comments that are just dashes, equals, or other dividers
	trimmed := strings.Trim(comment, " -=")
	return trimmed == ""
}

// Helper functions to convert GitHub API types to our types

func convertGitHubRepo(ghRepo *github.Repository, owner, repo string) *types.Repository {
	r := &types.Repository{
		ID:            fmt.Sprintf("%s/%s", owner, repo),
		Owner:         owner,
		Name:          repo,
		DefaultBranch: ghRepo.GetDefaultBranch(),
		Stars:         ghRepo.GetStargazersCount(),
		Forks:         ghRepo.GetForksCount(),
		Watchers:      ghRepo.GetWatchersCount(),
		Language:      ghRepo.GetLanguage(),
		Description:   ghRepo.GetDescription(),
		Homepage:      ghRepo.GetHomepage(),
		LastFetched:   time.Now(),
	}

	if ghRepo.License != nil {
		r.License = ghRepo.License.GetSPDXID()
	}

	if ghRepo.CreatedAt != nil {
		r.CreatedAt = ghRepo.CreatedAt.Time
	}
	if ghRepo.UpdatedAt != nil {
		r.UpdatedAt = ghRepo.UpdatedAt.Time
	}
	if ghRepo.PushedAt != nil {
		r.PushedAt = ghRepo.PushedAt.Time
	}

	// Topics
	r.Topics = ghRepo.Topics

	return r
}

func convertGitHubUser(ghUser *github.User) *types.Organization {
	return &types.Organization{
		ID:          ghUser.GetLogin(),
		Login:       ghUser.GetLogin(),
		Type:        ghUser.GetType(),
		Name:        ghUser.GetName(),
		Description: ghUser.GetBio(),
		Location:    ghUser.GetLocation(),
		Blog:        ghUser.GetBlog(),
		Email:       ghUser.GetEmail(),
		LastFetched: time.Now(),
	}
}
