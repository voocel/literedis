package dsset

import (
	"hash/fnv"
	"strconv"
	"sync"
	"time"
)

const (
	shardCount     = 32
	bitmapMaxValue = 1 << 16 // Use 2^16 as the maximum value for the bitmap
)

// OptimizedSet represents an optimized set data structure that can use different
// internal representations based on the nature of its elements.
type OptimizedSet struct {
	encoding int
	intset   *IntSet
	bitmap   []uint64
	shards   [shardCount]map[string]struct{}
	locks    [shardCount]sync.RWMutex
	expireAt time.Time
	mu       sync.RWMutex
}

func NewOptimizedSet() *OptimizedSet {
	s := &OptimizedSet{
		encoding: useIntSet,
		intset:   NewIntSet(),
	}
	for i := 0; i < shardCount; i++ {
		s.shards[i] = make(map[string]struct{})
	}
	return s
}

// Add adds one or more members to the set
// It automatically chooses and converts to the most appropriate internal representation
func (s *OptimizedSet) Add(members ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	added := 0
	for _, member := range members {
		if s.encoding == useIntSet {
			// Try to add to intset if possible
			if i, err := strconv.ParseInt(member, 10, 64); err == nil {
				if s.intset.Add(i) {
					added++
				}
				continue
			}
			// If not possible, convert to hash table
			s.convertToHashTable()
		}

		if s.encoding == useBitmap {
			// Try to add to bitmap if possible
			if i, err := strconv.ParseInt(member, 10, 64); err == nil && i >= 0 && i < bitmapMaxValue {
				index, bit := i/64, uint(i%64)
				if s.bitmap[index]&(1<<bit) == 0 {
					s.bitmap[index] |= 1 << bit
					added++
				}
				continue
			}
			// If not possible, convert to hash table
			s.convertToHashTable()
		}

		// Add to hash table
		shard := s.getShard(member)
		s.locks[shard].Lock()
		if _, exists := s.shards[shard][member]; !exists {
			s.shards[shard][member] = struct{}{}
			added++
		}
		s.locks[shard].Unlock()
	}
	return added
}

func (s *OptimizedSet) Remove(members ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := 0
	for _, member := range members {
		if s.encoding == useIntSet {
			if i, err := strconv.ParseInt(member, 10, 64); err == nil {
				if s.intset.Remove(i) {
					removed++
				}
				continue
			}
		}

		if s.encoding == useBitmap {
			if i, err := strconv.ParseInt(member, 10, 64); err == nil && i >= 0 && i < bitmapMaxValue {
				index, bit := i/64, uint(i%64)
				if s.bitmap[index]&(1<<bit) != 0 {
					s.bitmap[index] &^= 1 << bit
					removed++
				}
				continue
			}
		}

		shard := s.getShard(member)
		s.locks[shard].Lock()
		if _, exists := s.shards[shard][member]; exists {
			delete(s.shards[shard], member)
			removed++
		}
		s.locks[shard].Unlock()
	}
	return removed
}

func (s *OptimizedSet) IsMember(member string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.encoding == useIntSet {
		if i, err := strconv.ParseInt(member, 10, 64); err == nil {
			return s.intset.Contains(i)
		}
		return false
	}

	if s.encoding == useBitmap {
		if i, err := strconv.ParseInt(member, 10, 64); err == nil && i >= 0 && i < bitmapMaxValue {
			index, bit := i/64, uint(i%64)
			return s.bitmap[index]&(1<<bit) != 0
		}
		return false
	}

	shard := s.getShard(member)
	s.locks[shard].RLock()
	_, exists := s.shards[shard][member]
	s.locks[shard].RUnlock()
	return exists
}

func (s *OptimizedSet) Members() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var members []string
	if s.encoding == useIntSet {
		ints := s.intset.ToSlice()
		members = make([]string, len(ints))
		for i, v := range ints {
			members[i] = strconv.FormatInt(v, 10)
		}
	} else if s.encoding == useBitmap {
		for i, v := range s.bitmap {
			for bit := uint(0); bit < 64; bit++ {
				if v&(1<<bit) != 0 {
					members = append(members, strconv.FormatInt(int64(i*64+int(bit)), 10))
				}
			}
		}
	} else {
		for i := 0; i < shardCount; i++ {
			s.locks[i].RLock()
			for member := range s.shards[i] {
				members = append(members, member)
			}
			s.locks[i].RUnlock()
		}
	}
	return members
}

func (s *OptimizedSet) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.encoding == useIntSet {
		return s.intset.Len()
	}

	if s.encoding == useBitmap {
		count := 0
		for _, v := range s.bitmap {
			count += popcount(v)
		}
		return count
	}

	count := 0
	for i := 0; i < shardCount; i++ {
		s.locks[i].RLock()
		count += len(s.shards[i])
		s.locks[i].RUnlock()
	}
	return count
}

func (s *OptimizedSet) SetExpire(expireAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireAt = expireAt
}

func (s *OptimizedSet) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return !s.expireAt.IsZero() && time.Now().After(s.expireAt)
}

// convertToHashTable converts the set from intset or bitmap to hash table representation
func (s *OptimizedSet) convertToHashTable() {
	if s.encoding == useIntSet {
		for _, i := range s.intset.ToSlice() {
			member := strconv.FormatInt(i, 10)
			shard := s.getShard(member)
			s.shards[shard][member] = struct{}{}
		}
		s.intset = nil
	} else if s.encoding == useBitmap {
		for i, v := range s.bitmap {
			for bit := uint(0); bit < 64; bit++ {
				if v&(1<<bit) != 0 {
					member := strconv.FormatInt(int64(i*64+int(bit)), 10)
					shard := s.getShard(member)
					s.shards[shard][member] = struct{}{}
				}
			}
		}
		s.bitmap = nil
	}
	s.encoding = useHashTable
}

// getShard returns the shard index for a given key
func (s *OptimizedSet) getShard(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32()) % shardCount
}

func popcount(x uint64) int {
	return int(uint64(len64)*uint64(x>>uint(len64-1)) +
		uint64(len64tab[x&uint64(len64-1)]))
}

var len64 = 64
var len64tab = [64]byte{
	0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4,
	1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
	1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
}

// Union returns a new set that is the union of s and other
func (s *OptimizedSet) Union(other Set) Set {
	result := NewOptimizedSet()
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, member := range s.Members() {
		result.Add(member)
	}
	for _, member := range other.Members() {
		result.Add(member)
	}
	return result
}

// Intersection returns a new set that is the intersection of s and other
func (s *OptimizedSet) Intersection(other Set) Set {
	result := NewOptimizedSet()
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, member := range s.Members() {
		if other.IsMember(member) {
			result.Add(member)
		}
	}
	return result
}

// Difference returns a new set that is the difference of s and other
func (s *OptimizedSet) Difference(other Set) Set {
	result := NewOptimizedSet()
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, member := range s.Members() {
		if !other.IsMember(member) {
			result.Add(member)
		}
	}
	return result
}
