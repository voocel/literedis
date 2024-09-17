package storage

import (
	"strconv"
	"sync"
)

const (
	useIntSet = iota
	useHashTable
)

type IntSet struct {
	contents []int64
}

func NewIntSet() *IntSet {
	return &IntSet{contents: make([]int64, 0)}
}

func (is *IntSet) Add(value int64) bool {
	for i, v := range is.contents {
		if v == value {
			return false
		}
		if v > value {
			is.contents = append(is.contents, 0)
			copy(is.contents[i+1:], is.contents[i:])
			is.contents[i] = value
			return true
		}
	}
	is.contents = append(is.contents, value)
	return true
}

func (is *IntSet) Remove(value int64) bool {
	for i, v := range is.contents {
		if v == value {
			is.contents = append(is.contents[:i], is.contents[i+1:]...)
			return true
		}
		if v > value {
			return false
		}
	}
	return false
}

func (is *IntSet) Contains(value int64) bool {
	for _, v := range is.contents {
		if v == value {
			return true
		}
		if v > value {
			return false
		}
	}
	return false
}

type Set struct {
	encoding int
	intset   *IntSet
	dict     map[string]struct{}
}

type MemorySetStorage struct {
	data map[string]*Set
	mu   sync.RWMutex
}

func NewMemorySetStorage() *MemorySetStorage {
	return &MemorySetStorage{
		data: make(map[string]*Set),
	}
}

func (s *MemorySetStorage) SAdd(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; !ok {
		s.data[key] = &Set{encoding: useIntSet, intset: NewIntSet()}
	}

	added := 0
	for _, member := range members {
		if s.data[key].encoding == useIntSet {
			if i, err := strconv.ParseInt(member, 10, 64); err == nil {
				if s.data[key].intset.Add(i) {
					added++
				}
			} else {
				s.convertToHashTable(key)
			}
		}

		if s.data[key].encoding == useHashTable {
			if _, exists := s.data[key].dict[member]; !exists {
				if s.data[key].dict == nil {
					s.data[key].dict = make(map[string]struct{})
				}
				s.data[key].dict[member] = struct{}{}
				added++
			}
		}
	}

	return added, nil
}

func (s *MemorySetStorage) SMembers(key string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if set, ok := s.data[key]; ok {
		if set.encoding == useIntSet {
			members := make([]string, len(set.intset.contents))
			for i, v := range set.intset.contents {
				members[i] = strconv.FormatInt(v, 10)
			}
			return members, nil
		}
		members := make([]string, 0, len(set.dict))
		for member := range set.dict {
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
			if set.encoding == useIntSet {
				if i, err := strconv.ParseInt(member, 10, 64); err == nil {
					if set.intset.Remove(i) {
						removed++
					}
				}
			} else {
				if _, exists := set.dict[member]; exists {
					delete(set.dict, member)
					removed++
				}
			}
		}
		if (set.encoding == useIntSet && len(set.intset.contents) == 0) ||
			(set.encoding == useHashTable && len(set.dict) == 0) {
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
		if set.encoding == useIntSet {
			return len(set.intset.contents), nil
		}
		return len(set.dict), nil
	}

	return 0, nil
}

func (s *MemorySetStorage) convertToHashTable(key string) {
	set := s.data[key]
	if set.encoding == useHashTable {
		return
	}
	set.dict = make(map[string]struct{})
	for _, i := range set.intset.contents {
		set.dict[strconv.FormatInt(i, 10)] = struct{}{}
	}
	set.encoding = useHashTable
	set.intset = nil
}
