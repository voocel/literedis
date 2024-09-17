package dshash

import (
	"literedis/internal/datastruct/base"
	"sync"
)

type Hash interface {
	HSet(field string, value string) int
	HGet(field string) (string, bool)
	HDel(fields ...string) int
	HExists(field string) bool
	HLen() int
	HKeys() []string
	HVals() []string
	HGetAll() []string
}

type hashImpl struct {
	data map[string]string
	mu   sync.RWMutex
	base.DataStructure
}

// NewHash creates and returns a new Hash
func NewHash() Hash {
	return &hashImpl{
		data: make(map[string]string),
	}
}

func (h *hashImpl) HSet(field string, value string) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	_, exists := h.data[field]
	h.data[field] = value

	if exists {
		return 0
	}
	return 1
}

func (h *hashImpl) HGet(field string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	value, exists := h.data[field]
	return value, exists
}

func (h *hashImpl) HDel(fields ...string) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	count := 0
	for _, field := range fields {
		if _, exists := h.data[field]; exists {
			delete(h.data, field)
			count++
		}
	}
	return count
}

func (h *hashImpl) HExists(field string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.data[field]
	return exists
}

func (h *hashImpl) HLen() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.data)
}

func (h *hashImpl) HKeys() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	keys := make([]string, 0, len(h.data))
	for k := range h.data {
		keys = append(keys, k)
	}
	return keys
}

func (h *hashImpl) HVals() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	vals := make([]string, 0, len(h.data))
	for _, v := range h.data {
		vals = append(vals, v)
	}
	return vals
}

func (h *hashImpl) HGetAll() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]string, 0, len(h.data)*2)
	for k, v := range h.data {
		result = append(result, k, v)
	}
	return result
}
