package storage

import (
	"bytes"
	"testing"
)

func TestMemoryStringStorage_Set_Get(t *testing.T) {
	s := NewMemoryStringStorage()

	err := s.Set("key1", []byte("value1"))
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

func TestMemoryStringStorage_Del(t *testing.T) {
	s := NewMemoryStringStorage()

	s.Set("key1", []byte("value1"))
	s.Set("key2", []byte("value2"))

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

	s.Set("key1", []byte("value1"))

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
