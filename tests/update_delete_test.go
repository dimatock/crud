package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
)

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Create a user to update
	newUser := User{Username: "testuser", Email: "test@example.com"}
	createdUser, err := repo.Create(ctx, newUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update the user's email
	createdUser.Email = "updated@example.com"
	updatedUser, err := repo.Update(ctx, createdUser)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updatedUser.Email != "updated@example.com" {
		t.Errorf("Expected email to be updated to 'updated@example.com', got '%s'", updatedUser.Email)
	}

	// Verify the change in the database
	retrievedUser, err := repo.GetByID(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrievedUser.Email != "updated@example.com" {
		t.Errorf("Expected email in DB to be 'updated@example.com', got '%s'", retrievedUser.Email)
	}
}

func TestUpdateNonExistentUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Attempt to update a user that doesn't exist
	nonExistentUser := User{ID: 999, Username: "nonexistent", Email: "nonexistent@example.com"}
	_, err = repo.Update(ctx, nonExistentUser)
	if err == nil {
		t.Fatal("Expected an error when updating a non-existent user, but got nil")
	}

	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Create a user to delete
	newUser := User{Username: "testuser", Email: "test@example.com"}
	createdUser, err := repo.Create(ctx, newUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete the user
	err = repo.Delete(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify the user is deleted
	_, err = repo.GetByID(ctx, createdUser.ID)
	if err == nil {
		t.Fatal("Expected an error when getting a deleted user, but got nil")
	}

	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestDeleteNonExistentUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Attempt to delete a user that doesn't exist
	err = repo.Delete(ctx, 999)
	if err == nil {
		t.Fatal("Expected an error when deleting a non-existent user, but got nil")
	}

	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}
