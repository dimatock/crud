package tests

import (
	"context"
	"testing"

	"github.com/dimatock/crud"
	"github.com/stretchr/testify/require"
)

func TestCreateWithNilContext(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

	newUser := User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	// We are not testing for a panic here because the behavior of the database/sql
	// package with a nil context is not guaranteed. Some drivers might panic,
	// while others might deadlock. Instead, we follow the linter's advice and
	// use context.TODO() when we don't have a specific context to pass.
	_, err = repo.Create(context.TODO(), newUser)
	require.NoError(t, err, "Expected no error when creating a user with context.TODO()")
}
