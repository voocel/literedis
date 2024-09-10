package storage

import (
	"sync"
	"time"
)

type MemoryStorage struct {
	stringStorage *MemoryStringStorage
	hashStorage   *MemoryHashStorage
	listStorage   *MemoryListStorage
	expiry        map[string]time.Time
	mu            sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		stringStorage: NewMemoryStringStorage(),
		hashStorage:   NewMemoryHashStorage(),
		listStorage:   NewMemoryListStorage(),
		expiry:        make(map[string]time.Time),
	}
}

// ########################## String operations ##########################

func (m *MemoryStorage) Set(key string, value []byte, expiration time.Duration) error {
	err := m.stringStorage.Set(key, value, expiration)
	if err == nil && expiration > 0 {
		m.setExpiry(key, time.Now().Add(expiration))
	}
	return err
}

func (m *MemoryStorage) Get(key string) ([]byte, error) {
	if m.isExpired(key) {
		m.Del(key)
		return nil, ErrKeyNotFound
	}
	return m.stringStorage.Get(key)
}

// Hash operations
func (m *MemoryStorage) HSet(key string, fields map[string][]byte) (int, error) {
	if m.isExpired(key) {
		m.Del(key)
	}
	return m.hashStorage.HSet(key, fields)
}

func (m *MemoryStorage) HGet(key, field string) ([]byte, error) {
	if m.isExpired(key) {
		m.Del(key)
		return nil, ErrKeyNotFound
	}
	return m.hashStorage.HGet(key, field)
}

func (m *MemoryStorage) HDel(key string, fields ...string) (int, error) {
	if m.isExpired(key) {
		m.Del(key)
		return 0, nil
	}
	return m.hashStorage.HDel(key, fields...)
}

func (m *MemoryStorage) HLen(key string) (int, error) {
	if m.isExpired(key) {
		m.Del(key)
		return 0, nil
	}
	return m.hashStorage.HLen(key)
}

// ########################## List operations ##########################

func (m *MemoryStorage) LPush(key string, values ...[]byte) (int, error) {
	if m.isExpired(key) {
		m.Del(key)
	}
	return m.listStorage.LPush(key, values...)
}

func (m *MemoryStorage) RPush(key string, values ...[]byte) (int, error) {
	if m.isExpired(key) {
		m.Del(key)
	}
	return m.listStorage.RPush(key, values...)
}

func (m *MemoryStorage) LPop(key string) ([]byte, error) {
	if m.isExpired(key) {
		m.Del(key)
		return nil, ErrKeyNotFound
	}
	return m.listStorage.LPop(key)
}

func (m *MemoryStorage) RPop(key string) ([]byte, error) {
	if m.isExpired(key) {
		m.Del(key)
		return nil, ErrKeyNotFound
	}
	return m.listStorage.RPop(key)
}

func (m *MemoryStorage) LLen(key string) (int, error) {
	if m.isExpired(key) {
		m.Del(key)
		return 0, nil
	}
	return m.listStorage.LLen(key)
}

func (m *MemoryStorage) LRange(key string, start, stop int) ([][]byte, error) {
	if m.isExpired(key) {
		m.Del(key)
		return [][]byte{}, nil
	}
	return m.listStorage.LRange(key, start, stop)
}

// ########################## Generic operations ##########################

func (m *MemoryStorage) Del(key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, existsString := m.stringStorage.data[key]
	_, existsHash := m.hashStorage.data[key]
	_, existsList := m.listStorage.data[key]

	delete(m.stringStorage.data, key)
	delete(m.hashStorage.data, key)
	delete(m.listStorage.data, key)
	delete(m.expiry, key)

	return existsString || existsHash || existsList, nil
}

func (m *MemoryStorage) Exists(key string) (bool, error) {
	if m.isExpired(key) {
		m.Del(key)
		return false, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	_, existsString := m.stringStorage.data[key]
	_, existsHash := m.hashStorage.data[key]
	_, existsList := m.listStorage.data[key]

	return existsString || existsHash || existsList, nil
}

func (m *MemoryStorage) Expire(key string, expiration time.Duration) (bool, error) {
	if exists, _ := m.Exists(key); !exists {
		return false, nil
	}

	m.setExpiry(key, time.Now().Add(expiration))
	return true, nil
}

func (m *MemoryStorage) TTL(key string) (time.Duration, error) {
	if ttl, err := m.stringStorage.TTL(key); err == nil {
		return ttl, nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if expiry, ok := m.expiry[key]; ok {
		ttl := time.Until(expiry)
		if ttl < 0 {
			delete(m.expiry, key)
			return -2, nil
		}
		return ttl, nil
	}
	return -1, nil
}

func (m *MemoryStorage) Type(key string) (string, error) {
	if m.isExpired(key) {
		m.Del(key)
		return "", ErrKeyNotFound
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.stringStorage.data[key]; ok {
		return "string", nil
	}
	if _, ok := m.hashStorage.data[key]; ok {
		return "hash", nil
	}
	if _, ok := m.listStorage.data[key]; ok {
		return "list", nil
	}
	return "", ErrKeyNotFound
}

func (m *MemoryStorage) Flush() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stringStorage = NewMemoryStringStorage()
	m.hashStorage = NewMemoryHashStorage()
	m.listStorage = NewMemoryListStorage()
	m.expiry = make(map[string]time.Time)

	return nil
}

func (m *MemoryStorage) setExpiry(key string, expiry time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expiry[key] = expiry
}

func (m *MemoryStorage) isExpired(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if expiry, ok := m.expiry[key]; ok {
		if time.Now().After(expiry) {
			return true
		}
	}
	return false
}
