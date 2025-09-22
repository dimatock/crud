package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
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
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create posts table: %v", err)
	}

	return db
}

func TestListWithJoin(t *testing.T) {
	db := setupTestDBWithPosts(t)
	defer db.Close()

	userRepo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create user repository: %v", err)
	}

	postRepo, err := crud.NewRepository[Post](db, "posts", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create post repository: %v", err)
	}

	ctx := context.Background()

	// Insert some users and posts
	user1, err := userRepo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	_, err = postRepo.Create(ctx, Post{UserID: user1.ID, Title: "Post 1"})
	if err != nil {
		t.Fatalf("Failed to create post1: %v", err)
	}

	// Test listing users with a join to posts
	users, err := userRepo.List(ctx, crud.WithJoin("INNER JOIN posts ON posts.user_id = users.id"), crud.WithFilter("posts.title", "Post 1"))
	if err != nil {
		t.Fatalf("List with join failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
	if users[0].Username != "user1" {
		t.Errorf("Expected user1, got %s", users[0].Username)
	}
}

func TestListWithSubquery(t *testing.T) {
	db := setupTestDBWithPosts(t)
	defer db.Close()

	userRepo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create user repository: %v", err)
	}

	postRepo, err := crud.NewRepository[Post](db, "posts", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create post repository: %v", err)
	}

	ctx := context.Background()

	// Insert some users and posts
	user1, err := userRepo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	_, err = userRepo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}
	_, err = postRepo.Create(ctx, Post{UserID: user1.ID, Title: "Post 1"})
	if err != nil {
		t.Fatalf("Failed to create post1: %v", err)
	}

	// Test listing users with a subquery
	users, err := userRepo.List(ctx, crud.WithSubquery("id", "IN", "SELECT user_id FROM posts WHERE title = ?", "Post 1"))
	if err != nil {
		t.Fatalf("List with subquery failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
	if users[0].Username != "user1" {
		t.Errorf("Expected user1, got %s", users[0].Username)
	}
}
