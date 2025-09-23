package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UserWithNilPK struct {
	ID   *int   `db:"id,pk"`
	Name string `db:"name"`
}

func TestUpdate_NilPK(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "Failed to open SQLite database")
	defer db.Close()

	schema := `CREATE TABLE users_with_nil_pk (id INTEGER PRIMARY KEY, name TEXT);`
	_, err = db.Exec(schema)
	require.NoError(t, err, "Failed to create table")

	repo, err := crud.NewRepository[UserWithNilPK](db, "users_with_nil_pk", crud.SQLiteDialect{})
	require.NoError(t, err, "Failed to create repository")

	ctx := context.Background()

	// Create a user with a nil PK
	user := UserWithNilPK{Name: "test"}

	// Attempt to update the user
	_, err = repo.Update(ctx, user)
	require.Error(t, err, "Expected an error when updating a user with a nil primary key")
	assert.Equal(t, "primary key value not found in item to update", err.Error())
}
