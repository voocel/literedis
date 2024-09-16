package storage

import (
	"testing"
	"time"
)

func TestMemoryStorage_Set_Get(t *testing.T) {
	s := NewMemoryStorage()

	err := s.Set("key1", []byte("value1"), 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, err := s.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(value) != "value1" {
		t.Errorf("Expected 'value1', got '%s'", string(value))
	}

	_, err = s.Get("non_existent_key")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryStorage_Set_Expiration(t *testing.T) {
	s := NewMemoryStorage()

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

func TestMemoryStorage_Del(t *testing.T) {
	s := NewMemoryStorage()

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

func TestMemoryStorage_Exists(t *testing.T) {
	s := NewMemoryStorage()

	s.Set("key1", []byte("value1"), 0)

	exists := s.Exists("key1")
	if !exists {
		t.Errorf("Expected key to exist")
	}

	exists = s.Exists("non_existent_key")
	if exists {
		t.Errorf("Expected key not to exist")
	}
}

func TestMemoryStorage_Expire(t *testing.T) {
	s := NewMemoryStorage()

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
}

func TestMemoryStorage_TTL(t *testing.T) {
	s := NewMemoryStorage()

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
}

func TestMemoryStorage_Type(t *testing.T) {
	s := NewMemoryStorage()

	s.Set("string_key", []byte("value"), 0)
	s.HSet("hash_key", map[string][]byte{"field": []byte("value")})
	s.LPush("list_key", []byte("value"))
	s.SAdd("set_key", "value")

	tests := []struct {
		key          string
		expectedType string
	}{
		{"string_key", "string"},
		{"hash_key", "hash"},
		{"list_key", "list"},
		{"set_key", "set"},
		{"non_existent_key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			keyType, err := s.Type(tt.key)
			if tt.expectedType == "" {
				if err != ErrKeyNotFound {
					t.Errorf("Expected ErrKeyNotFound, got %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("Type failed: %v", err)
				}
				if keyType != tt.expectedType {
					t.Errorf("Expected type %s, got %s", tt.expectedType, keyType)
				}
			}
		})
	}
}

func TestMemoryStorage_HSet_HGet(t *testing.T) {
	s := NewMemoryStorage()

	_, err := s.HSet("hash1", map[string][]byte{"field1": []byte("value1"), "field2": []byte("value2")})
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}

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

func TestMemoryStorage_LPush_RPush(t *testing.T) {
	s := NewMemoryStorage()

	length, err := s.LPush("list1", []byte("value1"), []byte("value2"))
	if err != nil {
		t.Fatalf("LPush failed: %v", err)
	}
	if length != 2 {
		t.Errorf("Expected length 2, got %d", length)
	}

	length, err = s.RPush("list1", []byte("value3"))
	if err != nil {
		t.Fatalf("RPush failed: %v", err)
	}
	if length != 3 {
		t.Errorf("Expected length 3, got %d", length)
	}
}

func TestMemoryStorage_LPop_RPop(t *testing.T) {
	s := NewMemoryStorage()

	s.LPush("list1", []byte("value1"), []byte("value2"), []byte("value3"))

	value, err := s.LPop("list1")
	if err != nil {
		t.Fatalf("LPop failed: %v", err)
	}
	if string(value) != "value3" {
		t.Errorf("Expected 'value3', got '%s'", string(value))
	}

	value, err = s.RPop("list1")
	if err != nil {
		t.Fatalf("RPop failed: %v", err)
	}
	if string(value) != "value1" {
		t.Errorf("Expected 'value1', got '%s'", string(value))
	}
}

func TestMemoryStorage_LLen(t *testing.T) {
	s := NewMemoryStorage()

	s.LPush("list1", []byte("value1"), []byte("value2"), []byte("value3"))

	length, err := s.LLen("list1")
	if err != nil {
		t.Fatalf("LLen failed: %v", err)
	}
	if length != 3 {
		t.Errorf("Expected length 3, got %d", length)
	}
}

func TestMemoryStorage_LRange(t *testing.T) {
	s := NewMemoryStorage()

	s.RPush("list1", []byte("value1"), []byte("value2"), []byte("value3"), []byte("value4"), []byte("value5"))

	values, err := s.LRange("list1", 1, 3)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}
	if string(values[0]) != "value2" || string(values[1]) != "value3" || string(values[2]) != "value4" {
		t.Errorf("Unexpected values: %v", values)
	}
}

func TestMemoryStorage_SAdd(t *testing.T) {
	s := NewMemoryStorage()

	added, err := s.SAdd("set1", "value1", "value2", "value3")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	if added != 3 {
		t.Errorf("Expected 3 elements added, got %d", added)
	}

	added, err = s.SAdd("set1", "value2", "value3", "value4")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	if added != 1 {
		t.Errorf("Expected 1 element added, got %d", added)
	}
}

func TestMemoryStorage_SMembers(t *testing.T) {
	s := NewMemoryStorage()

	s.SAdd("set1", "value1", "value2", "value3")

	members, err := s.SMembers("set1")
	if err != nil {
		t.Fatalf("SMembers failed: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}

	memberMap := make(map[string]bool)
	for _, m := range members {
		memberMap[m] = true
	}
	for _, expected := range []string{"value1", "value2", "value3"} {
		if !memberMap[expected] {
			t.Errorf("Expected member %s not found", expected)
		}
	}
}

func TestMemoryStorage_SRem(t *testing.T) {
	s := NewMemoryStorage()

	s.SAdd("set1", "value1", "value2", "value3", "value4")

	removed, err := s.SRem("set1", "value2", "value3", "value5")
	if err != nil {
		t.Fatalf("SRem failed: %v", err)
	}
	if removed != 2 {
		t.Errorf("Expected 2 elements removed, got %d", removed)
	}

	members, _ := s.SMembers("set1")
	if len(members) != 2 {
		t.Errorf("Expected 2 remaining members, got %d", len(members))
	}
}

func TestMemoryStorage_SCard(t *testing.T) {
	s := NewMemoryStorage()

	s.SAdd("set1", "value1", "value2", "value3")

	count, err := s.SCard("set1")
	if err != nil {
		t.Fatalf("SCard failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count of 3, got %d", count)
	}

	s.SRem("set1", "value1")
	count, _ = s.SCard("set1")
	if count != 2 {
		t.Errorf("Expected count of 2 after removal, got %d", count)
	}
}

func TestMemoryStorage_Select(t *testing.T) {
	s := NewMemoryStorage()

	err := s.Select(0)
	if err != nil {
		t.Errorf("Select(0) failed: %v", err)
	}

	err = s.Select(15)
	if err != nil {
		t.Errorf("Select(15) failed: %v", err)
	}

	err = s.Select(16)
	if err != ErrInvalidDBIndex {
		t.Errorf("Expected ErrInvalidDBIndex, got %v", err)
	}

	err = s.Select(-1)
	if err != ErrInvalidDBIndex {
		t.Errorf("Expected ErrInvalidDBIndex, got %v", err)
	}
}

func TestMemoryStorage_FlushDB(t *testing.T) {
	s := NewMemoryStorage()

	s.Set("key1", []byte("value1"), 0)
	s.HSet("hash1", map[string][]byte{"field1": []byte("value1")})
	s.LPush("list1", []byte("value1"))
	s.SAdd("set1", "value1")

	err := s.FlushDB()
	if err != nil {
		t.Fatalf("FlushDB failed: %v", err)
	}

	if s.Exists("key1") || s.Exists("hash1") || s.Exists("list1") || s.Exists("set1") {
		t.Errorf("Expected all keys to be removed after FlushDB")
	}
}

func TestMemoryStorage_Flush(t *testing.T) {
	s := NewMemoryStorage()

	s.Set("key1", []byte("value1"), 0)
	s.Select(1)
	s.Set("key2", []byte("value2"), 0)

	err := s.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	s.Select(0)
	if s.Exists("key1") {
		t.Errorf("Expected key1 to be removed after Flush")
	}

	s.Select(1)
	if s.Exists("key2") {
		t.Errorf("Expected key2 to be removed after Flush")
	}
}

func TestMemoryStorage_MultiDB(t *testing.T) {
	s := NewMemoryStorage()

	s.Set("key1", []byte("value1"), 0)
	s.Select(1)
	s.Set("key2", []byte("value2"), 0)

	s.Select(0)
	value, err := s.Get("key1")
	if err != nil || string(value) != "value1" {
		t.Errorf("Expected to get 'value1' from DB 0")
	}
	_, err = s.Get("key2")
	if err != ErrKeyNotFound {
		t.Errorf("Expected key2 to not exist in DB 0")
	}

	s.Select(1)
	value, err = s.Get("key2")
	if err != nil || string(value) != "value2" {
		t.Errorf("Expected to get 'value2' from DB 1")
	}
	_, err = s.Get("key1")
	if err != ErrKeyNotFound {
		t.Errorf("Expected key1 to not exist in DB 1")
	}
}
