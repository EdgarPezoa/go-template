package ports

import (
	"context"
	"errors"
)

var (
	ErrInvalidID = errors.New("database id is invalid")
	ErrNotFound  = errors.New("resource not found")
)

type DatabasePort interface {
	DeleteOne(context.Context, string, any) (DeleteResult, error)
	DeleteMany(context.Context, string, any) (DeleteResult, error)
	FindOne(context.Context, string, any, any) error
	Find(context.Context, string, any, any) error
	Insert(context.Context, string, any) (string, error)
	Upsert(context.Context, string, any, any) (UpsertResult, error)
}

type DeleteResult struct {
	DeletedCount int64
}

type UpsertResult struct {
	MatchedCount  int64
	ModifiedCount int64
	UpsertedID    string
}

type IDFilter struct {
	ID string
}
