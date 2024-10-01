package consts

import "errors"

var (
	// General errors
	ErrInvalidArgument = errors.New("invalid argument")
	ErrUnknownCommand  = errors.New("unknown command")
	ErrSyntaxError     = errors.New("syntax error")
	ErrWrongType       = errors.New("operation against a key holding the wrong kind of value")

	// Key-value operation errors
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExists   = errors.New("key already exists")
	ErrKeyExpired  = errors.New("key has expired")

	// Data structure related errors
	ErrIndexOutOfRange = errors.New("index out of range")
	ErrNoSuchField     = errors.New("field does not exist")
	ErrNotInteger      = errors.New("value is not an integer")
	ErrNotFloat        = errors.New("value is not a valid float")

	// Cluster related errors
	ErrClusterNotEnabled = errors.New("cluster mode is not enabled")
	ErrWrongNode         = errors.New("wrong node for the given key")

	// Storage related errors
	ErrDBIndexOutOfRange = errors.New("database index is out of range")
	ErrMaxMemoryReached  = errors.New("maximum memory limit reached")

	// Other errors
	ErrOperationAborted = errors.New("operation aborted")
	ErrBackgroundSaving = errors.New("background save already in progress")
	ErrAuthRequired     = errors.New("authentication required")
	ErrAuthFailed       = errors.New("authentication failed")
)
