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
