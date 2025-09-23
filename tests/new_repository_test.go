package tests

import (
	"database/sql"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepository_MultiplePK(t *testing.T) {
	type UserWithMultiplePK struct {
		ID1 int `db:"id1,pk"`
		ID2 int `db:"id2,pk"`
	}

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "Failed to open SQLite database")
	defer db.Close()

	_, err = crud.NewRepository[UserWithMultiplePK](db, "users", crud.SQLiteDialect{})
	require.Error(t, err, "Expected an error when creating a repository for a struct with multiple primary keys")
	assert.Equal(t, "multiple primary key fields defined in UserWithMultiplePK", err.Error())
}

func TestNewRepository_NoDBTags(t *testing.T) {
	type UserWithNoDBTags struct {
		ID   int
		Name string
	}

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "Failed to open SQLite database")
	defer db.Close()

	_, err = crud.NewRepository[UserWithNoDBTags](db, "users", crud.SQLiteDialect{})
	require.Error(t, err, "Expected an error when creating a repository for a struct with no 'db' tags")
	assert.Equal(t, "no 'db' tags found in struct UserWithNoDBTags", err.Error())
}

func TestNewRepository_NoPK(t *testing.T) {
	type UserWithNoPK struct {
		ID int `db:"id"`
	}

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "Failed to open SQLite database")
	defer db.Close()

	_, err = crud.NewRepository[UserWithNoPK](db, "users", crud.SQLiteDialect{})
	require.Error(t, err, "Expected an error when creating a repository for a struct with no primary key")
	assert.Equal(t, "no primary key field defined with ',pk' tag in struct UserWithNoPK", err.Error())
}

func TestNewRepository_NonStruct(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "Failed to open SQLite database")
	defer db.Close()

	_, err = crud.NewRepository[int](db, "users", crud.SQLiteDialect{})
	require.Error(t, err, "Expected an error when creating a repository for a non-struct type")
	assert.Equal(t, "generic type T must be a struct, but got int", err.Error())
}
