package tests

import (
	"context"
	"testing"

	"github.com/dimatock/crud"
	"github.com/stretchr/testify/require"
)

func TestCreateDuplicateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

	ctx := context.Background()

	// Create a user
	newUser := User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	_, err = repo.Create(ctx, newUser)
	require.NoError(t, err, "Create failed")

	// Try to create another user with the same username
	duplicateUser := User{
		Username: "testuser",
		Email:    "test2@example.com",
	}

	_, err = repo.Create(ctx, duplicateUser)
	require.Error(t, err, "Expected an error when creating a user with a duplicate username")
}
