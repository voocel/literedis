package storage

import (
	"literedis/internal/consts"
	"literedis/internal/datastruct/dsstring"
	"sync"
)

type MemoryStringStorage struct {
	data map[string]*dsstring.SDS
	mu   sync.RWMutex
}

func NewMemoryStringStorage() *MemoryStringStorage {
	return &MemoryStringStorage{
		data: make(map[string]*dsstring.SDS),
	}
}

func (s *MemoryStringStorage) Set(key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = dsstring.NewSDS(string(value))
	return nil
}

func (s *MemoryStringStorage) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	return value.Bytes(), nil
}

func (m *MemoryStringStorage) Del(key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.data[key]
	delete(m.data, key)
	return exists, nil
}

func (m *MemoryStringStorage) Exists(key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.data[key]
	return exists, nil
}

func (m *MemoryStringStorage) Append(key string, value []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sds, exists := m.data[key]
	if !exists {
		sds = dsstring.NewSDS("")
		m.data[key] = sds
	}

	return sds.Append(value), nil
}

func (m *MemoryStringStorage) GetRange(key string, start, end int) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sds, exists := m.data[key]
	if !exists {
		return nil, consts.ErrKeyNotFound
	}

	return sds.GetRange(start, end), nil
}

func (m *MemoryStringStorage) SetRange(key string, offset int, value []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sds, exists := m.data[key]
	if !exists {
		sds = dsstring.NewSDS("")
		m.data[key] = sds
	}

	return sds.SetRange(offset, value), nil
}

func (m *MemoryStringStorage) StrLen(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sds, exists := m.data[key]
	if !exists {
		return 0, consts.ErrKeyNotFound
	}

	return int(sds.Len()), nil
}
