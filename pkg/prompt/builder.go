package prompt

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	ghclient "github.com/lima-catalog/lima-catalog/pkg/github"
	"github.com/lima-catalog/lima-catalog/pkg/types"
	"github.com/google/go-github/v57/github"
)

// Builder handles LLM prompt generation for Lima templates
type Builder struct {
	githubClient *ghclient.Client
	config       *PromptConfig
	ctx          context.Context
}

// NewBuilder creates a new prompt builder
func NewBuilder(ctx context.Context, githubToken string, config *PromptConfig) *Builder {
	if config == nil {
		config = DefaultPromptConfig()
	}

	return &Builder{
		githubClient: ghclient.NewClient(ctx, githubToken),
		config:       config,
		ctx:          ctx,
	}
}

// BuildPrompt generates an LLM prompt for a given template
func (b *Builder) BuildPrompt(owner, repo, templatePath string) (string, error) {
	// Build the context
	ctx, err := b.GatherContext(owner, repo, templatePath)
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

// FormatPrompt formats the context into an LLM prompt
func (b *Builder) FormatPrompt(ctx *TemplateContext) string {
	var buf bytes.Buffer

	writeHeader(&buf)
	writeTemplateInfo(&buf, ctx)
	writeRepositoryInfo(&buf, ctx)
	writeOrganizationInfo(&buf, ctx)
	writeReadmeContent(&buf, ctx)
	writeTemplateContent(&buf, ctx)
	writeComments(&buf, ctx)
	writeReferences(&buf, ctx, b.config)
	writeInstructions(&buf)

	return buf.String()
}

// writeHeader writes the initial prompt header
func writeHeader(buf *bytes.Buffer) {
	buf.WriteString("# Lima Template Analysis Request\n\n")
	buf.WriteString("Please analyze the following Lima VM template and provide:\n")
	buf.WriteString("1. A short description (max 100 characters) - Brief one-liner suitable for search results\n")
	buf.WriteString("2. A long description (max 500 characters) - Detailed explanation of what this template provides\n")
	buf.WriteString("3. A list of relevant keywords (5-10 keywords) - Technologies, use cases, and key features\n\n")
	buf.WriteString("## Context\n\n")
}

// writeTemplateInfo writes template file information
func writeTemplateInfo(buf *bytes.Buffer, ctx *TemplateContext) {
	buf.WriteString("### Template File\n\n")
	buf.WriteString(fmt.Sprintf("- **Repository**: %s\n", ctx.Template.Repo))
	buf.WriteString(fmt.Sprintf("- **Path**: %s\n", ctx.Template.Path))
	buf.WriteString("\n")
}

// writeRepositoryInfo writes repository context
func writeRepositoryInfo(buf *bytes.Buffer, ctx *TemplateContext) {
	if ctx.Repository == nil {
		return
	}

	buf.WriteString("### Repository Information\n\n")
	buf.WriteString(fmt.Sprintf("- **Name**: %s\n", ctx.Repository.Name))
	if ctx.Repository.Description != "" {
		buf.WriteString(fmt.Sprintf("- **Description**: %s\n", ctx.Repository.Description))
	}
	if len(ctx.Repository.Topics) > 0 {
		buf.WriteString(fmt.Sprintf("- **Topics**: %s\n", strings.Join(ctx.Repository.Topics, ", ")))
	}
	if ctx.Repository.Language != "" {
		buf.WriteString(fmt.Sprintf("- **Primary Language**: %s\n", ctx.Repository.Language))
	}
	buf.WriteString(fmt.Sprintf("- **Stars**: %d\n", ctx.Repository.Stars))
	buf.WriteString("\n")

	buf.WriteString("**IMPORTANT CAVEAT**: The repository's purpose may differ from the template's purpose. ")
	buf.WriteString("For example, the repository might be a CI/CD scaffolding project, a documentation repo, ")
	buf.WriteString("or a collection of templates, while the template itself provisions a specific VM environment. ")
	buf.WriteString("Use the repository context as helpful background, but prioritize the template content itself ")
	buf.WriteString("when determining the template's description and keywords. In some cases, the template makes ")
	buf.WriteString("the repository's project available in a VM - use your judgment.\n\n")
}

// writeOrganizationInfo writes organization/owner context
func writeOrganizationInfo(buf *bytes.Buffer, ctx *TemplateContext) {
	if ctx.Organization == nil {
		return
	}

	buf.WriteString("### Organization/Owner Information\n\n")
	buf.WriteString(fmt.Sprintf("- **Login**: %s\n", ctx.Organization.Login))
	buf.WriteString(fmt.Sprintf("- **Type**: %s\n", ctx.Organization.Type))
	if ctx.Organization.Name != "" {
		buf.WriteString(fmt.Sprintf("- **Name**: %s\n", ctx.Organization.Name))
	}
	if ctx.Organization.Description != "" {
		buf.WriteString(fmt.Sprintf("- **Description**: %s\n", ctx.Organization.Description))
	}
	if ctx.Organization.Location != "" {
		buf.WriteString(fmt.Sprintf("- **Location**: %s\n", ctx.Organization.Location))
	}
	buf.WriteString("\n")
}

// writeReadmeContent writes README content if available
func writeReadmeContent(buf *bytes.Buffer, ctx *TemplateContext) {
	if ctx.ReadmeContent == "" {
		return
	}

	buf.WriteString("### README Content\n\n")
	buf.WriteString("```\n")
	buf.WriteString(ctx.ReadmeContent)
	buf.WriteString("\n```\n\n")
}

// writeTemplateContent writes the template YAML content
func writeTemplateContent(buf *bytes.Buffer, ctx *TemplateContext) {
	buf.WriteString("### Template YAML Content\n\n")
	buf.WriteString("```yaml\n")
	buf.WriteString(ctx.TemplateContent)
	buf.WriteString("\n```\n\n")
}

// writeComments writes extracted YAML comments
func writeComments(buf *bytes.Buffer, ctx *TemplateContext) {
	if len(ctx.Comments) == 0 {
		return
	}

	buf.WriteString("### Template Comments\n\n")
	buf.WriteString("Key comments found in the template:\n\n")
	for _, comment := range ctx.Comments {
		buf.WriteString(fmt.Sprintf("- %s\n", comment))
	}
	buf.WriteString("\n")
}

// writeReferences writes template file references
func writeReferences(buf *bytes.Buffer, ctx *TemplateContext, config *PromptConfig) {
	if len(ctx.References) == 0 {
		return
	}

	buf.WriteString("### References to Template in Repository\n\n")
	buf.WriteString("The following files reference this template (showing context):\n\n")

	count := 0
	for _, ref := range ctx.References {
		if config.MaxReferenceFiles > 0 && count >= config.MaxReferenceFiles {
			buf.WriteString(fmt.Sprintf("... and %d more references (truncated)\n\n", len(ctx.References)-count))
			break
		}

		buf.WriteString(fmt.Sprintf("#### %s (line %d)\n\n", ref.FilePath, ref.LineNumber))
		buf.WriteString("```\n")

		// Before context
		for _, line := range ref.BeforeContext {
			buf.WriteString(line)
			buf.WriteString("\n")
		}

		// Match line
		buf.WriteString(">>> ")
		buf.WriteString(ref.MatchLine)
		buf.WriteString("\n")

		// After context
		for _, line := range ref.AfterContext {
			buf.WriteString(line)
			buf.WriteString("\n")
		}

		buf.WriteString("```\n\n")
		count++
	}
}

// writeInstructions writes the analysis instructions
func writeInstructions(buf *bytes.Buffer) {
	buf.WriteString("## Analysis Instructions\n\n")
	buf.WriteString("Based on the above context, please provide:\n\n")
	buf.WriteString("1. **Short Description** (max 100 chars): A concise one-liner that captures the essence of this template\n")
	buf.WriteString("2. **Long Description** (max 500 chars): A detailed explanation of:\n")
	buf.WriteString("   - What this template provides\n")
	buf.WriteString("   - What technologies/tools are included\n")
	buf.WriteString("   - What use cases it's designed for\n")
	buf.WriteString("   - Any notable features or configurations\n")
	buf.WriteString("3. **Keywords** (5-10 keywords): Relevant tags including:\n")
	buf.WriteString("   - Operating system(s)\n")
	buf.WriteString("   - Technologies and frameworks\n")
	buf.WriteString("   - Use cases (development, testing, security, etc.)\n")
	buf.WriteString("   - Key features (kubernetes, docker, etc.)\n\n")

	buf.WriteString("Please format your response as:\n\n")
	buf.WriteString("```json\n")
	buf.WriteString("{\n")
	buf.WriteString("  \"short_description\": \"...\",\n")
	buf.WriteString("  \"long_description\": \"...\",\n")
	buf.WriteString("  \"keywords\": [\"keyword1\", \"keyword2\", ...]\n")
	buf.WriteString("}\n")
	buf.WriteString("```\n")
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
		fmt.Sscanf(lineNumbers[0], "%d", &ref.LineNumber)

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
		IsFork:        ghRepo.GetFork(),
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
