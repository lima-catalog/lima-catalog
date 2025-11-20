// Package minhash provides MinHash signature generation for document similarity detection.
//
// MinHash is a locality-sensitive hashing technique that allows efficient estimation
// of Jaccard similarity between sets. This package implements the standard MinHash
// algorithm with configurable parameters.
//
// Basic usage:
//
//	// Create MinHash with default settings (128 hash functions, 5-word shingles)
//	mh := minhash.New()
//
//	// Generate signature from text
//	text := "This is a sample document for similarity detection"
//	signature := mh.Signature(text)
//
//	// Compare two documents
//	sig1 := mh.Signature(text1)
//	sig2 := mh.Signature(text2)
//	similarity := minhash.Similarity(sig1, sig2)  // Jaccard similarity estimate
//
// The package uses k-word shingling to convert text into sets of overlapping sequences,
// then generates a MinHash signature as an array of minimum hash values.
package minhash

import (
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

const (
	// DefaultNumHashes is the default number of hash functions (128)
	// More hashes = more accurate similarity estimation but larger signatures
	DefaultNumHashes = 128

	// DefaultShingleSize is the default k-gram size (5 words)
	// Larger k = more specific matching, smaller k = more general
	DefaultShingleSize = 5

	// MaxHashValue is the maximum value for a 32-bit hash
	MaxHashValue = math.MaxUint32
)

// MinHash represents a MinHash signature generator
type MinHash struct {
	numHashes    int  // Number of hash functions to use
	shingleSize  int  // Size of word shingles (k-grams)
	seeds        []uint32  // Seeds for hash functions
}

// Option configures a MinHash instance
type Option func(*MinHash)

// WithNumHashes sets the number of hash functions
func WithNumHashes(n int) Option {
	return func(mh *MinHash) {
		mh.numHashes = n
	}
}

// WithShingleSize sets the k-gram shingle size
func WithShingleSize(k int) Option {
	return func(mh *MinHash) {
		mh.shingleSize = k
	}
}

// New creates a new MinHash signature generator with the given options.
//
// Default configuration:
//   - numHashes: 128 (provides good accuracy/storage tradeoff)
//   - shingleSize: 5 words (good for document similarity)
//
// Example:
//
//	mh := minhash.New(
//	    minhash.WithNumHashes(256),  // More accurate
//	    minhash.WithShingleSize(3),   // Smaller shingles
//	)
func New(opts ...Option) *MinHash {
	mh := &MinHash{
		numHashes:   DefaultNumHashes,
		shingleSize: DefaultShingleSize,
	}

	for _, opt := range opts {
		opt(mh)
	}

	// Generate seeds for hash functions
	// We use different seeds to simulate different hash functions
	mh.seeds = make([]uint32, mh.numHashes)
	for i := 0; i < mh.numHashes; i++ {
		mh.seeds[i] = uint32(i + 1)
	}

	return mh
}

// Signature generates a MinHash signature from text.
//
// The signature is an array of uint32 values representing the minimum hash
// for each hash function. The length of the array equals numHashes.
//
// Steps:
//  1. Normalize text (lowercase, trim whitespace)
//  2. Extract word shingles (k-grams)
//  3. For each shingle, compute hash with all hash functions
//  4. Keep minimum hash value for each function
//
// Returns empty signature if text has no shingles.
func (mh *MinHash) Signature(text string) []uint32 {
	// Initialize signature with max values
	sig := make([]uint32, mh.numHashes)
	for i := range sig {
		sig[i] = MaxHashValue
	}

	// Extract shingles
	shingles := mh.extractShingles(text)
	if len(shingles) == 0 {
		return sig
	}

	// For each shingle, compute all hashes and update minimums
	for _, shingle := range shingles {
		for i, seed := range mh.seeds {
			h := hashWithSeed(shingle, seed)
			if h < sig[i] {
				sig[i] = h
			}
		}
	}

	return sig
}

// extractShingles converts text into a set of k-word shingles.
//
// Process:
//  1. Normalize text (lowercase, Unicode normalization)
//  2. Tokenize into words (split on whitespace/punctuation)
//  3. Create overlapping k-word sequences
//  4. Deduplicate shingles
//
// Example with k=3:
//   "This is a test document" -> ["this is a", "is a test", "a test document"]
func (mh *MinHash) extractShingles(text string) []string {
	// Normalize text
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)

	// Tokenize into words
	words := tokenize(text)
	if len(words) < mh.shingleSize {
		// Not enough words for a shingle, use what we have
		if len(words) == 0 {
			return []string{}
		}
		return []string{strings.Join(words, " ")}
	}

	// Extract shingles
	shingleSet := make(map[string]bool)
	for i := 0; i <= len(words)-mh.shingleSize; i++ {
		shingle := strings.Join(words[i:i+mh.shingleSize], " ")
		shingleSet[shingle] = true
	}

	// Convert set to slice
	shingles := make([]string, 0, len(shingleSet))
	for shingle := range shingleSet {
		shingles = append(shingles, shingle)
	}

	return shingles
}

// tokenize splits text into words, removing punctuation and extra whitespace.
//
// Handles:
//   - Multiple whitespace -> single space
//   - Punctuation removal
//   - Empty tokens filtered out
func tokenize(text string) []string {
	var words []string
	var currentWord strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			currentWord.WriteRune(r)
		} else if unicode.IsSpace(r) {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
		// Skip punctuation
	}

	// Add last word if exists
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// hashWithSeed computes a hash of the string with a given seed.
//
// Uses FNV-1a hash (fast, good distribution) combined with seed
// to simulate different hash functions.
func hashWithSeed(s string, seed uint32) uint32 {
	h := fnv.New32a()

	// Mix in the seed first
	seedBytes := []byte{
		byte(seed >> 24),
		byte(seed >> 16),
		byte(seed >> 8),
		byte(seed),
	}
	h.Write(seedBytes)

	// Hash the string
	h.Write([]byte(s))

	return h.Sum32()
}

// Similarity estimates the Jaccard similarity between two MinHash signatures.
//
// Jaccard similarity = |A ∩ B| / |A ∪ B|
//
// For MinHash signatures, we estimate this as:
//   similarity ≈ (number of matching hash values) / (total number of hashes)
//
// Returns a value between 0.0 (no similarity) and 1.0 (identical).
//
// The signatures must have the same length (same numHashes parameter).
// If lengths differ, returns 0.0.
func Similarity(sig1, sig2 []uint32) float64 {
	if len(sig1) != len(sig2) {
		return 0.0
	}

	if len(sig1) == 0 {
		return 0.0
	}

	matches := 0
	for i := range sig1 {
		if sig1[i] == sig2[i] {
			matches++
		}
	}

	return float64(matches) / float64(len(sig1))
}

// SignatureSize returns the number of hash functions (length of signature array)
func (mh *MinHash) SignatureSize() int {
	return mh.numHashes
}

// ShingleSize returns the k-gram size used for shingling
func (mh *MinHash) ShingleSize() int {
	return mh.shingleSize
}
