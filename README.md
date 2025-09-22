# CRUD - Generic CRUD Repository for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/dimatock/crud.svg)](https://pkg.go.dev/github.com/dimatock/crud)

This package provides a generic, based on generics, repository for performing
CRUD (Create, Read, Update, Delete) operations with your Go structs. It
simplifies database interactions by abstracting away repetitive SQL code.

## Features

- **Generic:** Works with any Go struct.
- **Flexible Primary Keys:** Supports both auto-incrementing integers and user-provided keys (e.g., UUIDs).
- **Upsert Operation:** Provides a `CreateOrUpdate` method for `INSERT ... ON CONFLICT` semantics.
- **ACID Transactions:** All operations can be performed within a database transaction for data consistency.
- **Pessimistic Locking:** Supports `FOR UPDATE` and other row-locking clauses via a `WithLock` option.
- **Simple Mapping:** Uses struct field tags (`db:"..."`) to map to table columns.
- **SQL Dialect Support:** Easily extensible for different databases (built-in support for MySQL, SQLite, and PostgreSQL).
- **Flexible Queries:** Allows building complex queries using options (filtering, sorting, pagination, joins).
- **Extensible:** Allows embedding the base repository into your own structs to add custom logic.

## Installation

```bash
go get github.com/dimatock/crud
```

## Quick Start

### 1. Define Your Model

Create a Go struct and add `db` tags to the fields you want to map to database
columns. Be sure to mark the primary key field with `,pk`.

```go
package main

// Model with an auto-incrementing integer PK
type User struct {
    ID       int    `db:"id,pk"`
    Username string `db:"username"`
    Email    string `db:"email"`
}

// Model with a user-provided string PK
type Product struct {
    ID   string `db:"id,pk"`
    Name string `db:"name"`
}
```

### 2. Initialize the Repository

```go
// Set up the database connection (e.g., in-memory SQLite for this example)
db, err := sql.Open("sqlite3", ":memory:")
// ... handle error

// Create repositories for your models
userRepo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
productRepo, err := crud.NewRepository[Product](db, "products", crud.SQLiteDialect{})
// ... handle errors
```

### 3. Use CRUD Operations

#### Create

The behavior of `Create` adapts based on the primary key type.

```go
ctx := context.Background()

// For auto-incrementing keys, the new ID is fetched and returned.
newUser := User{Username: "johndoe", Email: "john.doe@example.com"}
createdUser, err := userRepo.Create(ctx, newUser)
fmt.Printf("User created with new ID: %d\n", createdUser.ID)

// For user-provided keys (e.g., UUIDs), the original object is returned.
newProduct := Product{ID: "prod-123-xyz", Name: "My Awesome Product"}
createdProduct, err := productRepo.Create(ctx, newProduct)
fmt.Printf("Product created with specified ID: %s\n", createdProduct.ID)
```

#### CreateOrUpdate (Upsert)

This method inserts a record or updates it if a record with the same primary key already exists.

```go
// This user does not exist, so it will be created.
user1 := User{ID: 10, Username: "new-user", Email: "new@example.com"}
finalUser1, err := userRepo.CreateOrUpdate(ctx, user1)

// Now, let's update the user's email. Because ID 10 already exists,
// the existing record will be updated.
user1.Email = "updated@example.com"
finalUser1, err = userRepo.CreateOrUpdate(ctx, user1)

fmt.Printf("Upserted user has email: %s\n", finalUser1.Email)
```

#### GetByID, Update, Delete, List

These methods work as expected for all primary key types.

```go
// GetByID
user, err := userRepo.GetByID(ctx, 1)

// Update
user.Email = "new.email@example.com"
updatedUser, err := userRepo.Update(ctx, user)

// Delete
err = userRepo.Delete(ctx, 1)

// List with basic options
users, err := userRepo.List(ctx,
    crud.WithFilter("username", "johndoe"),
    crud.WithSort("id", crud.SortDesc),
)
```

#### Advanced Filtering

Combine multiple options to build complex queries.

```go
// Find all users named "user1" or "user3"
users, err := userRepo.List(ctx,
    crud.WithIn("username", "user1", "user3"),
)

// Find all products whose names start with "Awesome"
products, err := productRepo.List(ctx,
    crud.WithLike("name", "Awesome%"),
)

// Find all users with an ID greater than 5
users, err = userRepo.List(ctx,
    crud.WithOperator("id", ">", 5),
)
```

## Using Transactions

You can run multiple operations in a single atomic transaction. The repository is
immutable, so calling `WithTx` returns a *new* repository instance that is
scoped to the transaction.

```go
// 1. Create a base repository
userRepo, err := crud.NewRepository[User](db, "users", crud.SQLiteDialect{})
// ...

// 2. Begin a transaction
ctx := context.Background()
tx, err := db.BeginTx(ctx, nil)
// ...

// 3. Get a transactional repository and perform operations
txRepo := userRepo.WithTx(tx)
_, err = txRepo.Create(ctx, User{Username: "user1", Email: "u1@example.com"})
// ... handle error and rollback

// 4. Commit the transaction
err = tx.Commit()
// ...
```

## Pessimistic Locking

To prevent race conditions during read-modify-write cycles, you can apply a pessimistic lock (e.g., `FOR UPDATE`) to your `GetByID` or `List` calls. This feature **must be used within a transaction**.

The `WithLock` option accepts a raw string, allowing you to use dialect-specific clauses like `FOR UPDATE SKIP LOCKED`.

```go
// Assume txRepo is a repository created with WithTx(tx)

// 1. Select a user and lock the row
// The query will be SELECT ... FROM users WHERE id = ? FOR UPDATE
user, err := txRepo.GetByID(ctx, 1, crud.WithLock("FOR UPDATE"))
if err != nil {
    tx.Rollback()
    // ... handle error
}

// 2. Now you can safely modify the user object
user.Username = "locked-and-updated"

_, err = txRepo.Update(ctx, user)
if err != nil {
    tx.Rollback()
    // ... handle error
}

// 3. Commit the transaction to release the lock
err = tx.Commit()
// ...
```

## Extending the Repository

You can easily add your own methods by embedding `crud.RepositoryInterface` into
your own repository struct.

### 1. Define Your Interface and Struct

```go
// UserRepositoryInterface extends the base interface
type UserRepositoryInterface interface {
    crud.RepositoryInterface[User]
    GetByEmail(ctx context.Context, email string) (User, error)
}

// userRepository implements the new interface
type userRepository struct {
    crud.RepositoryInterface[User] // Embedding
}

// NewUserRepository creates a new instance of your custom repository
func NewUserRepository(
    db *sql.DB,
    dialect crud.Dialect,
) (UserRepositoryInterface, error) {
    baseRepo, err := crud.NewRepository[User](db, "users", dialect)
    if err != nil {
        return nil, err
    }
    return &userRepository{RepositoryInterface: baseRepo}, nil
}
```

### 2. Implement Custom Methods

```go
// GetByEmail finds a user by their email
func (r *userRepository) GetByEmail(
    ctx context.Context,
    email string,
) (User, error) {
    // Use the List method from the embedded repository with a filter
    users, err := r.List(ctx,
        crud.WithFilter("email", email),
        crud.WithLimit(1),
    )
    if err != nil {
        return User{}, err
    }
    if len(users) == 0 {
        return User{}, sql.ErrNoRows // User not found
    }
    return users[0], nil
}
```

### 3. Use the Custom Repository

```go
// Create the custom repository
customRepo, err := NewUserRepository(db, crud.SQLiteDialect{})
// ...

// Call both standard and custom methods
user, err := customRepo.GetByEmail(ctx, "john.doe@example.com")
// ...
```

## Testing

The project includes both unit tests and integration tests.

Unit tests use an in-memory SQLite database and can be run without any
additional configuration:

```bash
go test ./...
```

The integration tests run against a real MySQL database. To run them, you must
provide a Data Source Name (DSN) via the `MYSQL_DSN` environment variable.
These tests will be skipped if the variable is not set.

**Example:**

```bash
export MYSQL_DSN="user:pass@tcp(127.0.0.1:3306)/db?parseTime=true"
go test ./...
```
