package storage

import (
	"sync"
)

type MemorySetStorage struct {
	data map[string]map[string]struct{}
	mu   sync.RWMutex
}

func NewMemorySetStorage() *MemorySetStorage {
	return &MemorySetStorage{
		data: make(map[string]map[string]struct{}),
	}
}

func (s *MemorySetStorage) SAdd(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; !ok {
		s.data[key] = make(map[string]struct{})
	}

	added := 0
	for _, member := range members {
		if _, exists := s.data[key][member]; !exists {
			s.data[key][member] = struct{}{}
			added++
		}
	}

	return added, nil
}

func (s *MemorySetStorage) SMembers(key string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if set, ok := s.data[key]; ok {
		members := make([]string, 0, len(set))
		for member := range set {
			members = append(members, member)
		}
		return members, nil
	}

	return []string{}, nil
}

func (s *MemorySetStorage) SRem(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if set, ok := s.data[key]; ok {
		removed := 0
		for _, member := range members {
			if _, exists := set[member]; exists {
				delete(set, member)
				removed++
			}
		}
		if len(set) == 0 {
			delete(s.data, key)
		}
		return removed, nil
	}

	return 0, nil
}

func (s *MemorySetStorage) SCard(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if set, ok := s.data[key]; ok {
		return len(set), nil
	}

	return 0, nil
}
