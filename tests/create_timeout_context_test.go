package tests

import (
	"context"
	"testing"
	"time"

	"github.com/dimatock/crud"
	"github.com/stretchr/testify/require"
)

func TestCreateWithTimeoutContext(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	newUser := User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	_, err = repo.Create(ctx, newUser)
	require.Error(t, err, "Expected an error when creating a user with a timeout context")
}
