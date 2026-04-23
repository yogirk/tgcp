package core

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache()

	// Test Set and Get
	cache.Set("key1", "value1", 1*time.Minute)
	
	val, found := cache.Get("key1")
	if !found {
		t.Errorf("Expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// Test missing key
	_, found = cache.Get("missing")
	if found {
		t.Errorf("Expected missing key to return false")
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache()

	// Set with short TTL
	cache.Set("key1", "value1", 10*time.Millisecond)
	
	// Should be found immediately
	_, found := cache.Get("key1")
	if !found {
		t.Errorf("Expected to find key1 before expiration")
	}

	// Wait for expiration
	time.Sleep(15 * time.Millisecond)

	// Should not be found after expiration
	_, found = cache.Get("key1")
	if found {
		t.Errorf("Expected key1 to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache()

	cache.Set("key1", "value1", 1*time.Minute)
	cache.Delete("key1")

	_, found := cache.Get("key1")
	if found {
		t.Errorf("Expected key1 to be deleted")
	}
}

func TestCache_Flush(t *testing.T) {
	cache := NewCache()

	cache.Set("key1", "value1", 1*time.Minute)
	cache.Set("key2", "value2", 1*time.Minute)
	
	cache.Flush()

	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	
	if found1 || found2 {
		t.Errorf("Expected all keys to be flushed")
	}
}
