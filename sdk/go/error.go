package literedis

import (
	"errors"
)

type RedisError struct {
	Err error
}

func (e *RedisError) Error() string {
	return e.Err.Error()
}

func NewRedisError(msg string) *RedisError {
	return &RedisError{Err: errors.New(msg)}
}

func WrapRedisError(err error, msg string) *RedisError {
	return &RedisError{Err: errors.New(msg + ": " + err.Error())}
}
