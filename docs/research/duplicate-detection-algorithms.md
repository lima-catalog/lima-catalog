# Duplicate Detection Research for Lima Template Catalog

## Executive Summary

This document evaluates strategies for detecting duplicate or near-duplicate Lima templates in the catalog. We need a solution that:
- Works incrementally (analyze new templates against existing ones)
- Stores a compact signature per template for later comparison
- Detects both exact duplicates and near-duplicates (minor changes)
- Fits within the existing incremental analysis architecture
- Scales to 1000+ templates

**Recommendation**: **MinHash + LSH** (Strategy 2) provides the best balance of accuracy, performance, and implementation complexity for this use case.

---

## Context

### Current System
- **Templates cataloged** from official and community sources
- **~1.3KB per template** average size (948KB total)
- **Incremental analysis**: Templates only re-analyzed when SHA changes
- **Data storage**: JSON Lines format on data branch
- **Template structure**: YAML files with images, provision scripts, probes, parameters, env vars, etc.

### What Constitutes a Duplicate?
1. **Exact copies** - Same SHA (already handled by GitHub)
2. **Near-duplicates** - Minor changes (whitespace, comments, variable names)
3. **Forks/derivatives** - Templates derived from same base with modifications
4. **Content similarity** - Different repos but very similar implementation

### Requirements
- Store compact signature per template (ideally <1KB)
- Calculate similarity for new templates against all existing ones
- Configurable similarity threshold (e.g., >80% = duplicate, >50% = similar)
- Fast comparison (must scale to 1000+ templates)
- Incremental updates (don't recompute existing signatures)

---

## Strategy 1: SimHash (Locality-Sensitive Hashing)

### Description
SimHash generates a single fixed-size hash fingerprint (typically 64 or 128 bits) where similar documents produce similar hashes. Similarity is measured by Hamming distance between hashes.

### How It Works
1. Tokenize template content into features (words, lines, or shingles)
2. Hash each feature and weight by frequency
3. Combine hashes into a single fingerprint using bit voting
4. Compare fingerprints using Hamming distance

### Pros
- **Very compact storage**: 8-16 bytes per template (64-128 bit hash)
- **Fast comparison**: Hamming distance is O(1) bitwise operation
- **Battle-tested**: Used by Google for web duplicate detection
- **Simple implementation**: Single hash per document
- **Incremental friendly**: Compute once per template
- **Language-agnostic**: Works on any text content

### Cons
- **Limited similarity range**: Only detects high similarity (typically >75%)
- **Fixed threshold**: With 64-bit hash, hamming distance of 3-7 bits limits flexibility
- **No fine-grained control**: Can't easily adjust sensitivity after hashing
- **Binary decision**: Hard to rank by similarity (only distance metric)
- **Sensitive to large changes**: Significant modifications may exceed threshold

### Implementation Complexity
**Low** - Libraries available in Go (github.com/go-dedup/simhash)

### Storage Requirements
- **8 bytes per template** (64-bit hash)
- **~6KB total** for current catalog size
- **Negligible** impact on data size

### Performance
- **Hash generation**: ~1ms per template
- **Comparison**: ~1μs per pair (bitwise XOR + popcount)
- **All-pairs comparison**: ~1ms for current catalog size

### Use Case Fit
**Good for exact and near-exact duplicates** (>75% similarity)
**Not ideal for detecting more distant derivatives** (50-75% similarity)

---

## Strategy 2: MinHash + LSH (Recommended)

### Description
MinHash generates multiple hash values (typically 128-256) per document, representing the document as a set of shingles. LSH groups similar documents into buckets for efficient comparison.

### How It Works
1. Convert template to set of k-shingles (overlapping n-grams)
2. Generate MinHash signature (array of minimum hash values)
3. Use LSH to partition signatures into bands for approximate matching
4. Compare candidates using Jaccard similarity

### Pros
- **Wide similarity range**: Detects similarity as low as 5% up to 100%
- **Flexible thresholds**: Tune bands/rows to control sensitivity
- **Ranked results**: Get similarity scores, not just binary yes/no
- **Semantic understanding**: Shingles capture local structure
- **Standard technique**: Used by AltaVista, Google News
- **Proven for code**: Effective for code clone detection
- **Sub-linear search**: LSH enables O(n) search instead of O(n²)

### Cons
- **Larger storage**: 128-256 integers per template (512-1024 bytes)
- **More complex**: Requires LSH banding configuration
- **Slower generation**: More hashes to compute
- **Memory overhead**: Need to store shingle sets temporarily

### Implementation Complexity
**Medium** - Libraries available (github.com/ekzhu/minhash-lsh for reference, need Go port or implementation)

### Storage Requirements
- **512-1024 bytes per template** (128-256 uint32 hashes)
- **~366KB - 732KB total** for current catalog size
- **Acceptable** - still <1MB for all templates

### Performance
- **Hash generation**: ~10-20ms per template (depends on shingling)
- **LSH bucketing**: ~1ms per template
- **Candidate comparison**: ~10μs per pair (Jaccard on hash sets)
- **All-pairs with LSH**: ~50-100ms for current catalog size (sub-linear)

### Use Case Fit
**Excellent** - Detects everything from near-exact copies to distant derivatives
**Configurable** - Can tune for different similarity thresholds
**Scalable** - LSH provides sub-linear search

---

## Strategy 3: Fuzzy Hashing (ssdeep/TLSH)

### Description
Fuzzy hashing creates context-triggered piecewise hashes that can be compared using edit distance. Designed for malware detection and file similarity.

### How It Works
1. Break file into chunks using rolling hash (context-triggered)
2. Hash each chunk to create signature
3. Compare signatures using edit distance

### Pros
- **Purpose-built for similarity**: Specifically designed for near-duplicate detection
- **Robust to insertions/deletions**: Edit distance handles structural changes
- **Compact storage**: ~100-200 bytes per hash
- **Industry standard**: Used by VirusTotal for malware clustering
- **Handles varied file sizes**: Works for small and large files

### Cons
- **Minimum size requirement**: TLSH requires 50+ bytes, ssdeep ~4KB optimal
- **Templates may be too small**: Many templates are small YAML files
- **Designed for binary files**: Optimized for malware, not structured text
- **Limited to file similarity**: Not semantic or structure-aware
- **Slower comparison**: Edit distance is O(n²) on signature length
- **No LSH acceleration**: Must compare all pairs

### Implementation Complexity
**Medium** - C libraries available, need Go bindings or port

### Storage Requirements
- **100-200 bytes per template**
- **~70-140KB total** for current catalog size

### Performance
- **Hash generation**: ~5-10ms per template
- **Comparison**: ~100μs per pair (edit distance)
- **All-pairs comparison**: ~51 seconds for current catalog size (O(n²))

### Use Case Fit
**Poor** - Templates may be too small
**Not ideal** - Designed for binary files, not YAML structure
**Performance concerns** - O(n²) comparison doesn't scale well

---

## Strategy 4: Shingling + Jaccard Similarity

### Description
Convert templates to sets of k-shingles (overlapping sequences) and compute Jaccard similarity (intersection/union).

### How It Works
1. Extract k-shingles (e.g., 3-word or 5-character sequences)
2. Store set of shingles per template
3. Compute Jaccard similarity: |A ∩ B| / |A ∪ B|

### Pros
- **Intuitive metric**: Jaccard similarity is easy to understand and interpret
- **Flexible k-size**: Can tune shingle size for different granularity
- **Accurate**: Direct similarity measurement, no approximation
- **Character or word level**: Works on both tokens and character sequences
- **Simple implementation**: Just set operations

### Cons
- **Large storage**: Must store all shingles per template (kilobytes)
- **O(n²) comparison**: No acceleration structure
- **Slow for large sets**: Set intersection is expensive
- **Memory intensive**: Need to keep all shingle sets in memory
- **No sub-linear search**: Must compare every pair

### Implementation Complexity
**Low** - Just set operations, easy to implement

### Storage Requirements
- **1-5KB per template** (depends on k and template size)
- **~1-4MB total** for current catalog size
- **Significant** - 5x larger than current catalog

### Performance
- **Shingle extraction**: ~2ms per template
- **Comparison**: ~1ms per pair (set intersection)
- **All-pairs comparison**: ~5 minutes for current catalog size

### Use Case Fit
**Poor** - Storage and performance don't scale
**Only viable with LSH** - Need MinHash+LSH for acceleration

---

## Strategy 5: Semantic Embeddings (Neural)

### Description
Use language models to generate vector embeddings (typically 384-1536 dimensions) and compute cosine similarity.

### How It Works
1. Feed template content to pre-trained language model
2. Extract embedding vector (e.g., sentence-transformers, OpenAI ada-002)
3. Store embedding per template
4. Compute cosine similarity between vectors

### Pros
- **Semantic understanding**: Captures meaning, not just syntax
- **Detects paraphrasing**: Finds similar functionality with different code
- **Ranked results**: Continuous similarity scores
- **Transfer learning**: Pre-trained models understand code context
- **Modern approach**: State-of-the-art for text similarity

### Cons
- **Large storage**: 384-1536 floats = 1.5-6KB per template
- **Requires ML infrastructure**: Need model hosting or API
- **API costs**: OpenAI charges per token
- **Slow generation**: Neural inference takes 50-500ms per template
- **Over-engineered**: Templates are structured, not natural language
- **No structural awareness**: Doesn't understand YAML structure
- **Black box**: Hard to debug why two templates are similar

### Implementation Complexity
**High** - Requires ML infrastructure or API integration

### Storage Requirements
- **1.5-6KB per template** (384-1536 float32s)
- **~1-4.3MB total** for current catalog size

### Performance
- **Embedding generation**: 50-500ms per template (model dependent)
- **Comparison**: ~50μs per pair (dot product)
- **All-pairs comparison**: ~25 seconds for current catalog size

### Use Case Fit
**Overkill** - Templates are structured YAML, not natural language
**Too expensive** - API costs and infrastructure complexity
**Not incremental friendly** - Slow generation makes daily runs impractical

---

## Strategy 6: AST-Based Comparison

### Description
Parse YAML into Abstract Syntax Tree, normalize, and compare tree structures using tree edit distance or structural hashing.

### How It Works
1. Parse YAML to AST
2. Normalize (sort keys, canonicalize values)
3. Compute tree hash or tree edit distance
4. Compare structures

### Pros
- **Structure-aware**: Understands YAML semantics, not just text
- **Ignores formatting**: Whitespace and comment differences ignored
- **Accurate for semantic similarity**: Detects same functionality
- **Handles reordering**: Can detect similar templates with reordered keys
- **Precise**: Can identify exactly what differs (which fields)

### Cons
- **Complex implementation**: Tree edit distance is non-trivial
- **Slow comparison**: Tree edit distance is O(n³) in worst case
- **Large storage**: Must store tree structure or full content
- **YAML-specific**: Doesn't help with cross-format comparison
- **Misses semantic similarity**: Same function implemented differently looks different
- **No standard library**: Need to implement tree comparison

### Implementation Complexity
**High** - Tree algorithms are complex

### Storage Requirements
- **Store normalized YAML** (~1KB per template)
- **Or store tree hash** (~32 bytes, but loses detail)

### Performance
- **Parsing**: ~1ms per template
- **Normalization**: ~1ms per template
- **Tree edit distance**: ~10-100ms per pair (algorithm dependent)
- **All-pairs comparison**: ~10-100 minutes for current catalog size

### Use Case Fit
**Poor** - Too slow for all-pairs comparison
**Could work for exact matches** - Tree hashing is fast
**Better suited for small-scale comparison** - Not scalable to all templates

---

## Strategy 7: Feature Extraction + Hashing

### Description
Extract key features from templates (image URLs, provision script lines, parameters) and hash the feature set.

### How It Works
1. Extract structured features:
   - Image URLs (normalized domains)
   - Provision script lines (normalized)
   - Parameters and env vars
   - Keywords and categories
2. Create feature vector
3. Hash feature vector
4. Compare hashes or feature vectors

### Pros
- **Leverages existing analysis**: Uses data already extracted
- **Domain-specific**: Focuses on Lima-specific features
- **Compact**: Feature vectors are small
- **Fast**: Simple comparison operations
- **Semantic**: Captures what template does, not how it's written
- **Explainable**: Can show which features match

### Cons
- **Limited to extracted features**: Misses other differences
- **Not general-purpose**: Specific to Lima templates
- **Threshold tuning**: Need to weight feature importance
- **Misses novel patterns**: Only detects known feature types
- **Requires feature engineering**: Need to define "important" features

### Implementation Complexity
**Low** - We already extract these features

### Storage Requirements
- **Already stored** in notability metrics
- **No additional storage** needed
- **Could add feature hash** (32-64 bytes) for fast comparison

### Performance
- **Feature extraction**: Already done during analysis
- **Comparison**: ~10μs per pair (feature vector comparison)
- **All-pairs comparison**: ~7ms for current catalog size

### Use Case Fit
**Good for functional similarity** - Detects templates doing same thing
**Limited for exact duplicates** - May miss text-level copies
**Best as complement** - Combine with other strategies

---

## Strategy 8: Hybrid Approach

### Description
Combine multiple strategies for different levels of similarity detection.

### Example: MinHash + Feature Hashing
1. **Tier 1**: MinHash for text-level similarity (catches copies/forks)
2. **Tier 2**: Feature vectors for functional similarity (catches reimplementations)
3. **Tier 3**: Manual review for edge cases

### Pros
- **Best of both worlds**: Combines strengths of multiple approaches
- **Layered detection**: Different strategies for different duplicate types
- **Configurable**: Can tune each layer independently
- **Comprehensive**: Catches more duplicate types

### Cons
- **Complex implementation**: Multiple systems to maintain
- **Higher storage**: Sum of all strategies
- **More computation**: Multiple passes over data
- **Harder to explain**: Why is X similar to Y?

### Implementation Complexity
**High** - Multiple systems to integrate

### Use Case Fit
**Best for production** - Most comprehensive
**Overkill for MVP** - Start with single strategy first

---

## Comparison Matrix

| Strategy | Storage/Template | Total Storage | Generation Time | Comparison Time | All-Pairs | Similarity Range | Complexity | Scales to 1000+ |
|----------|------------------|---------------|-----------------|-----------------|-----------|------------------|------------|----------------|
| SimHash | 8 B | 6 KB | 1 ms | 1 μs | 1 ms | 75-100% | Low | ✅ Excellent |
| MinHash+LSH | 512-1024 B | 366-732 KB | 10-20 ms | 10 μs | 50-100 ms | 5-100% | Medium | ✅ Excellent |
| Fuzzy Hash | 100-200 B | 70-140 KB | 5-10 ms | 100 μs | 51 s | 60-100% | Medium | ⚠️ Slow |
| Shingling | 1-5 KB | 1-4 MB | 2 ms | 1 ms | 5 min | 0-100% | Low | ❌ Too slow |
| Embeddings | 1.5-6 KB | 1-4.3 MB | 50-500 ms | 50 μs | 25 s | 0-100% | High | ⚠️ Expensive |
| AST | 1 KB | 732 KB | 2 ms | 10-100 ms | 10-100 min | Structural | High | ❌ Too slow |
| Features | 0 B (existing) | 0 KB | 0 ms | 10 μs | 7 ms | Functional | Low | ✅ Excellent |
| Hybrid | Sum of above | Varies | Sum | Sum | Sum | All | High | Depends |

---

## Recommendation: MinHash + LSH

### Why MinHash + LSH?

1. **Optimal similarity range** - Detects 5-100% similarity, covering:
   - Exact duplicates (100%)
   - Near-duplicates with minor changes (80-99%)
   - Forks with moderate changes (50-79%)
   - Loosely related templates (5-49%)

2. **Scalable performance** - Sub-linear search with LSH
   - Generation: 10-20ms per template (acceptable for daily runs)
   - Comparison: ~100ms for all 716 templates
   - Scales to 10,000+ templates

3. **Reasonable storage** - 512-1024 bytes per template
   - ~732KB total (smaller than current catalog)
   - Fits easily in data branch
   - No significant impact on repository size

4. **Battle-tested** - Proven in production
   - Used by AltaVista for duplicate detection
   - Google News for personalization
   - Common in code clone detection research

5. **Configurable** - Tune for different use cases
   - Adjust bands/rows to control false positive/negative rate
   - Set similarity threshold per use case
   - Get ranked results, not just binary match

6. **Incremental friendly** - Fits existing architecture
   - Compute MinHash once per template
   - Store in templates.jsonl alongside other fields
   - Only recompute when SHA changes (same as current analysis)

### Implementation Plan

#### Phase 1: Core Implementation
1. Add MinHash package to Go codebase
2. Extract k-shingles from template content (k=5 words recommended)
3. Generate MinHash signatures (128 hashes recommended)
4. Store signatures in templates.jsonl

#### Phase 2: LSH Index
1. Implement LSH banding (recommend 16 bands × 8 rows for ~50% threshold)
2. Build LSH index during analysis
3. Query index for similar templates

#### Phase 3: Integration
1. Add duplicate detection step to analysis pipeline
2. Store similarity scores in templates.jsonl
3. Add "similar_to" field with list of similar template IDs

#### Phase 4: Frontend
1. Display duplicate warnings in UI
2. Show similar templates for each template
3. Add filter for "unique templates only"

### Configuration Recommendations

```go
// MinHash configuration (matches actual implementation)
const (
    ShingleSize     = 5     // 5-word shingles (good for YAML)
    NumHashes       = 128   // 128 hash functions (standard)
    NumBands        = 32    // LSH bands (actual default)
    RowsPerBand     = 4     // LSH rows per band (32*4=128)
    SimilarityThreshold = 0.5  // 50% Jaccard similarity
)

// Duplicate classification
// >90% = Exact duplicate (likely fork)
// 70-90% = Near duplicate (minor changes)
// 50-70% = Similar (moderate changes)
// <50% = Different
```

### Parameter Selection Guide

#### Number of Hash Functions

**Formula**: Error rate ≈ 1/√n

| NumHashes | Error Rate | Storage/Template | Use Case |
|-----------|------------|------------------|----------|
| 64 | ~12.5% | 256 bytes | Speed-critical, 100K+ templates |
| 128 | ~8.8% | 512 bytes | **Recommended** - balanced accuracy/storage |
| 256 | ~6.3% | 1024 bytes | High accuracy needed, storage not a concern |
| 512 | ~4.4% | 2048 bytes | Legal/compliance, maximum accuracy |

**For Lima Catalog**:
- **128 hashes** chosen for good accuracy with reasonable storage
- Current catalog: ~512 bytes per template (acceptable overhead)
- Scales to 10,000+ templates without issues
- Standard choice in academic research

**When to change**:
- **Increase to 256** if you need higher accuracy for critical applications
- **Decrease to 64** if speed is critical and ~12% error is acceptable

#### Shingle Size (k-grams)

**Tradeoff**: Specificity vs Coverage

| Shingle Size | Specificity | False Positives | Use Case |
|--------------|-------------|-----------------|----------|
| k=2 | Low | High | Very short documents (< 10 words) |
| k=3 | Medium-Low | Medium-High | Short documents (10-30 words) |
| k=5 | **Balanced** | Low | **YAML templates, structured text** |
| k=7 | Medium-High | Very Low | Long documents (> 500 words) |
| k=10 | High | Very Low | Code files, need exact matching |

**For Lima Templates**:
- **k=5 words** chosen because it captures meaningful phrases:
  - Commands: "apt get install docker io"
  - Paths: "cloud images ubuntu com releases"
  - Config: "location https cloud images"
- Balances detection of similar templates vs false positives
- Standard for document similarity (used by Google, AltaVista)

**Examples of k=5 shingles**:
```
Template: "# Ubuntu-based development environment with Docker"
Shingles:
  - "ubuntu based development environment with"
  - "based development environment with docker"
```

**When to change**:
- **Use k=3** if templates are very short (< 20 words) or you want looser matching
- **Use k=7** if templates are very long (> 500 words) or you need to reduce false positives

#### Why These Specific Values?

**k=5 words**:
1. **Not too small**: k=2-3 generates common shingles like "apt get install" that appear everywhere
2. **Not too large**: k=7-10 is too specific, won't match templates differing by one word
3. **Perfect for YAML**: Captures structural patterns in configuration files
4. **Academic standard**: Commonly used in research papers on document similarity

**128 hashes**:
1. **Error sweet spot**: 8.8% error is acceptable for duplicate detection
2. **Storage efficiency**: 512 bytes per template is small
3. **Proven scale**: Used successfully for millions of documents
4. **Standard practice**: Common in production systems (Google News, web crawlers)

#### Tuning for Different Use Cases

**High Precision** (minimize false positives):
```go
mh := minhash.New(
    minhash.WithNumHashes(256),  // More accurate
    minhash.WithShingleSize(7),   // More specific
)
// Good for: Legal compliance, exact duplicate detection
```

**High Recall** (maximize detection):
```go
mh := minhash.New(
    minhash.WithNumHashes(128),  // Balanced
    minhash.WithShingleSize(3),   // Less specific
)
// Good for: Finding all potential duplicates for review
```

**Speed Optimized** (large datasets):
```go
mh := minhash.New(
    minhash.WithNumHashes(64),   // Faster
    minhash.WithShingleSize(5),   // Standard
)
// Good for: 100K+ templates, real-time processing
```

**Recommended for Lima** (balanced):
```go
mh := minhash.New()  // Uses defaults
// NumHashes: 128 (8.8% error, 512 bytes)
// ShingleSize: 5 (balanced matching)
// Good for: General-purpose duplicate detection
```

### Storage Schema Addition

```json
{
  "id": "owner/repo/path/template.yaml",
  "minhash": [uint32, uint32, ...],  // 128 values, ~512 bytes
  "similar_templates": [
    {
      "id": "other/repo/template.yaml",
      "similarity": 0.85
    }
  ],
  "is_duplicate": false,
  "primary_template": null  // If duplicate, ID of primary template
}
```

### Alternative: Start with Feature-Based Detection

If MinHash+LSH seems too complex for MVP, consider starting with **Strategy 7 (Feature Extraction)**:

**Pros:**
- Zero implementation cost (features already extracted)
- Instant results (no new computation needed)
- Good for functional similarity

**Implementation:**
```go
// Create feature vector from existing notability metrics
type FeatureVector struct {
    ImageDomains     []string  // Normalized domains
    ProvisionHashes  []uint64  // Hash of each provision script
    ParamNames       []string  // Parameter names (sorted)
    EnvNames         []string  // Env var names (sorted)
    Category         string
    Keywords         []string  // Sorted
}

// Compute similarity as Jaccard on feature sets
similarity := len(intersection) / float64(len(union))
```

Then add MinHash later for text-level duplicate detection.

---

## Decision Matrix

Choose **MinHash+LSH** if:
- ✅ Need comprehensive duplicate detection
- ✅ Want to detect both exact and approximate duplicates
- ✅ Willing to invest ~2-3 days implementation time
- ✅ 732KB storage overhead is acceptable

Choose **Feature-Based** if:
- ✅ Need quick MVP (hours, not days)
- ✅ Primarily care about functional similarity
- ✅ Want zero storage overhead
- ✅ Can accept missing some text-level duplicates

Choose **Hybrid** if:
- ✅ Need production-grade duplicate detection
- ✅ Have multiple weeks for implementation
- ✅ Want layered detection with different strategies

**Do NOT choose:**
- ❌ SimHash - Too limited similarity range
- ❌ Fuzzy Hashing - Templates too small, poor fit
- ❌ Pure Shingling - Storage and performance issues
- ❌ Embeddings - Over-engineered and expensive
- ❌ AST - Too slow for all-pairs comparison

---

## Implementation Effort Estimates

| Strategy | Implementation Time | Testing Time | Total |
|----------|-------------------|--------------|-------|
| Feature-Based | 4 hours | 2 hours | 6 hours |
| SimHash | 8 hours | 4 hours | 12 hours |
| MinHash+LSH | 16 hours | 8 hours | 24 hours |
| Fuzzy Hashing | 12 hours | 6 hours | 18 hours |
| Embeddings | 24 hours | 8 hours | 32 hours |
| AST-Based | 40 hours | 16 hours | 56 hours |
| Hybrid (MinHash+Features) | 24 hours | 12 hours | 36 hours |

---

## Conclusion

**Recommended approach:** Implement **MinHash + LSH** for comprehensive duplicate detection.

**Backup plan:** Start with **Feature-Based Detection** for quick MVP, then add MinHash later for text-level duplicates.

**Do not use:** SimHash (too limited), Fuzzy Hashing (poor fit), Embeddings (overkill), or AST (too slow).

The MinHash+LSH approach provides the best balance of:
- Accuracy (5-100% similarity range)
- Performance (sub-linear search)
- Storage (reasonable overhead)
- Scalability (proven at large scale)
- Implementation complexity (medium, but manageable)

This fits perfectly with the incremental analysis architecture and provides configurable duplicate detection that can be tuned for different thresholds.
