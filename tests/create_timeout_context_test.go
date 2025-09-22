package tests

import (
	"context"
	"testing"
	"time"

	"github.com/dimatock/crud"
)

func TestCreateWithTimeoutContext(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	newUser := User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	_, err = repo.Create(ctx, newUser)
	if err == nil {
		t.Errorf("Expected an error when creating a user with a timeout context, but got nil")
	}
}
