package tests

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func setupPostgresTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_DSN environment variable not set, skipping PostgreSQL tests")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to open PostgreSQL database: %v", err)
	}

	// Clean up the table before the test
	_, err = db.Exec(`DROP TABLE IF EXISTS users;`)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}

	schema := `
	CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func TestPostgresCreateWithReturning(t *testing.T) {
	db := setupPostgresTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.PostgresDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository with Postgres dialect: %v", err)
	}

	ctx := context.Background()

	newUser := User{Username: "pg-user", Email: "pg@example.com"}

	// This should use the RETURNING fast path
	createdUser, err := repo.Create(ctx, newUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if createdUser.ID == 0 {
		t.Errorf("Expected created user to have a non-zero ID from RETURNING, got 0")
	}

	if createdUser.Username != newUser.Username {
		t.Errorf("Expected username %s, got %s", newUser.Username, createdUser.Username)
	}

	// Double-check with a GetByID
	retrievedUser, err := repo.GetByID(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrievedUser.Username != newUser.Username {
		t.Errorf("Username mismatch on retrieved user: got %s, want %s", retrievedUser.Username, newUser.Username)
	}
}

func TestPostgresGetByIDWithLock(t *testing.T) {
	db := setupPostgresTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.PostgresDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()
	createdUser, err := repo.Create(ctx, User{Username: "lock-user", Email: "lock@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	txRepo := repo.WithTx(tx)

	retrievedUser, err := txRepo.GetByID(ctx, createdUser.ID, crud.WithLock("FOR UPDATE"))
	if err != nil {
		t.Fatalf("GetByID with lock failed: %v", err)
	}

	if retrievedUser.Username != createdUser.Username {
		t.Errorf("Expected username %s, got %s", createdUser.Username, retrievedUser.Username)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}
}
