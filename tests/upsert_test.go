package tests

import (
	"context"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func TestCreateOrUpdate_SQLite(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// 1. Create scenario
	userToCreate := User{ID: 1, Username: "upsert-user", Email: "initial@example.com"}
	createdUser, err := repo.CreateOrUpdate(ctx, userToCreate)
	if err != nil {
		t.Fatalf("CreateOrUpdate (create) failed: %v", err)
	}

	if createdUser.Username != userToCreate.Username {
		t.Errorf("Expected username %s, got %s", userToCreate.Username, createdUser.Username)
	}

	// Verify count is 1
	users, _ := repo.List(ctx, crud.WithFilter("id", 1))
	if len(users) != 1 {
		t.Fatalf("Expected 1 user in DB after create, got %d", len(users))
	}

	// 2. Update scenario
	userToUpdate := User{ID: 1, Username: "upsert-user-updated", Email: "updated@example.com"}
	updatedUser, err := repo.CreateOrUpdate(ctx, userToUpdate)
	if err != nil {
		t.Fatalf("CreateOrUpdate (update) failed: %v", err)
	}

	if updatedUser.Username != userToUpdate.Username {
		t.Errorf("Expected updated username %s, got %s", userToUpdate.Username, updatedUser.Username)
	}

	// Verify count is still 1
	users, _ = repo.List(ctx, crud.WithFilter("id", 1))
	if len(users) != 1 {
		t.Fatalf("Expected 1 user in DB after update, got %d", len(users))
	}

	// Verify the data was updated
	finalUser, _ := repo.GetByID(ctx, 1)
	if finalUser.Username != userToUpdate.Username {
		t.Errorf("Username was not updated in DB. Expected %s, got %s", userToUpdate.Username, finalUser.Username)
	}
}

func TestCreateOrUpdate_MySQL(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// 1. Create scenario
	userToCreate := User{ID: 1, Username: "upsert-user", Email: "initial@example.com"}
	createdUser, err := repo.CreateOrUpdate(ctx, userToCreate)
	if err != nil {
		t.Fatalf("CreateOrUpdate (create) for MySQL failed: %v", err)
	}

	if createdUser.Username != userToCreate.Username {
		t.Errorf("Expected username %s, got %s", userToCreate.Username, createdUser.Username)
	}

	// 2. Update scenario
	userToUpdate := User{ID: 1, Username: "upsert-user-updated", Email: "updated@example.com"}
	updatedUser, err := repo.CreateOrUpdate(ctx, userToUpdate)
	if err != nil {
		t.Fatalf("CreateOrUpdate (update) for MySQL failed: %v", err)
	}

	if updatedUser.Username != userToUpdate.Username {
		t.Errorf("Expected updated username %s, got %s", userToUpdate.Username, updatedUser.Username)
	}

	// Verify count is still 1
	users, _ := repo.List(ctx, crud.WithFilter("id", 1))
	if len(users) != 1 {
		t.Fatalf("Expected 1 user in DB after update, got %d", len(users))
	}
}
