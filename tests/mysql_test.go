package tests

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/go-sql-driver/mysql"
)

func setupMySQLTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Skip("MYSQL_DSN environment variable not set, skipping MySQL tests")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("Failed to open MySQL database: %v", err)
	}

	// Clean up the table before the test
	_, err = db.Exec(`DROP TABLE IF EXISTS users;`)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}

	schema := `
	CREATE TABLE users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(255) NOT NULL UNIQUE,
		email VARCHAR(255) NOT NULL UNIQUE
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func TestMySQLConnection(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository with MySQL dialect: %v", err)
	}

	if repo == nil {
		t.Fatal("Expected a repository, but got nil")
	}
}

func TestMySQLCreateUser(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	newUser := User{Username: "testuser", Email: "test@example.com"}
	createdUser, err := repo.Create(ctx, newUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if createdUser.ID == 0 {
		t.Errorf("Expected created user to have an ID, got 0")
	}
}

func TestMySQLListUsers(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	users, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
}

func TestMySQLUpdateUser(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	createdUser, err := repo.Create(ctx, User{Username: "testuser", Email: "test@example.com"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	createdUser.Email = "updated@example.com"
	_, err = repo.Update(ctx, createdUser)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrievedUser, err := repo.GetByID(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrievedUser.Email != "updated@example.com" {
		t.Errorf("Expected email to be updated, got %s", retrievedUser.Email)
	}
}

func TestMySQLDeleteUser(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	createdUser, err := repo.Create(ctx, User{Username: "testuser", Email: "test@example.com"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = repo.Delete(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, createdUser.ID)
	if err == nil {
		t.Fatal("Expected an error when getting a deleted user, but got nil")
	}
}

func TestMySQLListWithOptions(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	_, err = repo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}
	_, err = repo.Create(ctx, User{Username: "user3", Email: "user3@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user3: %v", err)
	}

	users, err := repo.List(ctx, crud.WithSort("username", crud.SortAsc), crud.WithLimit(1), crud.WithOffset(1))
	if err != nil {
		t.Fatalf("List with options failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	if users[0].Username != "user2" {
		t.Errorf("Expected user2, got %s", users[0].Username)
	}
}

func TestMySQLGetByIDWithLock(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()
	createdUser, err := repo.Create(ctx, User{Username: "lock-user", Email: "lock@example.com"})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	txRepo := repo.WithTx(tx)

	retrievedUser, err := txRepo.GetByID(ctx, createdUser.ID, crud.WithLock("FOR UPDATE"))
	if err != nil {
		t.Fatalf("GetByID with lock failed: %v", err)
	}

	if retrievedUser.Username != createdUser.Username {
		t.Errorf("Expected username %s, got %s", createdUser.Username, retrievedUser.Username)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}
}
