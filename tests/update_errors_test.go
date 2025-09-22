package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
)

type UserWithNilPK struct {
	ID   *int   `db:"id,pk"`
	Name string `db:"name"`
}

func TestUpdate_NilPK(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	schema := `CREATE TABLE users_with_nil_pk (id INTEGER PRIMARY KEY, name TEXT);`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo, err := crud.NewRepository[UserWithNilPK](db, "users_with_nil_pk", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Create a user with a nil PK
	user := UserWithNilPK{Name: "test"}

	// Attempt to update the user
	_, err = repo.Update(ctx, user)
	if err == nil {
		t.Fatal("Expected an error when updating a user with a nil primary key, but got nil")
	}

	expectedError := "primary key value not found in item to update"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}
}
