package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemorySetStorage_SAdd(t *testing.T) {
	s := NewMemorySetStorage()

	added, err := s.SAdd("myset", "a", "b", "c")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	if added != 3 {
		t.Errorf("Expected 3 elements added, got %d", added)
	}

	added, err = s.SAdd("myset", "b", "c", "d")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	if added != 1 {
		t.Errorf("Expected 1 element added, got %d", added)
	}
}

func TestMemorySetStorage_SMembers(t *testing.T) {
	s := NewMemorySetStorage()
	s.SAdd("myset", "a", "b", "c")

	members, err := s.SMembers("myset")
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
	for _, expected := range []string{"a", "b", "c"} {
		if !memberMap[expected] {
			t.Errorf("Expected member %s not found", expected)
		}
	}
}

func TestMemorySetStorage_SRem(t *testing.T) {
	s := NewMemorySetStorage()
	s.SAdd("myset", "a", "b", "c", "d")

	removed, err := s.SRem("myset", "b", "c", "e")
	if err != nil {
		t.Fatalf("SRem failed: %v", err)
	}
	if removed != 2 {
		t.Errorf("Expected 2 elements removed, got %d", removed)
	}

	members, _ := s.SMembers("myset")
	if len(members) != 2 {
		t.Errorf("Expected 2 remaining members, got %d", len(members))
	}
}

func TestMemorySetStorage_SCard(t *testing.T) {
	s := NewMemorySetStorage()
	s.SAdd("myset", "a", "b", "c")

	count, err := s.SCard("myset")
	if err != nil {
		t.Fatalf("SCard failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count of 3, got %d", count)
	}

	s.SRem("myset", "a")
	count, _ = s.SCard("myset")
	if count != 2 {
		t.Errorf("Expected count of 2 after removal, got %d", count)
	}
}

func TestIntSetOptimization(t *testing.T) {
	s := NewMemorySetStorage()

	// 测试纯整数集合使用 IntSet
	added, _ := s.SAdd("intset", "1", "2", "3")
	assert.Equal(t, 3, added)

	members, _ := s.SMembers("intset")
	assert.ElementsMatch(t, []string{"1", "2", "3"}, members)

	// 测试添加非整数元素时转换为哈希表
	added, _ = s.SAdd("intset", "non-integer")
	assert.Equal(t, 1, added)

	members, _ = s.SMembers("intset")
	assert.ElementsMatch(t, []string{"1", "2", "3", "non-integer"}, members)
}
