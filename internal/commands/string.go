package commands

import "literedis/internal/storage"

type String struct {
	data storage.Storage
}

func NewString(data storage.Storage) *String {
	return &String{data: data}
}

func (s *String) Set(key string, value any) {
	s.data.Set(key, value)
}

func (s *String) Get(key string) (any, bool) {
	return s.data.Get(key)
}

func (s *String) Del(key string) {
	s.data.Del(key)
}
