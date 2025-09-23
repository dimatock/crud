package crud

import (
	"context"
	"database/sql"
)

// RepositoryInterface defines the interface for a generic CRUD repository.
type RepositoryInterface[T any] interface {
	// WithTx returns a new repository instance that will run queries within the given transaction.
	WithTx(tx *sql.Tx) RepositoryInterface[T]

	// Create inserts a new record into the database.
	Create(ctx context.Context, item T) (T, error)

	// CreateOrUpdate inserts a new record or updates it if it already exists.
	CreateOrUpdate(ctx context.Context, item T) (T, error)

	// GetByID retrieves a single record by its primary key.
	GetByID(ctx context.Context, id any, opts ...Option[T]) (T, error)

	// List retrieves a slice of records based on the provided options.
	List(ctx context.Context, opts ...Option[T]) ([]T, error)

	// Update modifies an existing record.
	Update(ctx context.Context, item T) (T, error)

	// Delete removes a record from the database by its primary key.
	Delete(ctx context.Context, id any) error
}
