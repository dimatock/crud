package tests

import (
	"context"
	"testing"

	"github.com/dimatock/crud"
)

func TestCreateWithCanceledContext(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	newUser := User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	_, err = repo.Create(ctx, newUser)
	if err == nil {
		t.Errorf("Expected an error when creating a user with a canceled context, but got nil")
	}
}
