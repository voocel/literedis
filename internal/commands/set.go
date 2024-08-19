package commands

import "literedis/internal/storage"

type Set struct {
	data storage.Storage
}

func NewSet(data storage.Storage) *Set {
	return &Set{data: data}
}

func (s *Set) SAdd(key string, value any) {
	s.data.Set(key, value)
}
