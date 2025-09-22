package tests

import (
	"database/sql"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
)

func TestNewRepository_MultiplePK(t *testing.T) {
	type UserWithMultiplePK struct {
		ID1 int `db:"id1,pk"`
		ID2 int `db:"id2,pk"`
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	_, err = crud.NewRepository[UserWithMultiplePK](db, "users", crud.SQLiteDialect{})
	if err == nil {
		t.Fatal("Expected an error when creating a repository for a struct with multiple primary keys, but got nil")
	}

	expectedError := "multiple primary key fields defined in UserWithMultiplePK"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}
}

func TestNewRepository_NoDBTags(t *testing.T) {
	type UserWithNoDBTags struct {
		ID   int
		Name string
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	_, err = crud.NewRepository[UserWithNoDBTags](db, "users", crud.SQLiteDialect{})
	if err == nil {
		t.Fatal("Expected an error when creating a repository for a struct with no 'db' tags, but got nil")
	}

	expectedError := "no 'db' tags found in struct UserWithNoDBTags"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}
}

func TestNewRepository_NoPK(t *testing.T) {
	type UserWithNoPK struct {
		ID int `db:"id"`
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	_, err = crud.NewRepository[UserWithNoPK](db, "users", crud.SQLiteDialect{})
	if err == nil {
		t.Fatal("Expected an error when creating a repository for a struct with no primary key, but got nil")
	}

	expectedError := "no primary key field defined with ',pk' tag in struct UserWithNoPK"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}
}

func TestNewRepository_NonStruct(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	_, err = crud.NewRepository[int](db, "users", crud.SQLiteDialect{})
	if err == nil {
		t.Fatal("Expected an error when creating a repository for a non-struct type, but got nil")
	}

	expectedError := "generic type T must be a struct, but got int"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}
}
