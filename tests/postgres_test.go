package tests

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPostgresTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_DSN environment variable not set, skipping PostgreSQL tests")
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "Failed to open PostgreSQL database")

	// Clean up the table before the test
	_, err = db.Exec(`DROP TABLE IF EXISTS users;`)
	require.NoError(t, err, "Failed to drop table")

	schema := `
	CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err, "Failed to create table")

	return db
}

func TestPostgresCreateWithReturning(t *testing.T) {
	db := setupPostgresTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.PostgresDialect{})
	require.NoError(t, err, "Failed to create repository with Postgres dialect")

	ctx := context.Background()

	newUser := User{Username: "pg-user", Email: "pg@example.com"}

	// This should use the RETURNING fast path
	createdUser, err := repo.Create(ctx, newUser)
	require.NoError(t, err, "Create failed")

	assert.NotEqual(t, 0, createdUser.ID, "Expected created user to have a non-zero ID from RETURNING")
	assert.Equal(t, newUser.Username, createdUser.Username)

	// Double-check with a GetByID
	retrievedUser, err := repo.GetByID(ctx, createdUser.ID)
	require.NoError(t, err, "GetByID failed")
	assert.Equal(t, newUser.Username, retrievedUser.Username)
}

func TestPostgresGetByIDWithLock(t *testing.T) {
	db := setupPostgresTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.PostgresDialect{})
	require.NoError(t, err)

	ctx := context.Background()
	createdUser, err := repo.Create(ctx, User{Username: "lock-user", Email: "lock@example.com"})
	require.NoError(t, err)

	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()

	txRepo := repo.WithTx(tx)

	retrievedUser, err := txRepo.GetByID(ctx, createdUser.ID, crud.WithLock[User]("FOR UPDATE"))
	require.NoError(t, err)

	assert.Equal(t, createdUser.Username, retrievedUser.Username)

	require.NoError(t, tx.Commit())
}
