package tests

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanError(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "Failed to open SQLite database")
	defer db.Close()

	schema := `CREATE TABLE users_with_scan_error (id INTEGER PRIMARY KEY, name TEXT);`
	_, err = db.Exec(schema)
	require.NoError(t, err, "Failed to create table")

	// Insert a row with a non-integer value in the name column
	_, err = db.Exec("INSERT INTO users_with_scan_error (id, name) VALUES (1, 'test')")
	require.NoError(t, err, "Failed to insert row")

	// Create a repository for a struct with an int field for the name column
	type UserWithIntName struct {
		ID   int `db:"id,pk"`
		Name int `db:"name"`
	}

	repo, err := crud.NewRepository[UserWithIntName](db, "users_with_scan_error", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

	ctx := context.Background()

	// Attempt to get the user, which should cause a scan error
	_, err = repo.GetByID(ctx, 1)
	require.Error(t, err, "Expected an error when scanning a row with incompatible types")

	assert.True(t, strings.Contains(err.Error(), "sql: Scan error"), "Expected error message to contain 'sql: Scan error'")
}
