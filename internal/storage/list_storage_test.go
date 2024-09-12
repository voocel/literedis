package storage

import (
	"literedis/internal/consts"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMemoryStorage_LPush tests the LPush operation
func TestMemoryStorage_LPush(t *testing.T) {
	storage := NewMemoryStorage()

	// Test pushing to a new list
	length, err := storage.LPush("key1", []byte("value1"), []byte("value2"))
	assert.NoError(t, err)
	assert.Equal(t, int64(2), length)

	// Test pushing to an existing list
	length, err = storage.LPush("key1", []byte("value3"))
	assert.NoError(t, err)
	assert.Equal(t, int64(3), length)
}

// TestMemoryStorage_RPush tests the RPush operation
func TestMemoryStorage_RPush(t *testing.T) {
	storage := NewMemoryStorage()

	// Test pushing to a new list
	length, err := storage.RPush("key1", []byte("value1"), []byte("value2"))
	assert.NoError(t, err)
	assert.Equal(t, int64(2), length)

	// Test pushing to an existing list
	length, err = storage.RPush("key1", []byte("value3"))
	assert.NoError(t, err)
	assert.Equal(t, int64(3), length)
}

// TestMemoryStorage_LPop tests the LPop operation
func TestMemoryStorage_LPop(t *testing.T) {
	storage := NewMemoryStorage()

	// Setup test data
	_, _ = storage.RPush("key1", []byte("value1"), []byte("value2"), []byte("value3"))

	// Test popping from the list
	value, err := storage.LPop("key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value1"), value)

	// Test popping from a non-existent key
	_, err = storage.LPop("non-existent")
	assert.Equal(t, consts.ErrKeyNotFound, err)

	// Test popping until the list is empty
	_, _ = storage.LPop("key1")
	_, _ = storage.LPop("key1")
	_, err = storage.LPop("key1")
	assert.Equal(t, consts.ErrKeyNotFound, err)
}

// TestMemoryStorage_RPop tests the RPop operation
func TestMemoryStorage_RPop(t *testing.T) {
	storage := NewMemoryStorage()

	// Setup test data
	_, _ = storage.RPush("key1", []byte("value1"), []byte("value2"), []byte("value3"))

	// Test popping from the list
	value, err := storage.RPop("key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value3"), value)

	// Test popping from a non-existent key
	_, err = storage.RPop("non-existent")
	assert.Equal(t, consts.ErrKeyNotFound, err)

	// Test popping until the list is empty
	_, _ = storage.RPop("key1")
	_, _ = storage.RPop("key1")
	_, err = storage.RPop("key1")
	assert.Equal(t, consts.ErrKeyNotFound, err)
}

// TestMemoryStorage_LRange tests the LRange operation
func TestMemoryStorageList_LRange(t *testing.T) {
	storage := NewMemoryStorage()

	// Setup test data
	_, _ = storage.RPush("key1", []byte("value1"), []byte("value2"), []byte("value3"), []byte("value4"), []byte("value5"))

	// Test normal range
	values, err := storage.LRange("key1", 1, 3)
	assert.NoError(t, err)
	assert.Equal(t, [][]byte{[]byte("value2"), []byte("value3"), []byte("value4")}, values)

	// Test range with negative indices
	values, err = storage.LRange("key1", -3, -1)
	assert.NoError(t, err)
	assert.Equal(t, [][]byte{[]byte("value3"), []byte("value4"), []byte("value5")}, values)

	// Test range on non-existent key
	_, err = storage.LRange("non-existent", 0, -1)
	assert.Equal(t, consts.ErrKeyNotFound, err)
}

// TestMemoryStorage_LIndex tests the LIndex operation
func TestMemoryStorage_LIndex(t *testing.T) {
	storage := NewMemoryStorage()

	// Setup test data
	_, _ = storage.RPush("key1", []byte("value1"), []byte("value2"), []byte("value3"))

	// Test getting an element by index
	value, err := storage.LIndex("key1", 1)
	assert.NoError(t, err)
	assert.Equal(t, []byte("value2"), value)

	// Test getting an element with negative index
	value, err = storage.LIndex("key1", -1)
	assert.NoError(t, err)
	assert.Equal(t, []byte("value3"), value)

	// Test getting an element from a non-existent key
	_, err = storage.LIndex("non-existent", 0)
	assert.Equal(t, consts.ErrKeyNotFound, err)

	// Test getting an element with out of range index
	_, err = storage.LIndex("key1", 10)
	assert.Equal(t, consts.ErrKeyNotFound, err)
}

// TestMemoryStorage_LSet tests the LSet operation
func TestMemoryStorage_LSet(t *testing.T) {
	storage := NewMemoryStorage()

	// Setup test data
	_, _ = storage.RPush("key1", []byte("value1"), []byte("value2"), []byte("value3"))

	// Test setting an element by index
	err := storage.LSet("key1", 1, []byte("new-value"))
	assert.NoError(t, err)

	// Verify the change
	value, _ := storage.LIndex("key1", 1)
	assert.Equal(t, []byte("new-value"), value)

	// Test setting an element with negative index
	err = storage.LSet("key1", -1, []byte("another-new-value"))
	assert.NoError(t, err)

	// Verify the change
	value, _ = storage.LIndex("key1", -1)
	assert.Equal(t, []byte("another-new-value"), value)

	// Test setting an element on a non-existent key
	err = storage.LSet("non-existent", 0, []byte("value"))
	assert.Equal(t, consts.ErrKeyNotFound, err)

	// Test setting an element with out of range index
	err = storage.LSet("key1", 10, []byte("value"))
	assert.Equal(t, consts.ErrIndexOutOfRange, err)
}

// TestMemoryStorage_ListExpiration tests the expiration of list keys
func TestMemoryStorage_ListExpiration(t *testing.T) {
	storage := NewMemoryStorage()

	// Setup test data
	_, _ = storage.RPush("key1", []byte("value1"), []byte("value2"))
	_, _ = storage.Expire("key1", time.Millisecond*100)

	// Wait for the key to expire
	time.Sleep(time.Millisecond * 150)

	// Try to access the expired key
	_, err := storage.LPop("key1")
	assert.Equal(t, consts.ErrKeyNotFound, err)

	// Try to push to the expired key (should create a new list)
	length, err := storage.RPush("key1", []byte("new-value"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1), length)

	// Verify the new value
	value, err := storage.LPop("key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("new-value"), value)
}
