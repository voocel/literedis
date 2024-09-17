package dsset

import (
	"strconv"
	"sync"
	"time"
)

type BasicSet struct {
	encoding int
	intset   *IntSet
	dict     map[string]struct{}
	expireAt time.Time
	mu       sync.RWMutex
}

func NewBasicSet() *BasicSet {
	return &BasicSet{
		encoding: useIntSet,
		intset:   NewIntSet(),
	}
}

func (s *BasicSet) Add(members ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	added := 0
	for _, member := range members {
		if s.encoding == useIntSet {
			if i, err := strconv.ParseInt(member, 10, 64); err == nil {
				if s.intset.Add(i) {
					added++
				}
			} else {
				s.convertToHashTable()
			}
		}

		if s.encoding == useHashTable {
			if _, exists := s.dict[member]; !exists {
				s.dict[member] = struct{}{}
				added++
			}
		}
	}
	return added
}

func (s *BasicSet) Remove(members ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := 0
	for _, member := range members {
		if s.encoding == useIntSet {
			if i, err := strconv.ParseInt(member, 10, 64); err == nil {
				if s.intset.Remove(i) {
					removed++
				}
			}
		} else {
			if _, exists := s.dict[member]; exists {
				delete(s.dict, member)
				removed++
			}
		}
	}
	return removed
}

func (s *BasicSet) IsMember(member string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.encoding == useIntSet {
		if i, err := strconv.ParseInt(member, 10, 64); err == nil {
			return s.intset.Contains(i)
		}
		return false
	}
	_, exists := s.dict[member]
	return exists
}

func (s *BasicSet) Members() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.encoding == useIntSet {
		ints := s.intset.ToSlice()
		members := make([]string, len(ints))
		for i, v := range ints {
			members[i] = strconv.FormatInt(v, 10)
		}
		return members
	}
	members := make([]string, 0, len(s.dict))
	for member := range s.dict {
		members = append(members, member)
	}
	return members
}

func (s *BasicSet) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.encoding == useIntSet {
		return s.intset.Len()
	}
	return len(s.dict)
}

func (s *BasicSet) SetExpire(expireAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireAt = expireAt
}

func (s *BasicSet) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return !s.expireAt.IsZero() && time.Now().After(s.expireAt)
}

func (s *BasicSet) convertToHashTable() {
	if s.encoding == useHashTable {
		return
	}
	s.dict = make(map[string]struct{})
	for _, i := range s.intset.ToSlice() {
		s.dict[strconv.FormatInt(i, 10)] = struct{}{}
	}
	s.encoding = useHashTable
	s.intset = nil
}

func (s *BasicSet) Union(other Set) Set {
	result := NewBasicSet()
	for _, member := range s.Members() {
		result.Add(member)
	}
	for _, member := range other.Members() {
		result.Add(member)
	}
	return result
}

func (s *BasicSet) Intersection(other Set) Set {
	result := NewBasicSet()
	for _, member := range s.Members() {
		if other.IsMember(member) {
			result.Add(member)
		}
	}
	return result
}

func (s *BasicSet) Difference(other Set) Set {
	result := NewBasicSet()
	for _, member := range s.Members() {
		if !other.IsMember(member) {
			result.Add(member)
		}
	}
	return result
}
