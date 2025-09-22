package tests

import (
	"context"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
)

func TestListWithIn(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Insert some users
	_, _ = repo.Create(ctx, User{Username: "user1", Email: "u1@example.com"})
	_, _ = repo.Create(ctx, User{Username: "user2", Email: "u2@example.com"})
	_, _ = repo.Create(ctx, User{Username: "user3", Email: "u3@example.com"})

	// Test WithIn
	users, err := repo.List(ctx, crud.WithIn("username", "user1", "user3"))
	if err != nil {
		t.Fatalf("List with WithIn failed: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("Expected 2 users, got %d", len(users))
	}

	// Check that we got the correct users
	foundUser1 := false
	foundUser3 := false
	for _, u := range users {
		if u.Username == "user1" {
			foundUser1 = true
		}
		if u.Username == "user3" {
			foundUser3 = true
		}
	}

	if !foundUser1 || !foundUser3 {
		t.Errorf("Did not retrieve the correct users with WithIn")
	}
}

func TestListWithIn_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Test WithIn with no values
	_, err = repo.List(ctx, crud.WithIn("username"))
	if err == nil {
		t.Fatal("Expected an error when calling WithIn with no values, but got nil")
	}

	expectedError := "WithIn option requires at least one value for column 'username'"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestListWithLike(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Insert some users
	_, _ = repo.Create(ctx, User{Username: "test-user-1", Email: "u1@example.com"})
	_, _ = repo.Create(ctx, User{Username: "test-user-2", Email: "u2@example.com"})
	_, _ = repo.Create(ctx, User{Username: "another-user", Email: "u3@example.com"})

	// Test WithLike
	users, err := repo.List(ctx, crud.WithLike("username", "test-user-%"))
	if err != nil {
		t.Fatalf("List with WithLike failed: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("Expected 2 users, got %d", len(users))
	}
}

func TestListWithOperator(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Insert some users
	u1, _ := repo.Create(ctx, User{Username: "user1", Email: "u1@example.com"})
	_, _ = repo.Create(ctx, User{Username: "user2", Email: "u2@example.com"})
	_, _ = repo.Create(ctx, User{Username: "user3", Email: "u3@example.com"})

	// Test WithOperator (e.g., >)
	users, err := repo.List(ctx, crud.WithOperator("id", ">", u1.ID))
	if err != nil {
		t.Fatalf("List with WithOperator failed: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("Expected 2 users with ID > %d, got %d", u1.ID, len(users))
	}

	// Test WithOperator (e.g., !=)
	users, err = repo.List(ctx, crud.WithOperator("username", "!=", "user2"))
	if err != nil {
		t.Fatalf("List with WithOperator failed: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("Expected 2 users, got %d", len(users))
	}
}