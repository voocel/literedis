package storage

import (
	"literedis/internal/cluster"
	"sync"
	"time"
)

const DefaultDBCount = 16

type Database struct {
	stringStorage *MemoryStringStorage
	hashStorage   *MemoryHashStorage
	listStorage   *MemoryListStorage
	setStorage    *MemorySetStorage
	expiry        map[string]time.Time
}

type MemoryStorage struct {
	databases      []*Database
	currentDBIndex int
	mu             sync.RWMutex
	cluster        *cluster.Cluster
}

func NewMemoryStorage() *MemoryStorage {
	ms := &MemoryStorage{
		databases:      make([]*Database, DefaultDBCount),
		currentDBIndex: 0,
	}
	for i := 0; i < DefaultDBCount; i++ {
		ms.databases[i] = &Database{
			stringStorage: NewMemoryStringStorage(),
			hashStorage:   NewMemoryHashStorage(),
			listStorage:   NewMemoryListStorage(),
			setStorage:    NewMemorySetStorage(),
			expiry:        make(map[string]time.Time),
		}
	}
	return ms
}

func (m *MemoryStorage) GetCluster() *cluster.Cluster {
	return m.cluster
}

func (m *MemoryStorage) SetCluster(c *cluster.Cluster) {
	m.cluster = c
}

func (m *MemoryStorage) getCurrentDB() *Database {
	return m.databases[m.currentDBIndex]
}

func (m *MemoryStorage) Select(index int) error {
	if index < 0 || index >= len(m.databases) {
		return ErrInvalidDBIndex
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentDBIndex = index
	return nil
}

// ########################## String operations ##########################

func (m *MemoryStorage) Set(key string, value []byte, expiration time.Duration) error {
	db := m.getCurrentDB()
	err := db.stringStorage.Set(key, value, expiration)
	if err == nil && expiration > 0 {
		db.expiry[key] = time.Now().Add(expiration)
	}
	return err
}

func (m *MemoryStorage) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return nil, ErrKeyNotFound
	}
	return db.stringStorage.Get(key)
}

// ########################## Hash operations ##########################

func (m *MemoryStorage) HSet(key string, fields map[string][]byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
	}
	return db.hashStorage.HSet(key, fields)
}

func (m *MemoryStorage) HGet(key, field string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return nil, ErrKeyNotFound
	}
	return db.hashStorage.HGet(key, field)
}

func (m *MemoryStorage) HDel(key string, fields ...string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return 0, nil
	}
	return db.hashStorage.HDel(key, fields...)
}

func (m *MemoryStorage) HLen(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return 0, nil
	}
	return db.hashStorage.HLen(key)
}

// ########################## List operations ##########################

func (m *MemoryStorage) LPush(key string, values ...[]byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
	}
	return db.listStorage.LPush(key, values...)
}

func (m *MemoryStorage) RPush(key string, values ...[]byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
	}
	return db.listStorage.RPush(key, values...)
}

func (m *MemoryStorage) LPop(key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return nil, ErrKeyNotFound
	}
	return db.listStorage.LPop(key)
}

func (m *MemoryStorage) RPop(key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return nil, ErrKeyNotFound
	}
	return db.listStorage.RPop(key)
}

func (m *MemoryStorage) LLen(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return 0, nil
	}
	return db.listStorage.LLen(key)
}

func (m *MemoryStorage) LRange(key string, start, stop int) ([][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return [][]byte{}, nil
	}
	return db.listStorage.LRange(key, start, stop)
}

// ########################## Set operations ##########################

func (m *MemoryStorage) SAdd(key string, members ...string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
	}
	return db.setStorage.SAdd(key, members...)
}

func (m *MemoryStorage) SMembers(key string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return []string{}, nil
	}
	return db.setStorage.SMembers(key)
}

func (m *MemoryStorage) SRem(key string, members ...string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return 0, nil
	}
	return db.setStorage.SRem(key, members...)
}

func (m *MemoryStorage) SCard(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return 0, nil
	}
	return db.setStorage.SCard(key)
}

// ########################## Generic operations ##########################

func (m *MemoryStorage) Del(key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	deleted := false
	if _, ok := db.stringStorage.data[key]; ok {
		delete(db.stringStorage.data, key)
		deleted = true
	}
	if _, ok := db.hashStorage.data[key]; ok {
		delete(db.hashStorage.data, key)
		deleted = true
	}
	if _, ok := db.listStorage.data[key]; ok {
		delete(db.listStorage.data, key)
		deleted = true
	}
	delete(db.expiry, key)
	return deleted, nil
}

func (m *MemoryStorage) Exists(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return false
	}
	_, existsString := db.stringStorage.data[key]
	_, existsHash := db.hashStorage.data[key]
	_, existsList := db.listStorage.data[key]
	return existsString || existsHash || existsList
}

func (m *MemoryStorage) Expire(key string, expiration time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if _, ok := db.stringStorage.data[key]; ok {
		db.expiry[key] = time.Now().Add(expiration)
		return true, nil
	}
	if _, ok := db.hashStorage.data[key]; ok {
		db.expiry[key] = time.Now().Add(expiration)
		return true, nil
	}
	if _, ok := db.listStorage.data[key]; ok {
		db.expiry[key] = time.Now().Add(expiration)
		return true, nil
	}
	return false, nil
}

func (m *MemoryStorage) TTL(key string) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if expiry, ok := db.expiry[key]; ok {
		ttl := time.Until(expiry)
		if ttl < 0 {
			m.deleteKey(db, key)
			return -2, nil // -2 表示键不存在
		}
		return ttl, nil
	}
	if m.Exists(key) {
		return -1, nil // -1 表示键没有过期时间
	}
	return -2, nil // -2 表示键不存在
}

func (m *MemoryStorage) Type(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return "", ErrKeyNotFound
	}
	if _, ok := db.stringStorage.data[key]; ok {
		return "string", nil
	}
	if _, ok := db.hashStorage.data[key]; ok {
		return "hash", nil
	}
	if _, ok := db.listStorage.data[key]; ok {
		return "list", nil
	}
	return "", ErrKeyNotFound
}

func (m *MemoryStorage) Flush() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.databases {
		m.databases[i] = &Database{
			stringStorage: NewMemoryStringStorage(),
			hashStorage:   NewMemoryHashStorage(),
			listStorage:   NewMemoryListStorage(),
			expiry:        make(map[string]time.Time),
		}
	}
	return nil
}

func (m *MemoryStorage) FlushDB() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.databases[m.currentDBIndex] = &Database{
		stringStorage: NewMemoryStringStorage(),
		hashStorage:   NewMemoryHashStorage(),
		listStorage:   NewMemoryListStorage(),
		expiry:        make(map[string]time.Time),
	}
	return nil
}

func (m *MemoryStorage) isExpired(db *Database, key string) bool {
	if expiry, ok := db.expiry[key]; ok {
		if time.Now().After(expiry) {
			return true
		}
	}
	return false
}

func (m *MemoryStorage) deleteKey(db *Database, key string) {
	delete(db.stringStorage.data, key)
	delete(db.hashStorage.data, key)
	delete(db.listStorage.data, key)
	delete(db.setStorage.data, key)
	delete(db.expiry, key)
}
