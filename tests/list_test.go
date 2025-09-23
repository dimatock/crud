package tests

import (
	"context"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	// Insert some users for testing
	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	require.NoError(t, err)

	// Test listing all users
	users, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestListWithWhereMethod(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	// Insert some users for testing
	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	require.NoError(t, err)

	// Test listing with a filter
	users, err := repo.List(ctx, repo.Where("username", "user1"))
	require.NoError(t, err)

	require.Len(t, users, 1)
	assert.Equal(t, "user1", users[0].Username)
}

func TestListWithRawWhere(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	// Insert some users for testing
	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, User{Username: "user3", Email: "user3@example.com"})
	require.NoError(t, err)

	// Test listing with a WHERE clause (OR condition)
	users, err := repo.List(ctx, repo.Where("username = ? OR username = ?", "user1", "user3"))
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestListWithSort(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	// Insert some users for testing
	_, err = repo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	require.NoError(t, err)

	// Test listing with sort
	users, err := repo.List(ctx, repo.OrderBy("username", crud.SortAsc))
	require.NoError(t, err)

	require.Len(t, users, 2)
	assert.Equal(t, "user1", users[0].Username)
	assert.Equal(t, "user2", users[1].Username)
}

func TestListWithLimitAndOffset(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	// Insert some users for testing
	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, User{Username: "user3", Email: "user3@example.com"})
	require.NoError(t, err)

	// Test listing with limit and offset
	users, err := repo.List(ctx, repo.Limit(1), repo.Offset(1), repo.OrderBy("username", crud.SortAsc))
	require.NoError(t, err)

	require.Len(t, users, 1)
	assert.Equal(t, "user2", users[0].Username)
}
