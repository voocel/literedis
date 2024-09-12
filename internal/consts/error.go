package consts

import "errors"

var (
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrUnknownCommand    = errors.New("unknown command")
	ErrWrongNode         = errors.New("wrong node for the given key")
	ErrKeyNotFound       = errors.New("key not found")
	ErrClusterNotEnabled = errors.New("cluster mode is not enabled")
	ErrKeyExists         = errors.New("key already exists")
	ErrIndexOutOfRange   = errors.New("invalid database index")
)
