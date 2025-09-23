package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dimatock/crud"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UUIDModel struct {
	ID   string `db:"id,pk"`
	Data string `db:"data"`
}

func TestUUIDPrimaryKey(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE uuid_models (id TEXT PRIMARY KEY, data TEXT)`);
	require.NoError(t, err)

	repo, err := crud.NewRepository[UUIDModel](db, "uuid_models", crud.SQLiteDialect{})
	require.NoError(t, err)

	// Test Create
	newID := uuid.New().String()
	created, err := repo.Create(context.Background(), UUIDModel{ID: newID, Data: "test data"})
	require.NoError(t, err)
	assert.Equal(t, newID, created.ID)
	assert.Equal(t, "test data", created.Data)

	// Test GetByID
	fetched, err := repo.GetByID(context.Background(), newID)
	require.NoError(t, err)
	assert.Equal(t, newID, fetched.ID)

	// Test List with filter
	items, err := repo.List(context.Background(), crud.WithFilter[UUIDModel]("id", newID))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, newID, items[0].ID)

	// Test Update
	created.Data = "updated data"
	updated, err := repo.Update(context.Background(), created)
	require.NoError(t, err)
	assert.Equal(t, "updated data", updated.Data)

	// Test Delete
	err = repo.Delete(context.Background(), newID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), newID)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
