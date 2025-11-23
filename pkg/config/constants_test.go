// Unit Tests: Configuration Constants (config package)
//
// High-level overview of what's being tested:
// - Default notability weight values and ranges
// - Verifying all weight fields are present (Message, ProvisionBase, ProvisionLine, etc.)
// - Ensuring weights fall within reasonable ranges
// - Validating non-zero values for key weights
// - Consistency of weight values across multiple calls
// - API rate limit constants (Core and Search API)
// - API delay constants (query, pagination, metadata)
// - Concurrency limits for metadata fetching

package config

import (
	"testing"
)

func TestDefaultNotabilityWeights(t *testing.T) {
	weights := DefaultNotabilityWeights()

	// Verify all expected fields are present and have reasonable values
	tests := []struct {
		name     string
		value    float64
		min      float64
		max      float64
	}{
		{"Message", weights.Message, 0.0, 1000.0},
		{"ProvisionBase", weights.ProvisionBase, 0.0, 1000.0},
		{"ProvisionLine", weights.ProvisionLine, 0.0, 100.0},
		{"ProbeBase", weights.ProbeBase, 0.0, 1000.0},
		{"ProbeLine", weights.ProbeLine, 0.0, 100.0},
		{"Parameter", weights.Parameter, 0.0, 1000.0},
		{"EnvVar", weights.EnvVar, 0.0, 1000.0},
		{"UnusualImage", weights.UnusualImage, 0.0, 1000.0},
		{"CommentLine", weights.CommentLine, 0.0, 100.0},
		{"StarsPerPoint", weights.StarsPerPoint, 0.0, 1000.0},
		{"MaxStarsPoints", weights.MaxStarsPoints, 0.0, 1000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.min || tt.value > tt.max {
				t.Errorf("%s weight %v is outside reasonable range [%v, %v]", tt.name, tt.value, tt.min, tt.max)
			}
		})
	}

	// Verify some key weights are non-zero (sanity check)
	if weights.Message == 0 {
		t.Error("expected Message weight to be non-zero")
	}

	if weights.ProvisionBase == 0 {
		t.Error("expected ProvisionBase weight to be non-zero")
	}

	if weights.StarsPerPoint == 0 {
		t.Error("expected StarsPerPoint weight to be non-zero")
	}
}

func TestDefaultNotabilityWeights_Consistency(t *testing.T) {
	// Call multiple times and verify same values
	weights1 := DefaultNotabilityWeights()
	weights2 := DefaultNotabilityWeights()

	if weights1.Message != weights2.Message {
		t.Error("DefaultNotabilityWeights should return consistent values")
	}

	if weights1.ProvisionBase != weights2.ProvisionBase {
		t.Error("DefaultNotabilityWeights should return consistent values")
	}

	if weights1.StarsPerPoint != weights2.StarsPerPoint {
		t.Error("DefaultNotabilityWeights should return consistent values")
	}
}

// Test constant values are as expected
func TestConstants(t *testing.T) {
	t.Run("MinCoreRateLimitRemaining", func(t *testing.T) {
		if MinCoreRateLimitRemaining <= 0 {
			t.Errorf("expected MinCoreRateLimitRemaining > 0, got %d", MinCoreRateLimitRemaining)
		}

		if MinCoreRateLimitRemaining > 5000 {
			t.Errorf("expected MinCoreRateLimitRemaining <= 5000, got %d", MinCoreRateLimitRemaining)
		}
	})

	t.Run("MinSearchRateLimitRemaining", func(t *testing.T) {
		if MinSearchRateLimitRemaining <= 0 {
			t.Errorf("expected MinSearchRateLimitRemaining > 0, got %d", MinSearchRateLimitRemaining)
		}

		if MinSearchRateLimitRemaining > 30 {
			t.Errorf("expected MinSearchRateLimitRemaining <= 30, got %d", MinSearchRateLimitRemaining)
		}
	})

	t.Run("SearchAPIQueryDelay", func(t *testing.T) {
		if SearchAPIQueryDelay <= 0 {
			t.Errorf("expected SearchAPIQueryDelay > 0, got %v", SearchAPIQueryDelay)
		}

		// Shouldn't be more than 1 minute
		if SearchAPIQueryDelay.Seconds() > 60 {
			t.Errorf("expected SearchAPIQueryDelay <= 60s, got %v", SearchAPIQueryDelay)
		}
	})

	t.Run("SearchAPIPaginationDelay", func(t *testing.T) {
		if SearchAPIPaginationDelay <= 0 {
			t.Errorf("expected SearchAPIPaginationDelay > 0, got %v", SearchAPIPaginationDelay)
		}

		if SearchAPIPaginationDelay.Seconds() > 60 {
			t.Errorf("expected SearchAPIPaginationDelay <= 60s, got %v", SearchAPIPaginationDelay)
		}
	})

	t.Run("MetadataAPIDelay", func(t *testing.T) {
		if MetadataAPIDelay <= 0 {
			t.Errorf("expected MetadataAPIDelay > 0, got %v", MetadataAPIDelay)
		}

		if MetadataAPIDelay.Seconds() > 60 {
			t.Errorf("expected MetadataAPIDelay <= 60s, got %v", MetadataAPIDelay)
		}
	})

	t.Run("MaxMetadataConcurrency", func(t *testing.T) {
		if MaxMetadataConcurrency <= 0 {
			t.Errorf("expected MaxMetadataConcurrency > 0, got %d", MaxMetadataConcurrency)
		}

		if MaxMetadataConcurrency > 100 {
			t.Errorf("expected MaxMetadataConcurrency <= 100, got %d", MaxMetadataConcurrency)
		}
	})
}
