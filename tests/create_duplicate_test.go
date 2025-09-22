package tests

import (
	"context"
	"testing"

	"github.com/dimatock/crud"
)

func TestCreateDuplicateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Create a user
	newUser := User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	_, err = repo.Create(ctx, newUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Try to create another user with the same username
	duplicateUser := User{
		Username: "testuser",
		Email:    "test2@example.com",
	}

	_, err = repo.Create(ctx, duplicateUser)
	if err == nil {
		t.Errorf("Expected an error when creating a user with a duplicate username, but got nil")
	}
}
