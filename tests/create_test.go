package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type User struct {
	ID       int    `db:"id,pk"`
	Username string `db:"username"`
	Email    string `db:"email"`
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}

	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Test creating a user
	newUser := User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	createdUser, err := repo.Create(ctx, newUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if createdUser.ID == 0 {
		t.Errorf("Expected created user to have an ID, got 0")
	}
	if createdUser.Username != newUser.Username {
		t.Errorf("Expected username %s, got %s", newUser.Username, createdUser.Username)
	}

	// Verify the user can be retrieved by ID
	retrievedUser, err := repo.GetByID(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrievedUser.ID != createdUser.ID {
		t.Errorf("Retrieved user ID mismatch: expected %d, got %d", createdUser.ID, retrievedUser.ID)
	}
	if retrievedUser.Username != createdUser.Username {
		t.Errorf("Retrieved username mismatch: expected %s, got %s", createdUser.Username, retrievedUser.Username)
	}
}

func TestSqlInjection(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Malicious input
	maliciousUsername := "admin' OR 1=1 --"
	maliciousEmail := "test@example.com'; DROP TABLE users; --"

	maliciousUser := User{
		Username: maliciousUsername,
		Email:    maliciousEmail,
	}

	// Attempt to create a user with malicious input
	createdMaliciousUser, err := repo.Create(ctx, maliciousUser)
	if err != nil {
		t.Fatalf("Create with malicious input failed: %v", err)
	}

	// Verify the malicious user was created as a literal string
	if createdMaliciousUser.Username != maliciousUsername {
		t.Errorf(
			"Malicious username not stored literally: expected %s, got %s", maliciousUsername,
			createdMaliciousUser.Username,
		)
	}
	if createdMaliciousUser.Email != maliciousEmail {
		t.Errorf(
			"Malicious email not stored literally: expected %s, got %s", maliciousEmail, createdMaliciousUser.Email,
		)
	}

	// Crucial check: Try to insert another user to see if the table is still intact
	normalUser := User{
		Username: "normaluser",
		Email:    "normal@example.com",
	}
	_, err = repo.Create(ctx, normalUser)
	if err != nil {
		// If this fails, it likely means the table was dropped or corrupted
		t.Fatalf("Table integrity check failed: could not create normal user. Error: %v", err)
	}

	// Verify that there are now two users in the database
	allUsers, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if len(allUsers) != 2 {
		t.Errorf("Expected 2 users in the database, got %d", len(allUsers))
	}

	// Verify that the malicious user's ID is valid (i.e., it was inserted)
	if createdMaliciousUser.ID == 0 {
		t.Errorf("Malicious user ID is 0, expected a valid ID")
	}
}
