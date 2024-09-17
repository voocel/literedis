package storage

import (
	"literedis/internal/datastruct/dshash"
	"sync"
)

type MemoryHashStorage struct {
	data map[string]dshash.Hash
	mu   sync.RWMutex
}

func NewMemoryHashStorage() *MemoryHashStorage {
	return &MemoryHashStorage{
		data: make(map[string]dshash.Hash),
	}
}

func (m *MemoryHashStorage) HSet(key string, fields map[string][]byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hash, ok := m.data[key]
	if !ok {
		hash = dshash.NewHash()
		m.data[key] = hash
	}

	count := 0
	for field, value := range fields {
		if hash.HSet(field, string(value)) == 1 {
			count++
		}
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

	value, exists := hash.HGet(field)
	if !exists {
		return nil, ErrKeyNotFound
	}

	return []byte(value), nil
}

func (m *MemoryHashStorage) HDel(key string, fields ...string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hash, ok := m.data[key]
	if !ok {
		return 0, nil
	}

	count := hash.HDel(fields...)
	if hash.HLen() == 0 {
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

	return hash.HLen(), nil
}

func (m *MemoryHashStorage) HExists(key, field string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hash, ok := m.data[key]
	if !ok {
		return false, nil
	}

	return hash.HExists(field), nil
}

func (m *MemoryHashStorage) HKeys(key string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hash, ok := m.data[key]
	if !ok {
		return nil, nil
	}

	return hash.HKeys(), nil
}

func (m *MemoryHashStorage) HVals(key string) ([][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hash, ok := m.data[key]
	if !ok {
		return nil, nil
	}

	values := hash.HVals()
	byteValues := make([][]byte, len(values))
	for i, v := range values {
		byteValues[i] = []byte(v)
	}

	return byteValues, nil
}

func (m *MemoryHashStorage) HGetAll(key string) (map[string][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hash, ok := m.data[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	all := hash.HGetAll()
	result := make(map[string][]byte, len(all))
	for i := 0; i < len(all); i += 2 {
		result[all[i]] = []byte(all[i+1])
	}

	return result, nil
}
