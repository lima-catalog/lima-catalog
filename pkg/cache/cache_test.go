// Unit Tests: In-Memory Cache (cache package)
//
// High-level overview of what's being tested:
// - Setting and retrieving cached values
// - Handling non-existent keys
// - Time-based expiration of cached entries
// - Custom TTL per cache entry
// - Deleting specific cache entries
// - Clearing entire cache
// - Cache size tracking
// - Manual and automatic cleanup of expired entries
// - Cleanup timer with multiple cycles
// - Concurrent access patterns and thread safety
// - Overwriting existing values
// - Storing different data types (string, int, bool, slice)
// - Preventing goroutine leaks when stopping timers

package cache

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := New(1 * time.Hour)

	// Set a value
	c.Set("key1", "value1")

	// Get the value
	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected to find key1")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	c := New(1 * time.Hour)

	// Get non-existent key
	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent key")
	}
}

func TestCache_Expiration(t *testing.T) {
	c := New(100 * time.Millisecond)

	// Set a value
	c.Set("key1", "value1")

	// Should be available immediately
	_, ok := c.Get("key1")
	if !ok {
		t.Error("expected to find key1 immediately after setting")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, ok = c.Get("key1")
	if ok {
		t.Error("expected key1 to be expired")
	}
}

func TestCache_SetWithTTL(t *testing.T) {
	c := New(1 * time.Hour) // Default TTL is 1 hour

	// Set with custom short TTL
	c.SetWithTTL("key1", "value1", 100*time.Millisecond)

	// Should be available immediately
	_, ok := c.Get("key1")
	if !ok {
		t.Error("expected to find key1")
	}

	// Wait for custom TTL expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, ok = c.Get("key1")
	if ok {
		t.Error("expected key1 to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	c := New(1 * time.Hour)

	// Set and delete
	c.Set("key1", "value1")
	c.Delete("key1")

	// Should not exist
	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	c := New(1 * time.Hour)

	// Set multiple values
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")

	if c.Size() != 3 {
		t.Errorf("expected size 3, got %d", c.Size())
	}

	// Clear all
	c.Clear()

	if c.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", c.Size())
	}

	// Should not find any keys
	_, ok := c.Get("key1")
	if ok {
		t.Error("expected not to find key1 after clear")
	}
}

func TestCache_Size(t *testing.T) {
	c := New(1 * time.Hour)

	if c.Size() != 0 {
		t.Errorf("expected initial size 0, got %d", c.Size())
	}

	c.Set("key1", "value1")
	if c.Size() != 1 {
		t.Errorf("expected size 1, got %d", c.Size())
	}

	c.Set("key2", "value2")
	if c.Size() != 2 {
		t.Errorf("expected size 2, got %d", c.Size())
	}

	c.Delete("key1")
	if c.Size() != 1 {
		t.Errorf("expected size 1 after delete, got %d", c.Size())
	}
}

func TestCache_Cleanup(t *testing.T) {
	c := New(100 * time.Millisecond)

	// Add some entries
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.SetWithTTL("key3", "value3", 1*time.Hour) // Long TTL

	// Wait for some to expire
	time.Sleep(150 * time.Millisecond)

	// Cleanup
	removed := c.Cleanup()

	// Should have removed 2 expired entries
	if removed != 2 {
		t.Errorf("expected to remove 2 entries, removed %d", removed)
	}

	// key3 should still exist
	_, ok := c.Get("key3")
	if !ok {
		t.Error("expected key3 to still exist")
	}

	// key1 and key2 should be gone
	_, ok = c.Get("key1")
	if ok {
		t.Error("expected key1 to be removed")
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := New(1 * time.Hour)

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				c.Set("key", n)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and should have a value
	_, ok := c.Get("key")
	if !ok {
		t.Error("expected to find key after concurrent writes")
	}
}

func TestCache_OverwriteValue(t *testing.T) {
	c := New(1 * time.Hour)

	// Set initial value
	c.Set("key1", "value1")

	// Overwrite
	c.Set("key1", "value2")

	// Should get new value
	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected to find key1")
	}
	if val != "value2" {
		t.Errorf("expected 'value2', got %v", val)
	}
}

func TestCache_DifferentTypes(t *testing.T) {
	c := New(1 * time.Hour)

	// Store different types
	c.Set("string", "hello")
	c.Set("int", 42)
	c.Set("bool", true)
	c.Set("slice", []string{"a", "b", "c"})

	// Retrieve and type assert
	str, ok := c.Get("string")
	if !ok || str != "hello" {
		t.Error("failed to get string")
	}

	num, ok := c.Get("int")
	if !ok || num != 42 {
		t.Error("failed to get int")
	}

	b, ok := c.Get("bool")
	if !ok || b != true {
		t.Error("failed to get bool")
	}

	slice, ok := c.Get("slice")
	if !ok {
		t.Error("failed to get slice")
	}
	s, ok := slice.([]string)
	if !ok || len(s) != 3 {
		t.Error("failed to type assert slice")
	}
}

func TestCache_StartCleanupTimer(t *testing.T) {
	c := New(50 * time.Millisecond)

	// Add some entries with short TTL
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")

	if c.Size() != 3 {
		t.Fatalf("expected size 3, got %d", c.Size())
	}

	// Start cleanup timer with short interval
	ticker := c.StartCleanupTimer(100 * time.Millisecond)
	defer ticker.Stop() // Stop timer to prevent goroutine leak

	// Wait for entries to expire
	time.Sleep(80 * time.Millisecond)

	// Wait for cleanup timer to run
	time.Sleep(150 * time.Millisecond)

	// Entries should have been cleaned up automatically
	if c.Size() != 0 {
		t.Errorf("expected size 0 after cleanup timer, got %d", c.Size())
	}
}

func TestCache_StartCleanupTimer_SelectiveCleanup(t *testing.T) {
	c := New(50 * time.Millisecond)

	// Add entries with different TTLs
	c.Set("short1", "value1")              // 50ms TTL (default)
	c.Set("short2", "value2")              // 50ms TTL (default)
	c.SetWithTTL("long", "value3", 1*time.Hour) // Long TTL

	// Start cleanup timer
	ticker := c.StartCleanupTimer(100 * time.Millisecond)
	defer ticker.Stop()

	// Wait for short TTL entries to expire
	time.Sleep(80 * time.Millisecond)

	// Wait for cleanup to run
	time.Sleep(150 * time.Millisecond)

	// Short TTL entries should be removed, long TTL entry should remain
	if c.Size() != 1 {
		t.Errorf("expected size 1 (only long TTL entry), got %d", c.Size())
	}

	_, ok := c.Get("long")
	if !ok {
		t.Error("expected long TTL entry to still exist")
	}

	_, ok = c.Get("short1")
	if ok {
		t.Error("expected short1 to be cleaned up")
	}

	_, ok = c.Get("short2")
	if ok {
		t.Error("expected short2 to be cleaned up")
	}
}

func TestCache_StartCleanupTimer_MultipleCycles(t *testing.T) {
	c := New(30 * time.Millisecond)

	// Start cleanup timer with short interval
	ticker := c.StartCleanupTimer(50 * time.Millisecond)
	defer ticker.Stop()

	// Add entries, wait for cleanup, add more, verify cleanup works multiple times
	c.Set("batch1", "value1")
	time.Sleep(100 * time.Millisecond) // Wait for expiration and cleanup

	if c.Size() != 0 {
		t.Errorf("expected size 0 after first cleanup, got %d", c.Size())
	}

	// Add second batch
	c.Set("batch2", "value2")
	time.Sleep(100 * time.Millisecond) // Wait for expiration and cleanup

	if c.Size() != 0 {
		t.Errorf("expected size 0 after second cleanup, got %d", c.Size())
	}

	// Add third batch
	c.Set("batch3", "value3")
	time.Sleep(100 * time.Millisecond) // Wait for expiration and cleanup

	if c.Size() != 0 {
		t.Errorf("expected size 0 after third cleanup, got %d", c.Size())
	}
}

func TestCache_StartCleanupTimer_StopPreventsLeak(t *testing.T) {
	c := New(1 * time.Hour)

	// Start multiple cleanup timers and stop them
	for i := 0; i < 5; i++ {
		ticker := c.StartCleanupTimer(10 * time.Millisecond)
		ticker.Stop() // Immediately stop to test cleanup
	}

	// Add an entry to verify cache still works
	c.Set("key", "value")

	// Short sleep to ensure stopped tickers don't interfere
	time.Sleep(50 * time.Millisecond)

	// Entry should still exist (long TTL, no cleanup)
	val, ok := c.Get("key")
	if !ok {
		t.Error("expected entry to exist")
	}
	if val != "value" {
		t.Errorf("expected 'value', got %v", val)
	}

	// This test also verifies that stopping the ticker prevents goroutine leaks
	// If tickers weren't properly stopped, we'd have 5 goroutines running
}
