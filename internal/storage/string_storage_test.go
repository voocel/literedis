package storage

import (
	"bytes"
	"testing"
	"time"
)

func TestMemoryStringStorage_Set_Get(t *testing.T) {
	s := NewMemoryStringStorage()

	err := s.Set("key1", []byte("value1"), 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, err := s.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !bytes.Equal(value, []byte("value1")) {
		t.Errorf("Expected 'value1', got '%s'", string(value))
	}

	_, err = s.Get("non_existent_key")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryStringStorage_Set_Expiration(t *testing.T) {
	s := NewMemoryStringStorage()

	err := s.Set("key1", []byte("value1"), 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	_, err = s.Get("key1")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryStringStorage_Del(t *testing.T) {
	s := NewMemoryStringStorage()

	s.Set("key1", []byte("value1"), 0)
	s.Set("key2", []byte("value2"), 0)

	deleted, err := s.Del("key1")
	if err != nil {
		t.Fatalf("Del failed: %v", err)
	}
	if !deleted {
		t.Errorf("Expected key to be deleted")
	}

	_, err = s.Get("key1")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}

	deleted, err = s.Del("non_existent_key")
	if err != nil {
		t.Fatalf("Del failed: %v", err)
	}
	if deleted {
		t.Errorf("Expected key not to be deleted")
	}
}

func TestMemoryStringStorage_Exists(t *testing.T) {
	s := NewMemoryStringStorage()

	s.Set("key1", []byte("value1"), 0)

	exists, err := s.Exists("key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Errorf("Expected key to exist")
	}

	exists, err = s.Exists("non_existent_key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Errorf("Expected key not to exist")
	}
}

func TestMemoryStringStorage_Expire(t *testing.T) {
	s := NewMemoryStringStorage()

	s.Set("key1", []byte("value1"), 0)

	ok, err := s.Expire("key1", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Expire failed: %v", err)
	}
	if !ok {
		t.Errorf("Expected Expire to return true")
	}

	time.Sleep(100 * time.Millisecond)

	_, err = s.Get("key1")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}

	ok, err = s.Expire("non_existent_key", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Expire failed: %v", err)
	}
	if ok {
		t.Errorf("Expected Expire to return false for non-existent key")
	}
}

func TestMemoryStringStorage_TTL(t *testing.T) {
	s := NewMemoryStringStorage()

	s.Set("key1", []byte("value1"), 1*time.Second)

	ttl, err := s.TTL("key1")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if ttl <= 0 || ttl > 1*time.Second {
		t.Errorf("Expected TTL to be between 0 and 1 second, got %v", ttl)
	}

	ttl, err = s.TTL("non_existent_key")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if ttl != -2 {
		t.Errorf("Expected TTL to be -2 for non-existent key, got %v", ttl)
	}

	s.Set("key2", []byte("value2"), 0)
	ttl, err = s.TTL("key2")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if ttl != -1 {
		t.Errorf("Expected TTL to be -1 for key with no expiration, got %v", ttl)
	}
}
