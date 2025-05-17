package _interface

import "context"

type TransactionFunc[T any] func(tx IRepository[T]) error

type IRepository[T any] interface {
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, entity *T) error

	CountByUUID(ctx context.Context, uuid any) (int64, error)
	FindByUUID(ctx context.Context, uuid any) (*T, error)
	FindAll(ctx context.Context) ([]T, error)
	Find(ctx context.Context, entity *T) error
	Exists(ctx context.Context, entity *T) bool
	FindByColumnValue(ctx context.Context, columnName string, value any) ([]T, error)

	BeginTx(ctx context.Context) IRepository[T]
	Commit() error
	Rollback() error
	Transaction(ctx context.Context, fn TransactionFunc[T]) error
}
