package tests

import (
	"context"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOrUpdate_SQLite(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

	ctx := context.Background()

	// 1. Create scenario
	userToCreate := User{ID: 1, Username: "upsert-user", Email: "initial@example.com"}
	createdUser, err := repo.CreateOrUpdate(ctx, userToCreate)
	require.NoError(t, err, "CreateOrUpdate (create) failed")

	assert.Equal(t, userToCreate.Username, createdUser.Username)

	// Verify count is 1
	users, err := repo.List(ctx, repo.Where("id", 1))
	require.NoError(t, err)
	require.Len(t, users, 1)

	// 2. Update scenario
	userToUpdate := User{ID: 1, Username: "upsert-user-updated", Email: "updated@example.com"}
	updatedUser, err := repo.CreateOrUpdate(ctx, userToUpdate)
	require.NoError(t, err, "CreateOrUpdate (update) failed")

	assert.Equal(t, userToUpdate.Username, updatedUser.Username)

	// Verify count is still 1
	users, err = repo.List(ctx, repo.Where("id", 1))
	require.NoError(t, err)
	require.Len(t, users, 1)

	// Verify the data was updated
	finalUser, err := repo.GetByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, userToUpdate.Username, finalUser.Username)
}

func TestCreateOrUpdate_MySQL(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	require.NoError(t, err, "Failed to create repository")

	ctx := context.Background()

	// 1. Create scenario
	userToCreate := User{ID: 1, Username: "upsert-user", Email: "initial@example.com"}
	createdUser, err := repo.CreateOrUpdate(ctx, userToCreate)
	require.NoError(t, err, "CreateOrUpdate (create) for MySQL failed")

	assert.Equal(t, userToCreate.Username, createdUser.Username)

	// 2. Update scenario
	userToUpdate := User{ID: 1, Username: "upsert-user-updated", Email: "updated@example.com"}
	updatedUser, err := repo.CreateOrUpdate(ctx, userToUpdate)
	require.NoError(t, err, "CreateOrUpdate (update) for MySQL failed")

	assert.Equal(t, userToUpdate.Username, updatedUser.Username)

	// Verify count is still 1
	users, err := repo.List(ctx, repo.Where("id", 1))
	require.NoError(t, err)
	assert.Len(t, users, 1)
}
