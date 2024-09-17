package storage

import (
	"testing"
)

func TestMemoryStorage_HSet(t *testing.T) {
	s := NewMemoryStorage()

	count, err := s.HSet("hash1", map[string][]byte{"field1": []byte("value1"), "field2": []byte("value2")})
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Test updating existing field
	count, err = s.HSet("hash1", map[string][]byte{"field1": []byte("new_value1")})
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 for updating existing field, got %d", count)
	}
}

func TestMemoryStorage_HGet(t *testing.T) {
	s := NewMemoryStorage()

	s.HSet("hash1", map[string][]byte{"field1": []byte("value1")})

	value, err := s.HGet("hash1", "field1")
	if err != nil {
		t.Fatalf("HGet failed: %v", err)
	}
	if string(value) != "value1" {
		t.Errorf("Expected 'value1', got '%s'", string(value))
	}

	_, err = s.HGet("hash1", "non_existent_field")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryStorage_HDel(t *testing.T) {
	s := NewMemoryStorage()

	s.HSet("hash1", map[string][]byte{"field1": []byte("value1"), "field2": []byte("value2")})

	deleted, err := s.HDel("hash1", "field1", "non_existent_field")
	if err != nil {
		t.Fatalf("HDel failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("Expected 1 field to be deleted, got %d", deleted)
	}

	_, err = s.HGet("hash1", "field1")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryStorage_HLen(t *testing.T) {
	s := NewMemoryStorage()

	s.HSet("hash1", map[string][]byte{"field1": []byte("value1"), "field2": []byte("value2")})

	length, err := s.HLen("hash1")
	if err != nil {
		t.Fatalf("HLen failed: %v", err)
	}
	if length != 2 {
		t.Errorf("Expected length 2, got %d", length)
	}
}
