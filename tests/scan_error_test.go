package tests

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
)

type UserWithScanError struct {
	ID   int    `db:"id,pk"`
	Name string `db:"name"`
}

func TestScanError(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	schema := `CREATE TABLE users_with_scan_error (id INTEGER PRIMARY KEY, name TEXT);`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert a row with a non-integer value in the name column
	_, err = db.Exec("INSERT INTO users_with_scan_error (id, name) VALUES (1, 'test')")
	if err != nil {
		t.Fatalf("Failed to insert row: %v", err)
	}

	// Create a repository for a struct with an int field for the name column
	type UserWithIntName struct {
		ID   int `db:"id,pk"`
		Name int `db:"name"`
	}

	repo, err := crud.NewRepository[UserWithIntName](db, "users_with_scan_error", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Attempt to get the user, which should cause a scan error
	_, err = repo.GetByID(ctx, 1)
	if err == nil {
		t.Fatal("Expected an error when scanning a row with incompatible types, but got nil")
	}

	if !strings.Contains(err.Error(), "sql: Scan error") {
		t.Errorf("Expected error message to contain 'sql: Scan error', but got '%s'", err.Error())
	}
}
