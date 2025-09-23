package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Non-transactional repo for verification
	baseRepo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create base repository")

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err, "Failed to begin transaction")

	// Create a transactional repo
	txRepo := baseRepo.WithTx(tx)

	// Create a user within the transaction
	newUser := User{Username: "tx-user", Email: "tx@example.com"}
	createdUser, err := txRepo.Create(ctx, newUser)
	require.NoError(t, err, "Create within transaction failed")

	// The user should exist within the transaction
	_, err = txRepo.GetByID(ctx, createdUser.ID)
	require.NoError(t, err, "GetByID within transaction failed")

	// Rollback the transaction
	require.NoError(t, tx.Rollback(), "Failed to rollback transaction")

	// The user should NOT exist outside the transaction
	_, err = baseRepo.GetByID(ctx, createdUser.ID)
	assert.ErrorIs(t, err, sql.ErrNoRows, "Expected sql.ErrNoRows after rollback")
}

func TestTransactionCommit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	baseRepo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create base repository")

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err, "Failed to begin transaction")

	// Create a transactional repo
	txRepo := baseRepo.WithTx(tx)

	// Create a user within the transaction
	newUser := User{Username: "commit-user", Email: "commit@example.com"}
	createdUser, err := txRepo.Create(ctx, newUser)
	require.NoError(t, err, "Create within transaction failed")

	// Commit the transaction
	require.NoError(t, tx.Commit(), "Failed to commit transaction")

	// The user SHOULD exist outside the transaction
	retrievedUser, err := baseRepo.GetByID(ctx, createdUser.ID)
	require.NoError(t, err, "Expected to find user after commit")

	assert.Equal(t, newUser.Username, retrievedUser.Username, "Username mismatch after commit")
}
