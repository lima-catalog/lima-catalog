package minhash

import (
	"fmt"
	"hash/fnv"
)

// LSH implements Locality-Sensitive Hashing for efficient similarity search.
//
// LSH enables sub-linear similarity search by grouping similar signatures into buckets.
// Instead of comparing every pair of signatures (O(n²)), we only compare signatures
// that share at least one bucket (typically O(n) with tuned parameters).
//
// How it works:
//  1. Divide signature into b bands of r rows each (b × r = signature length)
//  2. Hash each band to a bucket
//  3. Signatures in the same bucket are candidates for similarity checking
//
// Example with 128-hash signature:
//   - 32 bands × 4 rows = 128 hashes
//   - Each band is 4 consecutive hash values
//   - Two signatures with 50% similarity have ~99% chance of sharing a bucket
//
// This gives an S-curve threshold function that detects similarities above ~42%.
type LSH struct {
	numBands   int              // Number of bands (b)
	rowsPerBand int             // Rows per band (r)
	buckets    map[uint64][]string // Band hash -> list of template IDs
}

const (
	// DefaultNumBands is the default number of LSH bands (32)
	//
	// Selection Guidelines:
	//   - More bands = lower threshold for detection
	//   - Fewer bands = higher threshold for detection
	//   - Must satisfy: numBands × rowsPerBand = signature length
	//
	// For 128-hash signatures:
	//   - 8 bands × 16 rows = 128 (threshold ~88%, very high)
	//   - 16 bands × 8 rows = 128 (threshold ~71%, high)
	//   - 32 bands × 4 rows = 128 (threshold ~42%, recommended)
	//   - 64 bands × 2 rows = 128 (threshold ~12%, very low)
	//
	// Formula for threshold approximation:
	//   threshold ≈ (1/b)^(1/r)
	//
	// Formula for detection probability:
	//   P(detected | similarity s) = 1 - (1 - s^r)^b
	//
	// With b=32, r=4 (default):
	//   - s=0.8 → P≈1.00 (100% detection)
	//   - s=0.5 → P≈0.99 (99% detection)
	//   - s=0.4 → P≈0.42 (42% detection, near threshold)
	//   - s=0.3 → P≈0.01 (1% detection, filters out)
	DefaultNumBands = 32

	// DefaultRowsPerBand is the default rows per band (4)
	//
	// Must satisfy: DefaultNumBands × DefaultRowsPerBand = DefaultNumHashes
	// 32 × 4 = 128 ✓
	DefaultRowsPerBand = 4
)

// NewLSH creates a new LSH index.
//
// Parameters:
//   - numBands: Number of bands to divide signature into
//   - rowsPerBand: Number of rows (hash values) per band
//
// The product numBands × rowsPerBand must equal the signature length.
// For default 128-hash signatures, use 16 bands × 8 rows.
//
// Returns error if numBands or rowsPerBand is invalid (<= 0).
func NewLSH(numBands, rowsPerBand int) (*LSH, error) {
	if numBands <= 0 {
		return nil, fmt.Errorf("numBands must be positive, got %d", numBands)
	}
	if rowsPerBand <= 0 {
		return nil, fmt.Errorf("rowsPerBand must be positive, got %d", rowsPerBand)
	}

	return &LSH{
		numBands:    numBands,
		rowsPerBand: rowsPerBand,
		buckets:     make(map[uint64][]string),
	}, nil
}

// Add indexes a signature in the LSH structure.
//
// Parameters:
//   - id: Unique identifier for this signature (e.g., template ID)
//   - signature: MinHash signature to index
//
// The signature is divided into bands, each band is hashed, and the ID is
// added to the corresponding bucket for each band.
//
// Returns error if signature length doesn't match numBands × rowsPerBand.
func (lsh *LSH) Add(id string, signature []uint32) error {
	expectedLen := lsh.numBands * lsh.rowsPerBand
	if len(signature) != expectedLen {
		return fmt.Errorf("signature length %d doesn't match expected %d (numBands %d × rowsPerBand %d)",
			len(signature), expectedLen, lsh.numBands, lsh.rowsPerBand)
	}

	// Process each band
	for bandIdx := 0; bandIdx < lsh.numBands; bandIdx++ {
		// Extract band (slice of rows)
		start := bandIdx * lsh.rowsPerBand
		end := start + lsh.rowsPerBand
		band := signature[start:end]

		// Hash the band
		bandHash := hashBand(band, bandIdx)

		// Add to bucket
		lsh.buckets[bandHash] = append(lsh.buckets[bandHash], id)
	}

	return nil
}

// Query finds candidate similar signatures.
//
// Parameters:
//   - signature: MinHash signature to query
//
// Returns:
//   - Map of template IDs to the number of shared buckets
//   - Templates that share more buckets are more likely to be similar
//
// The query process:
//  1. Hash each band of the query signature
//  2. Look up templates in corresponding buckets
//  3. Count how many buckets each template shares with the query
//
// Candidates should be verified with actual similarity calculation
// to confirm they exceed the desired threshold.
func (lsh *LSH) Query(signature []uint32) (map[string]int, error) {
	expectedLen := lsh.numBands * lsh.rowsPerBand
	if len(signature) != expectedLen {
		return nil, fmt.Errorf("signature length %d doesn't match expected %d (numBands %d × rowsPerBand %d)",
			len(signature), expectedLen, lsh.numBands, lsh.rowsPerBand)
	}

	candidates := make(map[string]int)

	// Process each band
	for bandIdx := 0; bandIdx < lsh.numBands; bandIdx++ {
		// Extract band
		start := bandIdx * lsh.rowsPerBand
		end := start + lsh.rowsPerBand
		band := signature[start:end]

		// Hash the band
		bandHash := hashBand(band, bandIdx)

		// Get templates in this bucket
		if ids, ok := lsh.buckets[bandHash]; ok {
			for _, id := range ids {
				candidates[id]++
			}
		}
	}

	return candidates, nil
}

// Size returns the number of unique template IDs indexed.
//
// Note: This counts unique IDs, not total bucket entries.
// Each template appears in multiple buckets (one per band).
func (lsh *LSH) Size() int {
	seen := make(map[string]bool)
	for _, ids := range lsh.buckets {
		for _, id := range ids {
			seen[id] = true
		}
	}
	return len(seen)
}

// NumBuckets returns the total number of non-empty buckets.
func (lsh *LSH) NumBuckets() int {
	return len(lsh.buckets)
}

// Clear removes all indexed signatures from the LSH structure.
func (lsh *LSH) Clear() {
	lsh.buckets = make(map[uint64][]string)
}

// hashBand computes a hash for a band of hash values.
//
// We include the band index in the hash to ensure different bands
// hash to different values even if they contain the same hash values.
func hashBand(band []uint32, bandIdx int) uint64 {
	h := fnv.New64a()

	// Include band index to distinguish bands
	bandBytes := []byte{
		byte(bandIdx >> 24),
		byte(bandIdx >> 16),
		byte(bandIdx >> 8),
		byte(bandIdx),
	}
	h.Write(bandBytes)

	// Hash each value in the band
	for _, val := range band {
		valBytes := []byte{
			byte(val >> 24),
			byte(val >> 16),
			byte(val >> 8),
			byte(val),
		}
		h.Write(valBytes)
	}

	return h.Sum64()
}

// EstimateThreshold estimates the similarity threshold for LSH detection.
//
// The threshold is the similarity value at which there's a 50% probability
// of the two signatures sharing at least one bucket.
//
// Formula: threshold ≈ (1/b)^(1/r)
// where b = numBands, r = rowsPerBand
//
// Example:
//   - 16 bands × 8 rows → threshold ≈ 0.43
//   - 8 bands × 16 rows → threshold ≈ 0.57
//   - 32 bands × 4 rows → threshold ≈ 0.32
func (lsh *LSH) EstimateThreshold() float64 {
	// threshold ≈ (1/b)^(1/r)
	b := float64(lsh.numBands)
	r := float64(lsh.rowsPerBand)
	return pow(1.0/b, 1.0/r)
}

// pow computes x^y using simple iteration (for small positive integer y).
// This avoids importing math package just for power function.
func pow(x, y float64) float64 {
	if y == 0 {
		return 1.0
	}
	if y == 1 {
		return x
	}

	// For fractional y, use approximation: x^y ≈ exp(y * log(x))
	// Simple Newton's method approximation for exp and log
	// This is good enough for threshold estimation

	// log(x) approximation using Taylor series
	logX := 0.0
	term := (x - 1.0) / (x + 1.0)
	logX = 2.0 * term
	termPow := term
	for i := 1; i < 10; i++ {
		termPow *= term * term
		logX += 2.0 * termPow / float64(2*i+1)
	}

	// exp(y * log(x)) approximation
	yLogX := y * logX
	result := 1.0
	term = 1.0
	for i := 1; i < 20; i++ {
		term *= yLogX / float64(i)
		result += term
	}

	return result
}
