package ports

import "context"

var ErrInvalidKey = errorString("redis key is invalid")

type errorString string

func (e errorString) Error() string { return string(e) }

type RedisPort interface {
	Get(context.Context, string) (string, string, error)
	Set(context.Context, string, string) (string, error)
	HGet(context.Context, string, string) (string, string, error)
	HSet(context.Context, string, string, string) (string, error)
	Close() error
}
