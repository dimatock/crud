package tests

import (
	"context"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
)

func TestListUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Insert some users for testing
	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	_, err = repo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Test listing all users
	users, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestListWithFilter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Insert some users for testing
	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	_, err = repo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Test listing with a filter
	users, err := repo.List(ctx, crud.WithFilter("username", "user1"))
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
	if users[0].Username != "user1" {
		t.Errorf("Expected user1, got %s", users[0].Username)
	}
}

func TestListWithWhere(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Insert some users for testing
	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	_, err = repo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}
	_, err = repo.Create(ctx, User{Username: "user3", Email: "user3@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user3: %v", err)
	}

	// Test listing with a WHERE clause (OR condition)
	users, err := repo.List(ctx, crud.WithWhere("username = ? OR username = ?", "user1", "user3"))
	if err != nil {
		t.Fatalf("List with where failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestListWithSort(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Insert some users for testing
	_, err = repo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}
	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	// Test listing with sort
	users, err := repo.List(ctx, crud.WithSort("username", crud.SortAsc))
	if err != nil {
		t.Fatalf("List with sort failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
	if users[0].Username != "user1" {
		t.Errorf("Expected user1 as first user, got %s", users[0].Username)
	}
	if users[1].Username != "user2" {
		t.Errorf("Expected user2 as second user, got %s", users[1].Username)
	}
}

func TestListWithLimitAndOffset(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Insert some users for testing
	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	_, err = repo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}
	_, err = repo.Create(ctx, User{Username: "user3", Email: "user3@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user3: %v", err)
	}

	// Test listing with limit and offset
	users, err := repo.List(ctx, crud.WithLimit(1), crud.WithOffset(1), crud.WithSort("username", crud.SortAsc))
	if err != nil {
		t.Fatalf("List with limit and offset failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
	if users[0].Username != "user2" {
		t.Errorf("Expected user2, got %s", users[0].Username)
	}
}
