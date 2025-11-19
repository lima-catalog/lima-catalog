package discovery

import (
	"fmt"
	"testing"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

func TestIdentifyUnusualImages(t *testing.T) {
	// Create a test official domains map (top-level domains from official Lima images)
	officialDomains := map[string]bool{
		"ubuntu.com":        true, // cloud-images.ubuntu.com -> ubuntu.com
		"alpinelinux.org":   true, // dl-cdn.alpinelinux.org -> alpinelinux.org
		"debian.org":        true, // cloud.debian.org -> debian.org
		"fedoraproject.org": true, // download.fedoraproject.org -> fedoraproject.org
		"pkgbuild.com":      true, // geo.mirror.pkgbuild.com -> pkgbuild.com (Arch Linux)
	}

	tests := []struct {
		name     string
		images   []string
		expected []string
	}{
		{
			name: "All official image domains",
			images: []string{
				"https://cloud-images.ubuntu.com/releases/22.04/ubuntu-22.04-server-cloudimg-amd64.img",
				"https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/x86_64/alpine-virt-3.18.0-x86_64.iso",
			},
			expected: []string{},
		},
		{
			name: "Mixed official and unusual domains",
			images: []string{
				"https://cloud-images.ubuntu.com/releases/22.04/ubuntu.img",
				"https://nixos.org/channels/nixos-23.05/nixos.iso",
				"https://dl-cdn.alpinelinux.org/alpine/v3.18/alpine.iso",
				"https://example.com/custom.img",
			},
			expected: []string{"nixos.org", "example.com"},
		},
		{
			name: "All unusual domains",
			images: []string{
				"https://custom.example.com/image.img",
				"https://special.domain.org/system.qcow2",
			},
			expected: []string{"example.com", "domain.org"}, // Now using top-level domains
		},
		{
			name:     "Empty list",
			images:   []string{},
			expected: []string{},
		},
		{
			name: "Deduplicates same domain from multiple images",
			images: []string{
				"https://nixos.org/channels/23.05/nixos.iso",
				"https://nixos.org/channels/23.11/nixos.iso",
				"https://example.com/v1/image.img",
			},
			expected: []string{"nixos.org", "example.com"},
		},
		{
			name: "Skips template:// references",
			images: []string{
				"template://_images/ubuntu.yaml",
				"https://nixos.org/channels/nixos.iso",
			},
			expected: []string{"nixos.org"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IdentifyUnusualImages(tt.images, officialDomains)
			if len(result) != len(tt.expected) {
				t.Errorf("IdentifyUnusualImages() got %d unusual domains, want %d", len(result), len(tt.expected))
				t.Errorf("Got: %v", result)
				t.Errorf("Want: %v", tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("IdentifyUnusualImages() result[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestCalculateNotabilityScore(t *testing.T) {
	tests := []struct {
		name      string
		metrics   *types.NotabilityMetrics
		repoStars int
		minScore  float64 // Minimum expected score
		maxScore  float64 // Maximum expected score
	}{
		{
			name:      "Nil metrics",
			metrics:   nil,
			repoStars: 0,
			minScore:  0.0,
			maxScore:  0.0,
		},
		{
			name: "Template with message (high priority)",
			metrics: &types.NotabilityMetrics{
				MessageLength:    100,
				MessageLineCount: 5, // 5 lines adds 5 points
			},
			repoStars: 0,
			minScore:  55.0, // 50 (base) + 5 (lines)
			maxScore:  55.0,
		},
		{
			name: "Template with provision scripts",
			metrics: &types.NotabilityMetrics{
				ProvisionCount:       2,
				ProvisionSubstantial: 2, // Both scripts are substantial
				ProvisionTotalLines:  50,
			},
			repoStars: 0,
			minScore:  25.0, // 2*10 + 50/10 = 25
			maxScore:  25.0,
		},
		{
			name: "Template with params and env",
			metrics: &types.NotabilityMetrics{
				ParamCount: 3,
				EnvCount:   5,
			},
			repoStars: 0,
			minScore:  110.0, // 3*20 + 5*10 = 110
			maxScore:  110.0,
		},
		{
			name: "Template with probes",
			metrics: &types.NotabilityMetrics{
				ProbeCount:       2,
				ProbeSubstantial: 2, // Both probes are substantial
				ProbeTotalLines:  20,
			},
			repoStars: 0,
			minScore:  12.0, // 2*5 + 20/10 = 12
			maxScore:  12.0,
		},
		{
			name: "Template with unusual images",
			metrics: &types.NotabilityMetrics{
				UnusualImages: []string{"nixos.org", "gentoo.org"},
			},
			repoStars: 0,
			minScore:  30.0, // 30 points once (not per domain)
			maxScore:  30.0,
		},
		{
			name: "Template with stars (capped at 50)",
			metrics: &types.NotabilityMetrics{
				MessageLength:    50,
				MessageLineCount: 3, // 3 lines adds 3 points
			},
			repoStars: 1000, // Should cap at 50 points
			minScore:  103.0, // 50 (message base) + 3 (lines) + 50 (capped stars)
			maxScore:  103.0,
		},
		{
			name: "Template with comments",
			metrics: &types.NotabilityMetrics{
				CommentLineCount: 25,
			},
			repoStars: 0,
			minScore:  50.0, // 25 comments * 2 points each
			maxScore:  50.0,
		},
		{
			name: "Complex template (everything)",
			metrics: &types.NotabilityMetrics{
				MessageLength:        100,
				MessageLineCount:     5, // 5 lines in message
				ProvisionCount:       3,
				ProvisionSubstantial: 3, // All 3 scripts are substantial
				ProvisionTotalLines:  100,
				ProbeCount:           2,
				ProbeSubstantial:     2, // Both probes are substantial
				ProbeTotalLines:      20,
				ParamCount:           4,
				EnvCount:             6,
				CommentLineCount:     15,
				UnusualImages:        []string{"nixos.org"},
			},
			repoStars: 500,
			minScore:  50 + 5 + 30 + 10 + 10 + 2 + 80 + 60 + 30 + 30 + 50, // Total = 357
			maxScore:  357.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateNotabilityScore(tt.metrics, tt.repoStars)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("CalculateNotabilityScore() = %.2f, want between %.2f and %.2f", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestPopulateNotabilityMetrics(t *testing.T) {
	// Create test official knowledge
	officialKnowledge := &OfficialKnowledge{
		Images: []string{
			"ubuntu.com",
			"alpinelinux.org",
		},
		KnownLines: OfficialKnownLines{
			Comments:  []string{},  // Empty for this test
			Provision: []string{},
			Probes:    []string{},
			Messages:  []string{},
		},
	}

	// Create dummy provision lines (non-empty)
	provisionLines := make([]string, 75)
	for i := range provisionLines {
		provisionLines[i] = fmt.Sprintf("provision line %d", i+1)
	}

	// Create dummy probe lines (non-empty)
	probeLines := make([]string, 10)
	for i := range probeLines {
		probeLines[i] = fmt.Sprintf("probe line %d", i+1)
	}

	info := &TemplateInfo{
		Images: []string{
			"https://cloud-images.ubuntu.com/releases/22.04/ubuntu.img",
			"https://nixos.org/channels/nixos.iso",
			"https://dl-cdn.alpinelinux.org/alpine/v3.18/alpine.iso",
		},
		MessageLength:        150,
		ProvisionCount:       3,
		ProvisionTotalLines:  75,
		ProvisionScriptLines: []int{25, 30, 20}, // 3 scripts with 25, 30, and 20 lines each (all >10)
		ProbeCount:           1,
		ProbeTotalLines:      10,
		ProbeScriptLines:     []int{10}, // 1 probe with 10 lines (not >10, so min 1)
		ParamCount:           5,
		EnvCount:             3,
		CommentLineCount:     20,
		CommentLines:         []string{"This is a test comment", "Another comment"}, // Normalized (no # prefix)
		ProvisionLines:       provisionLines,                                         // Fill with real non-empty lines
		ProbeLines:           probeLines,
		MessageLines:         []string{"Line 1", "Line 2"},
	}

	metrics := PopulateNotabilityMetrics(info, officialKnowledge)

	if metrics.MessageLength != 150 {
		t.Errorf("MessageLength = %d, want 150", metrics.MessageLength)
	}
	if metrics.ProvisionCount != 3 {
		t.Errorf("ProvisionCount = %d, want 3", metrics.ProvisionCount)
	}
	if metrics.ProvisionTotalLines != 75 {
		t.Errorf("ProvisionTotalLines = %d, want 75", metrics.ProvisionTotalLines)
	}
	if metrics.ProbeCount != 1 {
		t.Errorf("ProbeCount = %d, want 1", metrics.ProbeCount)
	}
	if metrics.ProbeTotalLines != 10 {
		t.Errorf("ProbeTotalLines = %d, want 10", metrics.ProbeTotalLines)
	}
	if metrics.ParamCount != 5 {
		t.Errorf("ParamCount = %d, want 5", metrics.ParamCount)
	}
	if metrics.EnvCount != 3 {
		t.Errorf("EnvCount = %d, want 3", metrics.EnvCount)
	}
	if metrics.CommentLineCount != 2 {
		t.Errorf("CommentLineCount = %d, want 2 (unique comments from CommentLines)", metrics.CommentLineCount)
	}
	if metrics.MessageLineCount != 2 {
		t.Errorf("MessageLineCount = %d, want 2 (from MessageLines)", metrics.MessageLineCount)
	}
	if metrics.ProvisionSubstantial != 3 {
		t.Errorf("ProvisionSubstantial = %d, want 3 (all 3 scripts have >10 lines)", metrics.ProvisionSubstantial)
	}
	if metrics.ProbeSubstantial != 1 {
		t.Errorf("ProbeSubstantial = %d, want 1 (minimum 1, probe has exactly 10 lines)", metrics.ProbeSubstantial)
	}

	// Should identify nixos.org as unusual domain
	if len(metrics.UnusualImages) != 1 {
		t.Errorf("UnusualImages length = %d, want 1", len(metrics.UnusualImages))
	}
	if len(metrics.UnusualImages) > 0 && metrics.UnusualImages[0] != "nixos.org" {
		t.Errorf("UnusualImages[0] = %q, want %q", metrics.UnusualImages[0], "nixos.org")
	}
}
