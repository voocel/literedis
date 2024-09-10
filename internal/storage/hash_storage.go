package storage

import (
	"sync"
	"time"
)

type HashData struct {
	Fields map[string][]byte
	Expiry time.Time
}

type MemoryHashStorage struct {
	data map[string]*HashData
	mu   sync.RWMutex
}

func NewMemoryHashStorage() *MemoryHashStorage {
	return &MemoryHashStorage{
		data: make(map[string]*HashData),
	}
}

func (m *MemoryHashStorage) HSet(key string, fields map[string][]byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hash, ok := m.data[key]
	if !ok {
		hash = &HashData{
			Fields: make(map[string][]byte),
		}
		m.data[key] = hash
	}

	count := 0
	for field, value := range fields {
		if _, exists := hash.Fields[field]; !exists {
			count++
		}
		hash.Fields[field] = value
	}

	return count, nil
}

func (m *MemoryHashStorage) HGet(key, field string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hash, ok := m.data[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	value, ok := hash.Fields[field]
	if !ok {
		return nil, ErrKeyNotFound
	}

	return value, nil
}

func (m *MemoryHashStorage) HDel(key string, fields ...string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hash, ok := m.data[key]
	if !ok {
		return 0, nil
	}

	count := 0
	for _, field := range fields {
		if _, exists := hash.Fields[field]; exists {
			delete(hash.Fields, field)
			count++
		}
	}

	if len(hash.Fields) == 0 {
		delete(m.data, key)
	}

	return count, nil
}

func (m *MemoryHashStorage) HLen(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hash, ok := m.data[key]
	if !ok {
		return 0, nil
	}

	return len(hash.Fields), nil
}
