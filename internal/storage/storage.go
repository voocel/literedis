package storage

type Storage interface {
	Get(key string) (any, bool)
	Set(key string, value any)
	Del(key string)
	Size() int
	Keys() []string
	Flush()
}
