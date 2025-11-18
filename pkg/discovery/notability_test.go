package discovery

import (
	"testing"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

func TestIdentifyUnusualImages(t *testing.T) {
	tests := []struct {
		name     string
		images   []string
		expected []string
	}{
		{
			name:     "All official images",
			images:   []string{"ubuntu", "alpine", "debian"},
			expected: []string{},
		},
		{
			name:     "Mixed official and unusual",
			images:   []string{"ubuntu", "nixos", "alpine", "gentoo"},
			expected: []string{"nixos", "gentoo"},
		},
		{
			name:     "All unusual images",
			images:   []string{"custom", "special"},
			expected: []string{"custom", "special"},
		},
		{
			name:     "Empty list",
			images:   []string{},
			expected: []string{},
		},
		{
			name:     "Case insensitive matching",
			images:   []string{"Ubuntu", "ALPINE", "CustomOS"},
			expected: []string{"CustomOS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IdentifyUnusualImages(tt.images)
			if len(result) != len(tt.expected) {
				t.Errorf("IdentifyUnusualImages() got %d unusual images, want %d", len(result), len(tt.expected))
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
				MessageLength: 100,
			},
			repoStars: 0,
			minScore:  100.0,
			maxScore:  100.0,
		},
		{
			name: "Template with provision scripts",
			metrics: &types.NotabilityMetrics{
				ProvisionCount:      2,
				ProvisionTotalLines: 50,
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
				ProbeCount:      2,
				ProbeTotalLines: 20,
			},
			repoStars: 0,
			minScore:  12.0, // 2*5 + 20/10 = 12
			maxScore:  12.0,
		},
		{
			name: "Template with unusual images",
			metrics: &types.NotabilityMetrics{
				UnusualImages: []string{"nixos", "gentoo"},
			},
			repoStars: 0,
			minScore:  60.0, // 2*30 = 60
			maxScore:  60.0,
		},
		{
			name: "Template with stars (capped at 50)",
			metrics: &types.NotabilityMetrics{
				MessageLength: 50,
			},
			repoStars: 1000, // Should cap at 50 points
			minScore:  150.0, // 100 (message) + 50 (capped stars)
			maxScore:  150.0,
		},
		{
			name: "Complex template (everything)",
			metrics: &types.NotabilityMetrics{
				MessageLength:       100,
				ProvisionCount:      3,
				ProvisionTotalLines: 100,
				ProbeCount:          2,
				ProbeTotalLines:     20,
				ParamCount:          4,
				EnvCount:            6,
				UnusualImages:       []string{"nixos"},
			},
			repoStars: 500,
			minScore:  100 + 30 + 10 + 10 + 2 + 80 + 60 + 30 + 50, // Total = 372
			maxScore:  372.0,
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
	info := &TemplateInfo{
		Images:              []string{"ubuntu", "nixos", "alpine"},
		MessageLength:       150,
		ProvisionCount:      3,
		ProvisionTotalLines: 75,
		ProbeCount:          1,
		ProbeTotalLines:     10,
		ParamCount:          5,
		EnvCount:            3,
	}

	metrics := PopulateNotabilityMetrics(info)

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

	// Should identify nixos as unusual
	if len(metrics.UnusualImages) != 1 {
		t.Errorf("UnusualImages length = %d, want 1", len(metrics.UnusualImages))
	}
	if len(metrics.UnusualImages) > 0 && metrics.UnusualImages[0] != "nixos" {
		t.Errorf("UnusualImages[0] = %q, want %q", metrics.UnusualImages[0], "nixos")
	}
}
