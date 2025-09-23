package tests

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/dimatock/crud"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMySQLTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Skip("MYSQL_DSN environment variable not set, skipping MySQL tests")
	}

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err, "Failed to open MySQL database")

	// Clean up the table before the test
	_, err = db.Exec(`DROP TABLE IF EXISTS users;`)
	require.NoError(t, err, "Failed to drop table")

	schema := `
	CREATE TABLE users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(255) NOT NULL UNIQUE,
		email VARCHAR(255) NOT NULL UNIQUE
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err, "Failed to create table")

	return db
}

func TestMySQLConnection(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	require.NoError(t, err, "Failed to create repository with MySQL dialect")
	require.NotNil(t, repo, "Expected a repository, but got nil")
}

func TestMySQLCreateUser(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	newUser := User{Username: "testuser", Email: "test@example.com"}
	createdUser, err := repo.Create(ctx, newUser)
	require.NoError(t, err)

	assert.NotEqual(t, 0, createdUser.ID, "Expected created user to have an ID")
}

func TestMySQLListUsers(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	require.NoError(t, err)

	users, err := repo.List(ctx)
	require.NoError(t, err)

	assert.Len(t, users, 1)
}

func TestMySQLUpdateUser(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	createdUser, err := repo.Create(ctx, User{Username: "testuser", Email: "test@example.com"})
	require.NoError(t, err)

	createdUser.Email = "updated@example.com"
	_, err = repo.Update(ctx, createdUser)
	require.NoError(t, err)

	retrievedUser, err := repo.GetByID(ctx, createdUser.ID)
	require.NoError(t, err)

	assert.Equal(t, "updated@example.com", retrievedUser.Email)
}

func TestMySQLDeleteUser(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	createdUser, err := repo.Create(ctx, User{Username: "testuser", Email: "test@example.com"})
	require.NoError(t, err)

	err = repo.Delete(ctx, createdUser.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, createdUser.ID)
	require.Error(t, err, "Expected an error when getting a deleted user")
}

func TestMySQLListWithOptions(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	require.NoError(t, err)

	ctx := context.Background()

	_, err = repo.Create(ctx, User{Username: "user1", Email: "user1@example.com"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, User{Username: "user2", Email: "user2@example.com"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, User{Username: "user3", Email: "user3@example.com"})
	require.NoError(t, err)

	users, err := repo.List(ctx, crud.WithSort[User]("username", crud.SortAsc), crud.WithLimit[User](1), crud.WithOffset[User](1))
	require.NoError(t, err)

	require.Len(t, users, 1)
	assert.Equal(t, "user2", users[0].Username)
}

func TestMySQLGetByIDWithLock(t *testing.T) {
	db := setupMySQLTestDB(t)
	defer db.Close()

	repo, err := crud.NewRepository[User](db, "users", crud.MySQLDialect{})
	require.NoError(t, err)

	ctx := context.Background()
	createdUser, err := repo.Create(ctx, User{Username: "lock-user", Email: "lock@example.com"})
	require.NoError(t, err)

	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()

	txRepo := repo.WithTx(tx)

	retrievedUser, err := txRepo.GetByID(ctx, createdUser.ID, crud.WithLock[User]("FOR UPDATE"))
	require.NoError(t, err)

	assert.Equal(t, createdUser.Username, retrievedUser.Username)

	require.NoError(t, tx.Commit())
}
