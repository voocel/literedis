package storage

import (
	"testing"
)

func TestMemoryHashStorage_HSet(t *testing.T) {
	s := NewMemoryHashStorage()

	count, err := s.HSet("myhash", map[string][]byte{"field1": []byte("value1"), "field2": []byte("value2")})
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Test updating existing field
	count, err = s.HSet("myhash", map[string][]byte{"field1": []byte("new_value1")})
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 for updating existing field, got %d", count)
	}
}

func TestMemoryHashStorage_HGet(t *testing.T) {
	s := NewMemoryHashStorage()

	s.HSet("myhash", map[string][]byte{"field1": []byte("value1")})

	value, err := s.HGet("myhash", "field1")
	if err != nil {
		t.Fatalf("HGet failed: %v", err)
	}
	if string(value) != "value1" {
		t.Errorf("Expected 'value1', got '%s'", string(value))
	}

	_, err = s.HGet("myhash", "non_existent_field")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}

	_, err = s.HGet("non_existent_hash", "field1")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryHashStorage_HDel(t *testing.T) {
	s := NewMemoryHashStorage()

	s.HSet("myhash", map[string][]byte{"field1": []byte("value1"), "field2": []byte("value2")})

	count, err := s.HDel("myhash", "field1", "field2", "non_existent_field")
	if err != nil {
		t.Fatalf("HDel failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 fields deleted, got %d", count)
	}

	count, err = s.HDel("non_existent_hash", "field1")
	if err != nil {
		t.Fatalf("HDel failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 fields deleted for non-existent hash, got %d", count)
	}
}

func TestMemoryHashStorage_HLen(t *testing.T) {
	s := NewMemoryHashStorage()

	s.HSet("myhash", map[string][]byte{"field1": []byte("value1"), "field2": []byte("value2")})

	length, err := s.HLen("myhash")
	if err != nil {
		t.Fatalf("HLen failed: %v", err)
	}
	if length != 2 {
		t.Errorf("Expected length 2, got %d", length)
	}

	length, err = s.HLen("non_existent_hash")
	if err != nil {
		t.Fatalf("HLen failed: %v", err)
	}
	if length != 0 {
		t.Errorf("Expected length 0 for non-existent hash, got %d", length)
	}
}
