package commands

import (
	"literedis/internal/storage"
	"testing"
)

func TestHandleSAdd(t *testing.T) {
	s := storage.NewMemoryStorage()

	msg, err := handleSAdd(s, []string{"myset", "a", "b", "c"})
	if err != nil {
		t.Fatalf("handleSAdd failed: %v", err)
	}
	if msg.Type != "Integer" || msg.Content.(int) != 3 {
		t.Errorf("Expected Integer 3, got %v %v", msg.Type, msg.Content)
	}

	msg, err = handleSAdd(s, []string{"myset", "b", "c", "d"})
	if err != nil {
		t.Fatalf("handleSAdd failed: %v", err)
	}
	if msg.Type != "Integer" || msg.Content.(int) != 1 {
		t.Errorf("Expected Integer 1, got %v %v", msg.Type, msg.Content)
	}
}

func TestHandleSMembers(t *testing.T) {
	s := storage.NewMemoryStorage()
	s.SAdd("myset", "a", "b", "c")

	msg, err := handleSMembers(s, []string{"myset"})
	if err != nil {
		t.Fatalf("handleSMembers failed: %v", err)
	}
	if msg.Type != "Array" {
		t.Errorf("Expected Array type, got %v", msg.Type)
	}
	members := msg.Content.([]string)
	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}
	expectedMembers := map[string]bool{"a": true, "b": true, "c": true}
	for _, m := range members {
		if !expectedMembers[m] {
			t.Errorf("Unexpected member: %s", m)
		}
	}
}

func TestHandleSRem(t *testing.T) {
	s := storage.NewMemoryStorage()
	s.SAdd("myset", "a", "b", "c", "d")

	msg, err := handleSRem(s, []string{"myset", "b", "c", "e"})
	if err != nil {
		t.Fatalf("handleSRem failed: %v", err)
	}
	if msg.Type != "Integer" || msg.Content.(int) != 2 {
		t.Errorf("Expected Integer 2, got %v %v", msg.Type, msg.Content)
	}
}

func TestHandleSCard(t *testing.T) {
	s := storage.NewMemoryStorage()
	s.SAdd("myset", "a", "b", "c")

	msg, err := handleSCard(s, []string{"myset"})
	if err != nil {
		t.Fatalf("handleSCard failed: %v", err)
	}
	if msg.Type != "Integer" || msg.Content.(int) != 3 {
		t.Errorf("Expected Integer 3, got %v %v", msg.Type, msg.Content)
	}
}
