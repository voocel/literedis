package storage

import (
	"literedis/internal/datastruct/dszset"
	"sync"
)

type MemoryZSetStorage struct {
	data map[string]dszset.ZSet
	mu   sync.RWMutex
}

func NewMemoryZSetStorage() *MemoryZSetStorage {
	return &MemoryZSetStorage{
		data: make(map[string]dszset.ZSet),
	}
}

func (s *MemoryZSetStorage) ZAdd(key string, score float64, member string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	zset, ok := s.data[key]
	if !ok {
		zset = dszset.NewZSet()
		s.data[key] = zset
	}

	if zset.Add(score, member) {
		return 1, nil
	}
	return 0, nil
}

func (s *MemoryZSetStorage) ZScore(key, member string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	zset, ok := s.data[key]
	if !ok {
		return 0, false
	}

	return zset.Score(member)
}

func (s *MemoryZSetStorage) ZRem(key string, member string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	zset, ok := s.data[key]
	if !ok {
		return 0, nil
	}

	if zset.Remove(member) {
		if zset.Len() == 0 {
			delete(s.data, key)
		}
		return 1, nil
	}
	return 0, nil
}

func (s *MemoryZSetStorage) ZRange(key string, start, stop int64) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	zset, ok := s.data[key]
	if !ok {
		return []string{}, nil
	}

	return zset.Range(start, stop), nil
}

func (s *MemoryZSetStorage) ZCard(key string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	zset, ok := s.data[key]
	if !ok {
		return 0, nil
	}

	return zset.Len(), nil
}

// 新增的方法

func (s *MemoryZSetStorage) ZRangeByScore(key string, min, max float64) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	zset, ok := s.data[key]
	if !ok {
		return []string{}, nil
	}

	return zset.RangeByScore(min, max), nil
}

func (s *MemoryZSetStorage) ZIncrBy(key string, increment float64, member string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	zset, ok := s.data[key]
	if !ok {
		zset = dszset.NewZSet()
		s.data[key] = zset
	}

	return zset.IncrBy(increment, member), nil
}
