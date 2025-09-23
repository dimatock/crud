package tests

import (
	"context"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListWithIn(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	// Insert some users
	_, _ = repo.Create(ctx, User{Username: "user1", Email: "u1@example.com"})
	_, _ = repo.Create(ctx, User{Username: "user2", Email: "u2@example.com"})
	_, _ = repo.Create(ctx, User{Username: "user3", Email: "u3@example.com"})

	// Test WithIn
	users, err := repo.List(ctx, crud.WithIn[User]("username", "user1", "user3"))
	require.NoError(t, err)
	require.Len(t, users, 2)

	// Check that we got the correct users
	foundUser1 := false
	foundUser3 := false
	for _, u := range users {
		if u.Username == "user1" {
			foundUser1 = true
		}
		if u.Username == "user3" {
			foundUser3 = true
		}
	}

	assert.True(t, foundUser1, "Did not retrieve user1")
	assert.True(t, foundUser3, "Did not retrieve user3")
}

func TestListWithIn_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	// Test WithIn with no values
	_, err = repo.List(ctx, crud.WithIn[User]("username"))
	require.Error(t, err)
	assert.Equal(t, "WithIn option requires at least one value for column 'username'", err.Error())
}

func TestListWithLike(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	// Insert some users
	_, _ = repo.Create(ctx, User{Username: "test-user-1", Email: "u1@example.com"})
	_, _ = repo.Create(ctx, User{Username: "test-user-2", Email: "u2@example.com"})
	_, _ = repo.Create(ctx, User{Username: "another-user", Email: "u3@example.com"})

	// Test WithLike
	users, err := repo.List(ctx, crud.WithLike[User]("username", "test-user-%"))
	require.NoError(t, err)
	require.Len(t, users, 2)
}

func TestListWithOperator(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	// Insert some users
	u1, _ := repo.Create(ctx, User{Username: "user1", Email: "u1@example.com"})
	_, _ = repo.Create(ctx, User{Username: "user2", Email: "u2@example.com"})
	_, _ = repo.Create(ctx, User{Username: "user3", Email: "u3@example.com"})

	// Test WithOperator (e.g., >)
	users, err := repo.List(ctx, crud.WithOperator[User]("id", ">", u1.ID))
	require.NoError(t, err)
	require.Len(t, users, 2)

	// Test WithOperator (e.g., !=)
	users, err = repo.List(ctx, crud.WithOperator[User]("username", "!=", "user2"))
	require.NoError(t, err)
	require.Len(t, users, 2)
}