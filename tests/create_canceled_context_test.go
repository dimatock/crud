package tests

import (
	"context"
	"testing"

	"github.com/dimatock/crud"
	"github.com/stretchr/testify/require"
)

func TestCreateWithCanceledContext(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	newUser := User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	_, err = repo.Create(ctx, newUser)
	require.Error(t, err, "Expected an error when creating a user with a canceled context")
}
