package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/dimatock/crud"
	_ "github.com/mattn/go-sqlite3"
)

// ComplexModel tests various data types.
type ComplexModel struct {
	ID          int            `db:"id,pk"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Age         sql.NullInt64  `db:"age"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   *time.Time     `db:"updated_at"`
}

// UUIDModel tests string primary keys.
type UUIDModel struct {
	ID   string `db:"id,pk"`
	Data string `db:"data"`
}

func setupComplexModelTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}

	schema := `
	CREATE TABLE complex_models (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		description TEXT,
		age INTEGER,
		created_at DATETIME,
		updated_at DATETIME
	);`
	_, err = db.Exec(schema)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func setupUUIDModelTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}

	schema := `
	CREATE TABLE uuid_models (
		id TEXT PRIMARY KEY,
		data TEXT
	);`
	_, err = db.Exec(schema)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func TestCRUDWithComplexTypes(t *testing.T) {
	db := setupComplexModelTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[ComplexModel](db, "complex_models", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second) // Truncate for DB compatibility

	// --- Test Create with values ---
	newItem := ComplexModel{
		Name:        "Test Item",
		Description: sql.NullString{String: "A description", Valid: true},
		Age:         sql.NullInt64{Int64: 30, Valid: true},
		CreatedAt:   now,
		UpdatedAt:   &now,
	}

	created, err := repo.Create(ctx, newItem)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("Create should have returned an ID")
	}

	// --- Test GetByID ---
	retrieved, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Name != newItem.Name {
		t.Errorf("Name mismatch: got %v, want %v", retrieved.Name, newItem.Name)
	}
	if retrieved.Description.String != newItem.Description.String {
		t.Errorf("Description mismatch: got %v, want %v", retrieved.Description.String, newItem.Description.String)
	}
	if retrieved.Age.Int64 != newItem.Age.Int64 {
		t.Errorf("Age mismatch: got %v, want %v", retrieved.Age.Int64, newItem.Age.Int64)
	}
	if !retrieved.CreatedAt.Equal(newItem.CreatedAt) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", retrieved.CreatedAt, newItem.CreatedAt)
	}
	if retrieved.UpdatedAt == nil || !retrieved.UpdatedAt.Equal(*newItem.UpdatedAt) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", retrieved.UpdatedAt, newItem.UpdatedAt)
	}

	// --- Test Update ---
	updatedAt := time.Now().UTC().Truncate(time.Second).Add(5 * time.Minute)
	retrieved.Name = "Updated Name"
	retrieved.Description = sql.NullString{} // Test NULL value
	retrieved.UpdatedAt = &updatedAt

	updated, err := repo.Update(ctx, retrieved)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Update did not update name correctly")
	}
	if updated.Description.Valid {
		t.Errorf("Update did not set description to NULL")
	}

	// --- Verify Update in DB ---
	verified, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID after update failed: %v", err)
	}
	if verified.Name != "Updated Name" {
		t.Errorf("Name was not updated in DB")
	}
	if verified.Description.Valid {
		t.Errorf("Description was not set to NULL in DB")
	}
	if !verified.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt was not updated in DB")
	}
}

func TestCRUDWithStringPK(t *testing.T) {
	db := setupUUIDModelTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[UUIDModel](db, "uuid_models", crud.SQLiteDialect{})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()
	id := "a-unique-identifier"

	// --- Test Create ---
	newItem := UUIDModel{ID: id, Data: "initial data"}
	created, err := repo.Create(ctx, newItem)
	if err != nil {
		t.Fatalf("Create with string PK failed: %v", err)
	}
	if created.ID != newItem.ID {
		t.Errorf("Create returned wrong ID: got %s, want %s", created.ID, newItem.ID)
	}

	// --- Test GetByID ---
	retrieved, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.ID != id || retrieved.Data != "initial data" {
		t.Errorf("GetByID retrieved incorrect data: got %+v", retrieved)
	}

	// --- Test Update ---
	retrieved.Data = "updated data"
	updated, err := repo.Update(ctx, retrieved)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Data != "updated data" {
		t.Errorf("Update did not return updated data")
	}

	// Verify update in DB
	verified, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID after update failed: %v", err)
	}
	if verified.Data != "updated data" {
		t.Errorf("Data was not updated in DB")
	}

	// --- Test Delete ---
	err = repo.Delete(ctx, id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, id)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows after delete, but got %v", err)
	}
}
