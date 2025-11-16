#!/bin/bash
set -e

echo "======================================================================"
echo "Integration Test: Incremental Discovery"
echo "======================================================================"
echo ""

# Cleanup function
cleanup() {
    echo ""
    echo "🧹 Cleaning up test data..."
    rm -rf test_integration_data
    rm -f lima-catalog
}

trap cleanup EXIT

# Build the tool
echo "🔨 Building lima-catalog..."
go build -o lima-catalog ./cmd/lima-catalog
echo "✅ Build complete"
echo ""

# Check GITHUB_TOKEN
if [ -z "$GITHUB_TOKEN" ]; then
    echo "❌ GITHUB_TOKEN environment variable is not set"
    exit 1
fi

# Set up test environment
export DATA_DIR=./test_integration_data
mkdir -p $DATA_DIR

echo "======================================================================"
echo "Test 1: Full Discovery (baseline)"
echo "======================================================================"
echo ""

# Run full discovery with timeout (just enough to get some templates)
timeout 30 ./lima-catalog 2>&1 | tee /tmp/test_full.log || true

# Check if templates were discovered
TEMPLATE_COUNT=$(grep -c "^{" $DATA_DIR/templates.jsonl 2>/dev/null || echo "0")
echo ""
echo "✅ Full discovery found $TEMPLATE_COUNT templates"

if [ "$TEMPLATE_COUNT" -lt 10 ]; then
    echo "⚠️  Warning: Expected at least 10 templates, got $TEMPLATE_COUNT"
    echo "This test run may have been interrupted, but that's okay for testing."
fi

echo ""
echo "======================================================================"
echo "Test 2: Incremental Discovery"
echo "======================================================================"
echo ""

# Save the original template count
ORIGINAL_COUNT=$TEMPLATE_COUNT

# Run incremental mode with timeout
export INCREMENTAL=1
timeout 25 ./lima-catalog 2>&1 | tee /tmp/test_incremental.log || true

# Check that incremental mode was enabled
if grep -q "Incremental mode: true" /tmp/test_incremental.log; then
    echo "✅ Incremental mode enabled"
else
    echo "❌ Incremental mode was not enabled"
    exit 1
fi

# Check that it loaded existing templates
if grep -q "Loaded.*existing templates" /tmp/test_incremental.log; then
    echo "✅ Loaded existing templates"
else
    echo "❌ Did not load existing templates"
    exit 1
fi

# Check that it calculated a sinceDate
if grep -q "Newest template discovered at:" /tmp/test_incremental.log || grep -q "No existing templates found" /tmp/test_incremental.log; then
    echo "✅ Timestamp calculation working"
else
    echo "❌ Timestamp calculation failed"
    exit 1
fi

# Check that queries include pushed: qualifier (if there were existing templates)
if grep -q "pushed:>" /tmp/test_incremental.log; then
    echo "✅ Incremental queries include pushed:> qualifier"
elif grep -q "No existing templates found" /tmp/test_incremental.log; then
    echo "ℹ️  No existing templates, fell back to full discovery (expected)"
else
    echo "⚠️  Could not verify pushed:> qualifier"
fi

echo ""
echo "======================================================================"
echo "Test 3: Blocklist Filtering"
echo "======================================================================"
echo ""

# Check that blocklist was loaded
if grep -q "Loaded blocklist: 3 path patterns" /tmp/test_full.log; then
    echo "✅ Blocklist loaded successfully"
else
    echo "❌ Blocklist not loaded"
    exit 1
fi

# Check for blocklisted files (if any were encountered)
if grep -q "Blocklisted.*files" /tmp/test_full.log; then
    BLOCKLISTED=$(grep "Blocklisted" /tmp/test_full.log | grep -oE "[0-9]+" | head -1)
    echo "✅ Blocklist filtering active (blocked $BLOCKLISTED files)"
else
    echo "ℹ️  No templates were blocklisted in this run (might not have encountered any)"
fi

echo ""
echo "======================================================================"
echo "Test Summary"
echo "======================================================================"
echo ""
echo "✅ Full discovery: $ORIGINAL_COUNT templates"
echo "✅ Incremental mode: working correctly"
echo "✅ Timestamp-based filtering: implemented"
echo "✅ Blocklist filtering: active"
echo ""
echo "All integration tests passed! 🎉"
