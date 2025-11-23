package combiner

import (
	"testing"

	"github.com/lima-catalog/lima-catalog/pkg/discovery"
)

func TestCalculateRanks(t *testing.T) {
	// Create test templates with different scores
	templates := []CombinedTemplate{
		{
			ID: "template1",
			NotabilityScoreBreakdown: &discovery.NotabilityScoreBreakdown{
				Message:    100.0,
				Provision:  50.0,
				Parameters: 20.0,
				Total:      170.0,
			},
		},
		{
			ID: "template2",
			NotabilityScoreBreakdown: &discovery.NotabilityScoreBreakdown{
				Message:    50.0,
				Provision:  50.0, // Tie with template1
				Parameters: 40.0,
				Total:      140.0,
			},
		},
		{
			ID: "template3",
			NotabilityScoreBreakdown: &discovery.NotabilityScoreBreakdown{
				Message:    75.0,
				Provision:  30.0,
				Parameters: 20.0, // Tie with template1
				Total:      125.0,
			},
		},
	}

	ranks := calculateRanks(templates)

	// Verify all templates got ranks
	if len(ranks) != 3 {
		t.Errorf("Expected ranks for 3 templates, got %d", len(ranks))
	}

	// Test template1 ranks
	t1Ranks := ranks["template1"]
	if t1Ranks["message"] != 1 {
		t.Errorf("template1 message rank: expected 1, got %d", t1Ranks["message"])
	}
	if t1Ranks["provision"] != 1 { // Tied with template2
		t.Errorf("template1 provision rank: expected 1 (tied), got %d", t1Ranks["provision"])
	}
	if t1Ranks["parameters"] != 2 { // Tied with template3 for 2nd place
		t.Errorf("template1 parameters rank: expected 2 (tied), got %d", t1Ranks["parameters"])
	}
	if t1Ranks["total"] != 1 {
		t.Errorf("template1 total rank: expected 1, got %d", t1Ranks["total"])
	}

	// Test template2 ranks
	t2Ranks := ranks["template2"]
	if t2Ranks["message"] != 3 {
		t.Errorf("template2 message rank: expected 3, got %d", t2Ranks["message"])
	}
	if t2Ranks["provision"] != 1 { // Tied with template1
		t.Errorf("template2 provision rank: expected 1 (tied), got %d", t2Ranks["provision"])
	}
	if t2Ranks["parameters"] != 1 {
		t.Errorf("template2 parameters rank: expected 1, got %d", t2Ranks["parameters"])
	}
	if t2Ranks["total"] != 2 {
		t.Errorf("template2 total rank: expected 2, got %d", t2Ranks["total"])
	}

	// Test template3 ranks
	t3Ranks := ranks["template3"]
	if t3Ranks["message"] != 2 {
		t.Errorf("template3 message rank: expected 2, got %d", t3Ranks["message"])
	}
	if t3Ranks["provision"] != 3 {
		t.Errorf("template3 provision rank: expected 3, got %d", t3Ranks["provision"])
	}
	if t3Ranks["parameters"] != 2 { // Tied with template1
		t.Errorf("template3 parameters rank: expected 2 (tied), got %d", t3Ranks["parameters"])
	}
	if t3Ranks["total"] != 3 {
		t.Errorf("template3 total rank: expected 3, got %d", t3Ranks["total"])
	}
}

func TestCalculateRanksEmpty(t *testing.T) {
	ranks := calculateRanks([]CombinedTemplate{})
	if ranks != nil {
		t.Errorf("Expected nil ranks for empty templates, got %v", ranks)
	}
}

func TestCalculateRanksAllComponentsIncluded(t *testing.T) {
	// Create a template with all score components
	templates := []CombinedTemplate{
		{
			ID: "template1",
			NotabilityScoreBreakdown: &discovery.NotabilityScoreBreakdown{
				Message:            50.0,
				Provision:          10.0,
				Parameters:         20.0,
				EnvVars:            5.0,
				Probes:             3.0,
				ImageName:          30.0,
				Comments:           8.0,
				ValidationWarnings: -50.0,
				Stars:              25.0,
				Total:              101.0,
			},
		},
	}

	ranks := calculateRanks(templates)

	// Verify all components from registry plus total have ranks
	expectedKeys := append(discovery.GetScoreComponentKeys(), "total")
	t1Ranks := ranks["template1"]

	for _, key := range expectedKeys {
		if _, hasKey := t1Ranks[key]; !hasKey {
			t.Errorf("Missing rank for component: %s", key)
		}
		// With only one template, all ranks should be 1
		if t1Ranks[key] != 1 {
			t.Errorf("Component %s rank: expected 1, got %d", key, t1Ranks[key])
		}
	}
}
