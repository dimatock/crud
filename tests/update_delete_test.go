package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

	ctx := context.Background()

	// Create a user to update
	newUser := User{Username: "testuser", Email: "test@example.com"}
	createdUser, err := repo.Create(ctx, newUser)
	require.NoError(t, err, "Create failed")

	// Update the user's email
	createdUser.Email = "updated@example.com"
	updatedUser, err := repo.Update(ctx, createdUser)
	require.NoError(t, err, "Update failed")

	assert.Equal(t, "updated@example.com", updatedUser.Email)

	// Verify the change in the database
	retrievedUser, err := repo.GetByID(ctx, createdUser.ID)
	require.NoError(t, err, "GetByID failed")

	assert.Equal(t, "updated@example.com", retrievedUser.Email)
}

func TestUpdateNonExistentUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

	ctx := context.Background()

	// Attempt to update a user that doesn't exist
	nonExistentUser := User{ID: 999, Username: "nonexistent", Email: "nonexistent@example.com"}
	_, err = repo.Update(ctx, nonExistentUser)
	require.Error(t, err, "Expected an error when updating a non-existent user")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

	ctx := context.Background()

	// Create a user to delete
	newUser := User{Username: "testuser", Email: "test@example.com"}
	createdUser, err := repo.Create(ctx, newUser)
	require.NoError(t, err, "Create failed")

	// Delete the user
	err = repo.Delete(ctx, createdUser.ID)
	require.NoError(t, err, "Delete failed")

	// Verify the user is deleted
	_, err = repo.GetByID(ctx, createdUser.ID)
	require.Error(t, err, "Expected an error when getting a deleted user")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDeleteNonExistentUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

	ctx := context.Background()

	// Attempt to delete a user that doesn't exist
	err = repo.Delete(ctx, 999)
	require.Error(t, err, "Expected an error when deleting a non-existent user")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
