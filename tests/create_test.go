package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type User struct {
	ID       int    `db:"id,pk"`
	Username string `db:"username"`
	Email    string `db:"email"`
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "Failed to open SQLite database")

	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err, "Failed to create table")

	return db
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

	ctx := context.Background()

	// Test creating a user
	newUser := User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	createdUser, err := repo.Create(ctx, newUser)
	require.NoError(t, err, "Create failed")

	assert.NotEqual(t, 0, createdUser.ID, "Expected created user to have an ID")
	assert.Equal(t, newUser.Username, createdUser.Username)

	// Verify the user can be retrieved by ID
	retrievedUser, err := repo.GetByID(ctx, createdUser.ID)
	require.NoError(t, err, "GetByID failed")

	assert.Equal(t, createdUser.ID, retrievedUser.ID)
	assert.Equal(t, createdUser.Username, retrievedUser.Username)
}

func TestSqlInjection(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

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
	require.NoError(t, err, "Create with malicious input failed")

	// Verify the malicious user was created as a literal string
	assert.Equal(t, maliciousUsername, createdMaliciousUser.Username)
	assert.Equal(t, maliciousEmail, createdMaliciousUser.Email)

	// Crucial check: Try to insert another user to see if the table is still intact
	normalUser := User{
		Username: "normaluser",
		Email:    "normal@example.com",
	}
	_, err = repo.Create(ctx, normalUser)
	require.NoError(t, err, "Table integrity check failed: could not create normal user")

	// Verify that there are now two users in the database
	allUsers, err := repo.List(ctx)
	require.NoError(t, err, "Failed to list users")
	assert.Len(t, allUsers, 2)

	// Verify that the malicious user's ID is valid (i.e., it was inserted)
	assert.NotEqual(t, 0, createdMaliciousUser.ID, "Malicious user ID is 0, expected a valid ID")
}
