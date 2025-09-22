package crud

import (
	"context"
	"database/sql"
)

// RepositoryInterface defines the interface for a generic CRUD repository.
type RepositoryInterface[T any] interface {
	Create(ctx context.Context, item T) (T, error)
	CreateOrUpdate(ctx context.Context, item T) (T, error)
	GetByID(ctx context.Context, id any, opts ...Option) (T, error)
	Update(ctx context.Context, item T) (T, error)
	Delete(ctx context.Context, id any) error
	List(ctx context.Context, opts ...Option) ([]T, error)
	WithTx(tx *sql.Tx) RepositoryInterface[T]
}
