package storage

import (
	"bytes"
	"testing"
)

func TestMemoryListStorage_LPush_LPop(t *testing.T) {
	s := NewMemoryListStorage()

	length, err := s.LPush("mylist", []byte("value1"), []byte("value2"))
	if err != nil {
		t.Fatalf("LPush failed: %v", err)
	}
	if length != 2 {
		t.Errorf("Expected length 2, got %d", length)
	}

	value, err := s.LPop("mylist")
	if err != nil {
		t.Fatalf("LPop failed: %v", err)
	}
	if !bytes.Equal(value, []byte("value2")) {
		t.Errorf("Expected 'value2', got '%s'", string(value))
	}

	value, err = s.LPop("mylist")
	if err != nil {
		t.Fatalf("LPop failed: %v", err)
	}
	if !bytes.Equal(value, []byte("value1")) {
		t.Errorf("Expected 'value1', got '%s'", string(value))
	}

	_, err = s.LPop("mylist")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryListStorage_RPush_RPop(t *testing.T) {
	s := NewMemoryListStorage()

	length, err := s.RPush("mylist", []byte("value1"), []byte("value2"))
	if err != nil {
		t.Fatalf("RPush failed: %v", err)
	}
	if length != 2 {
		t.Errorf("Expected length 2, got %d", length)
	}

	value, err := s.RPop("mylist")
	if err != nil {
		t.Fatalf("RPop failed: %v", err)
	}
	if !bytes.Equal(value, []byte("value2")) {
		t.Errorf("Expected 'value2', got '%s'", string(value))
	}

	value, err = s.RPop("mylist")
	if err != nil {
		t.Fatalf("RPop failed: %v", err)
	}
	if !bytes.Equal(value, []byte("value1")) {
		t.Errorf("Expected 'value1', got '%s'", string(value))
	}

	_, err = s.RPop("mylist")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryListStorage_LLen(t *testing.T) {
	s := NewMemoryListStorage()

	s.LPush("mylist", []byte("value1"), []byte("value2"))

	length, err := s.LLen("mylist")
	if err != nil {
		t.Fatalf("LLen failed: %v", err)
	}
	if length != 2 {
		t.Errorf("Expected length 2, got %d", length)
	}

	length, err = s.LLen("non_existent_list")
	if err != nil {
		t.Fatalf("LLen failed: %v", err)
	}
	if length != 0 {
		t.Errorf("Expected length 0 for non-existent list, got %d", length)
	}
}

func TestMemoryListStorage_LRange(t *testing.T) {
	s := NewMemoryListStorage()

	s.RPush("mylist", []byte("one"), []byte("two"), []byte("three"))

	values, err := s.LRange("mylist", 0, -1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}
	if !bytes.Equal(values[0], []byte("one")) || !bytes.Equal(values[1], []byte("two")) || !bytes.Equal(values[2], []byte("three")) {
		t.Errorf("Unexpected values: %v", values)
	}

	values, err = s.LRange("mylist", 0, 1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
	if !bytes.Equal(values[0], []byte("one")) || !bytes.Equal(values[1], []byte("two")) {
		t.Errorf("Unexpected values: %v", values)
	}

	values, err = s.LRange("non_existent_list", 0, -1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(values) != 0 {
		t.Errorf("Expected empty slice for non-existent list, got %v", values)
	}
}
