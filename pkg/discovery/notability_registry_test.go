package discovery

import (
	"testing"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

func TestScoreComponentRegistry(t *testing.T) {
	// Test that the registry has the expected number of components
	expectedComponents := 9 // message, provision, parameters, env_vars, probes, image_name, comments, validation_warnings, stars
	if len(ScoreComponentRegistry) != expectedComponents {
		t.Errorf("Expected %d components in registry, got %d", expectedComponents, len(ScoreComponentRegistry))
	}

	// Test that all keys are unique
	seenKeys := make(map[string]bool)
	for _, component := range ScoreComponentRegistry {
		if seenKeys[component.Key] {
			t.Errorf("Duplicate key in registry: %s", component.Key)
		}
		seenKeys[component.Key] = true

		// Verify metadata is not empty
		if component.Key == "" {
			t.Error("Found component with empty key")
		}
		if component.DisplayName == "" {
			t.Errorf("Component %s has empty DisplayName", component.Key)
		}
		if component.Description == "" {
			t.Errorf("Component %s has empty Description", component.Key)
		}
	}
}

func TestGetScoreComponentKeys(t *testing.T) {
	keys := GetScoreComponentKeys()

	// Should match registry length
	if len(keys) != len(ScoreComponentRegistry) {
		t.Errorf("GetScoreComponentKeys returned %d keys, expected %d", len(keys), len(ScoreComponentRegistry))
	}

	// Should match registry order
	for i, key := range keys {
		if key != ScoreComponentRegistry[i].Key {
			t.Errorf("Key at index %d: got %s, expected %s", i, key, ScoreComponentRegistry[i].Key)
		}
	}
}

func TestBreakdownToMap(t *testing.T) {
	breakdown := NotabilityScoreBreakdown{
		Message:            50.0,
		Provision:          10.0,
		Parameters:         20.0,
		EnvVars:            10.0,
		Probes:             5.0,
		ImageName:          30.0,
		Comments:           10.0,
		ValidationWarnings: -50.0,
		Stars:              25.0,
		Total:              110.0,
	}

	m := breakdown.ToMap()

	// Check that all fields are present
	expectedKeys := []string{"message", "provision", "parameters", "env_vars", "probes", "image_name", "comments", "validation_warnings", "stars", "total"}
	for _, key := range expectedKeys {
		if _, exists := m[key]; !exists {
			t.Errorf("ToMap missing key: %s", key)
		}
	}

	// Check specific values
	if m["message"] != 50.0 {
		t.Errorf("Expected message=50.0, got %f", m["message"])
	}
	if m["validation_warnings"] != -50.0 {
		t.Errorf("Expected validation_warnings=-50.0, got %f", m["validation_warnings"])
	}
}

func TestBreakdownValidateCompleteness(t *testing.T) {
	// Create a complete breakdown using the actual calculation function
	metrics := &types.NotabilityMetrics{
		MessageLength:    100,
		MessageLineCount: 5,
		ParamCount:       2,
		EnvCount:         1,
	}

	breakdown := CalculateNotabilityScoreWithBreakdown(metrics, "test-org", "test-repo", 100)

	// Validate completeness - should not error
	if err := breakdown.ValidateCompleteness(); err != nil {
		t.Errorf("ValidateCompleteness failed on valid breakdown: %v", err)
	}

	// Verify all registry components are present in the breakdown map
	breakdownMap := breakdown.ToMap()
	for _, component := range ScoreComponentRegistry {
		if _, exists := breakdownMap[component.Key]; !exists {
			t.Errorf("Breakdown missing registry component: %s", component.Key)
		}
	}
}

func TestBreakdownCalculationIncludesAllComponents(t *testing.T) {
	// Test with various metrics to ensure all components are calculated
	metrics := &types.NotabilityMetrics{
		MessageLength:         100,
		MessageLineCount:      10,
		ProvisionCount:        2,
		ProvisionSubstantial:  1,
		ProvisionTotalLines:   20,
		ProbeCount:            1,
		ProbeSubstantial:      1,
		ProbeTotalLines:       5,
		ParamCount:            3,
		EnvCount:              5,
		CommentLineCount:      10,
		ValidationWarnings:    2,
		ValidationWarningMsgs: []string{"warning1", "warning2"},
		UnusualImages:         []string{"example.com"},
		AllImages:             []string{"https://example.com/image.iso"},
	}

	breakdown := CalculateNotabilityScoreWithBreakdown(metrics, "test-org", "test-repo", 500)

	// All components should be calculated (non-zero or explicitly zero)
	// Check that the total is the sum of all parts
	expectedTotal := breakdown.Message + breakdown.Provision + breakdown.Parameters +
		breakdown.EnvVars + breakdown.Probes + breakdown.ImageName +
		breakdown.Comments + breakdown.ValidationWarnings + breakdown.Stars

	if breakdown.Total != expectedTotal {
		t.Errorf("Total score mismatch: got %f, expected %f", breakdown.Total, expectedTotal)
		t.Logf("Breakdown: Message=%f, Provision=%f, Parameters=%f, EnvVars=%f, Probes=%f, ImageName=%f, Comments=%f, ValidationWarnings=%f, Stars=%f",
			breakdown.Message, breakdown.Provision, breakdown.Parameters, breakdown.EnvVars,
			breakdown.Probes, breakdown.ImageName, breakdown.Comments, breakdown.ValidationWarnings, breakdown.Stars)
	}

	// Verify completeness
	if err := breakdown.ValidateCompleteness(); err != nil {
		t.Errorf("Breakdown validation failed: %v", err)
	}
}
