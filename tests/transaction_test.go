package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
)

func TestTransactionRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Non-transactional repo for verification
	baseRepo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create base repository: %v", err)
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Create a transactional repo
	txRepo := baseRepo.WithTx(tx)

	// Create a user within the transaction
	newUser := User{Username: "tx-user", Email: "tx@example.com"}
	createdUser, err := txRepo.Create(ctx, newUser)
	if err != nil {
		t.Fatalf("Create within transaction failed: %v", err)
	}

	// The user should exist within the transaction
	_, err = txRepo.GetByID(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("GetByID within transaction failed: %v", err)
	}

	// Rollback the transaction
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// The user should NOT exist outside the transaction
	_, err = baseRepo.GetByID(ctx, createdUser.ID)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows after rollback, but got %v", err)
	}
}

func TestTransactionCommit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	baseRepo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create base repository: %v", err)
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Create a transactional repo
	txRepo := baseRepo.WithTx(tx)

	// Create a user within the transaction
	newUser := User{Username: "commit-user", Email: "commit@example.com"}
	createdUser, err := txRepo.Create(ctx, newUser)
	if err != nil {
		t.Fatalf("Create within transaction failed: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// The user SHOULD exist outside the transaction
	retrievedUser, err := baseRepo.GetByID(ctx, createdUser.ID)
	if err != nil {
		t.Errorf("Expected to find user after commit, but got error %v", err)
	}

	if retrievedUser.Username != newUser.Username {
		t.Errorf("Username mismatch after commit: got %s, want %s", retrievedUser.Username, newUser.Username)
	}
}
