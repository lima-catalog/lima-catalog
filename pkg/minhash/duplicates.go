package minhash

import (
	"sort"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

// DuplicateDetector finds similar templates using LSH and MinHash
type DuplicateDetector struct {
	lsh               *LSH
	minHash           *MinHash
	signatures        map[string][]uint32 // Template ID -> MinHash signature
	similarityThreshold float64            // Minimum similarity to report (0.0-1.0)
}

// NewDuplicateDetector creates a new duplicate detector.
//
// Parameters:
//   - minHash: MinHash instance to use for signature generation
//   - similarityThreshold: Minimum Jaccard similarity to report (e.g., 0.5 for 50%)
//
// The LSH configuration is automatically set based on the threshold:
//   - threshold >= 0.7: 16 bands × 8 rows (strict matching)
//   - threshold >= 0.5: 32 bands × 4 rows (balanced, recommended)
//   - threshold < 0.5:  64 bands × 2 rows (loose matching)
//
// Returns error if threshold is invalid (<0 or >1).
func NewDuplicateDetector(minHash *MinHash, similarityThreshold float64) (*DuplicateDetector, error) {
	if similarityThreshold < 0 || similarityThreshold > 1 {
		return nil, ErrInvalidThreshold
	}

	// Choose LSH configuration based on threshold
	numBands, rowsPerBand := chooseLSHConfig(minHash.SignatureSize(), similarityThreshold)

	lsh, err := NewLSH(numBands, rowsPerBand)
	if err != nil {
		return nil, err
	}

	return &DuplicateDetector{
		lsh:                 lsh,
		minHash:             minHash,
		signatures:          make(map[string][]uint32),
		similarityThreshold: similarityThreshold,
	}, nil
}

// chooseLSHConfig selects optimal LSH banding configuration for a given threshold.
//
// Formula: threshold ≈ (1/b)^(1/r)
// We need to find b and r such that b × r = signatureSize and threshold ≈ (1/b)^(1/r)
//
// For 128-hash signatures:
//   - 8×16: threshold ~88% (very strict)
//   - 16×8: threshold ~71% (strict)
//   - 32×4: threshold ~42% (balanced)
//   - 64×2: threshold ~15% (loose)
func chooseLSHConfig(signatureSize int, threshold float64) (numBands, rowsPerBand int) {
	// Default to 32×4 (threshold ~42%)
	numBands, rowsPerBand = 32, 4

	if signatureSize != 128 {
		// For non-standard signature sizes, use balanced configuration
		numBands = signatureSize / 4
		rowsPerBand = 4
		return
	}

	// Select configuration based on threshold
	if threshold >= 0.7 {
		// Strict matching: 16 bands × 8 rows (threshold ~71%)
		numBands, rowsPerBand = 16, 8
	} else if threshold >= 0.5 {
		// Balanced: 32 bands × 4 rows (threshold ~42%)
		numBands, rowsPerBand = 32, 4
	} else {
		// Loose matching: 64 bands × 2 rows (threshold ~15%)
		numBands, rowsPerBand = 64, 2
	}

	return
}

// Add indexes a template's signature for duplicate detection.
//
// If the template doesn't have a MinHash signature, it's skipped.
func (dd *DuplicateDetector) Add(template *types.Template) error {
	if len(template.MinHashSignature) == 0 {
		// No signature, skip
		return nil
	}

	// Store signature
	dd.signatures[template.ID] = template.MinHashSignature

	// Add to LSH index
	return dd.lsh.Add(template.ID, template.MinHashSignature)
}

// FindSimilar finds templates similar to the given template.
//
// Process:
//  1. Query LSH index to get candidates (fast, sub-linear)
//  2. Compute actual Jaccard similarity for each candidate
//  3. Filter by similarity threshold
//  4. Sort by similarity (highest first)
//  5. Classify as exact/near/similar duplicate
//
// Returns a list of similar templates with similarity scores.
// The query template itself is excluded from results.
func (dd *DuplicateDetector) FindSimilar(template *types.Template) ([]types.SimilarTemplate, error) {
	if len(template.MinHashSignature) == 0 {
		// No signature, can't find similarities
		return []types.SimilarTemplate{}, nil
	}

	// Query LSH for candidates
	candidates, err := dd.lsh.Query(template.MinHashSignature)
	if err != nil {
		return nil, err
	}

	// Compute actual similarities for candidates
	var similar []types.SimilarTemplate
	for candidateID, sharedBands := range candidates {
		// Skip self
		if candidateID == template.ID {
			continue
		}

		// Get candidate signature
		candidateSig, ok := dd.signatures[candidateID]
		if !ok {
			// Signature not found (shouldn't happen)
			continue
		}

		// Compute actual Jaccard similarity
		similarity := Similarity(template.MinHashSignature, candidateSig)

		// Filter by threshold
		if similarity < dd.similarityThreshold {
			continue
		}

		// Classify duplicate type
		duplicateType := classifyDuplicate(similarity)

		similar = append(similar, types.SimilarTemplate{
			ID:            candidateID,
			Similarity:    similarity,
			DuplicateType: duplicateType,
			SharedBands:   sharedBands,
		})
	}

	// Sort by similarity (highest first)
	sort.Slice(similar, func(i, j int) bool {
		return similar[i].Similarity > similar[j].Similarity
	})

	return similar, nil
}

// classifyDuplicate classifies the type of duplicate based on similarity score.
//
// Classification:
//   - >0.9: "exact" - Exact or near-exact duplicate (likely fork)
//   - 0.7-0.9: "near" - Near duplicate (minor changes)
//   - 0.5-0.7: "similar" - Similar template (moderate changes)
//   - <0.5: (filtered out by threshold)
func classifyDuplicate(similarity float64) string {
	if similarity > 0.9 {
		return "exact"
	} else if similarity >= 0.7 {
		return "near"
	} else {
		return "similar"
	}
}

// DetectDuplicates processes all templates and populates SimilarTemplates field.
//
// Process:
//  1. Build LSH index from all templates
//  2. For each template, find similar templates
//  3. Populate SimilarTemplates field
//
// Returns the updated templates with SimilarTemplates populated.
// Templates without MinHash signatures are skipped.
func (dd *DuplicateDetector) DetectDuplicates(templates []types.Template) ([]types.Template, error) {
	// Clear any existing data
	dd.lsh.Clear()
	dd.signatures = make(map[string][]uint32)

	// Build LSH index
	for i := range templates {
		if err := dd.Add(&templates[i]); err != nil {
			return nil, err
		}
	}

	// Find similar templates for each template
	for i := range templates {
		similar, err := dd.FindSimilar(&templates[i])
		if err != nil {
			return nil, err
		}
		templates[i].SimilarTemplates = similar
	}

	return templates, nil
}

// Size returns the number of templates indexed
func (dd *DuplicateDetector) Size() int {
	return len(dd.signatures)
}

// Clear removes all indexed templates
func (dd *DuplicateDetector) Clear() {
	dd.lsh.Clear()
	dd.signatures = make(map[string][]uint32)
}

// ErrInvalidThreshold is returned when similarity threshold is invalid
var ErrInvalidThreshold error = &duplicateError{"similarity threshold must be between 0.0 and 1.0"}

type duplicateError struct {
	msg string
}

func (e *duplicateError) Error() string {
	return e.msg
}
