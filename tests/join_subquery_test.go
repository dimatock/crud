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

type Post struct {
	ID     int    `db:"id,pk"`
	UserID int    `db:"user_id"`
	Title  string `db:"title"`
}

func setupTestDBWithPosts(t *testing.T) *sql.DB {
	db := setupTestDB(t)

	schema := `
	CREATE TABLE posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		title TEXT NOT NULL
	);
	`
	_, err := db.Exec(schema)
	require.NoError(t, err, "Failed to create posts table")

	return db
}

func TestListWithJoin(t *testing.T) {
	db := setupTestDBWithPosts(t)
	defer db.Close()

	userRepo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err)

	postRepo, err := crud.NewRepository[Post](db, "posts", crud.SQLiteDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	// Insert some users and posts
	user1, err := userRepo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	require.NoError(t, err)
	_, err = postRepo.Create(ctx, Post{UserID: user1.ID, Title: "Post 1"})
	require.NoError(t, err)

	// Test listing users with a join to posts
	users, err := userRepo.List(ctx, crud.WithJoin[User]("INNER JOIN posts ON posts.user_id = users.id"), crud.WithFilter[User]("posts.title", "Post 1"))
	require.NoError(t, err)

	require.Len(t, users, 1)
	assert.Equal(t, "user1", users[0].Username)
}

func TestListWithSubquery(t *testing.T) {
	db := setupTestDBWithPosts(t)
	defer db.Close()

	userRepo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err)

	postRepo, err := crud.NewRepository[Post](db, "posts", crud.SQLiteDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	// Insert some users and posts
	user1, err := userRepo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	require.NoError(t, err)
	_, err = userRepo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	require.NoError(t, err)
	_, err = postRepo.Create(ctx, Post{UserID: user1.ID, Title: "Post 1"})
	require.NoError(t, err)

	// Test listing users with a subquery
	users, err := userRepo.List(ctx, crud.WithSubquery[User]("id", "IN", "SELECT user_id FROM posts WHERE title = ?", "Post 1"))
	require.NoError(t, err)

	require.Len(t, users, 1)
	assert.Equal(t, "user1", users[0].Username)
}
