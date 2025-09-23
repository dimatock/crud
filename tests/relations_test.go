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

// --- Test Models ---

type RelUser struct {
	ID      int         `db:"id,pk"`
	Name    string      `db:"name"`
	Posts   []*RelPost  `db:"-"`
	Profile *RelProfile `db:"-"`
}

type RelPost struct {
	ID     int     `db:"id,pk"`
	UserID int     `db:"user_id"`
	Title  string  `db:"title"`
	User   *RelUser `db:"-"`
}

type RelProfile struct {
	ID     int    `db:"id,pk"`
	UserID int    `db:"user_id"`
	Bio    string `db:"bio"`
}

func setupRelationsDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);
		CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER, title TEXT);
		CREATE TABLE profiles (id INTEGER PRIMARY KEY, user_id INTEGER, bio TEXT);
	`)
	require.NoError(t, err)

	// Seed data
	_, err = db.Exec(`
		INSERT INTO users (id, name) VALUES (1, 'John Doe'), (2, 'Jane Doe');
		INSERT INTO posts (id, user_id, title) VALUES (101, 1, 'Post 1 by John'), (102, 1, 'Post 2 by John'), (103, 2, 'Post 1 by Jane');
		INSERT INTO profiles (id, user_id, bio) VALUES (1, 1, 'Johns Bio');
	`)
	require.NoError(t, err)

	return db
}

func TestEagerLoading(t *testing.T) {
	db := setupRelationsDB(t)
	defer db.Close()

	dialect := crud.SQLiteDialect{}
	userRepo, err := crud.NewRepository[RelUser](db, "users", dialect)
	require.NoError(t, err)
	postRepo, err := crud.NewRepository[RelPost](db, "posts", dialect)
	require.NoError(t, err)
	profileRepo, err := crud.NewRepository[RelProfile](db, "profiles", dialect)
	require.NoError(t, err)

	t.Run("WithOneToMany", func(t *testing.T) {
		mapper := crud.OneToManyMapper[RelUser, RelPost, int]{
			Fetcher: func(ctx context.Context, userIDs []int) ([]RelPost, error) {
				return postRepo.List(ctx, crud.WithIn[RelPost]("user_id", crud.IntsToAnys(userIDs)...))
			},
			GetPK:      func(u *RelUser) int { return u.ID },
			GetFK:      func(p *RelPost) int { return p.UserID },
			SetRelated: func(u *RelUser, p []*RelPost) { u.Posts = p },
		}

		users, err := userRepo.List(context.Background(), crud.With[RelUser](mapper))
		require.NoError(t, err)
		require.Len(t, users, 2)

		john := users[0]
		jane := users[1]

		assert.Equal(t, "John Doe", john.Name)
		require.NotNil(t, john.Posts)
		assert.Len(t, john.Posts, 2)
		assert.Equal(t, "Post 1 by John", john.Posts[0].Title)

		assert.Equal(t, "Jane Doe", jane.Name)
		require.NotNil(t, jane.Posts)
		assert.Len(t, jane.Posts, 1)
		assert.Equal(t, "Post 1 by Jane", jane.Posts[0].Title)
	})

	t.Run("WithManyToOne", func(t *testing.T) {
		mapper := crud.ManyToOneMapper[RelPost, RelUser, int]{
			Fetcher: func(ctx context.Context, userIDs []int) ([]RelUser, error) {
				return userRepo.List(ctx, crud.WithIn[RelUser]("id", crud.IntsToAnys(userIDs)...))
			},
			GetFK:      func(p *RelPost) int { return p.UserID },
			GetPK:      func(u *RelUser) int { return u.ID },
			SetRelated: func(p *RelPost, u *RelUser) { p.User = u },
		}

		posts, err := postRepo.List(context.Background(), crud.With[RelPost](mapper))
		require.NoError(t, err)
		require.Len(t, posts, 3)

		for _, post := range posts {
			require.NotNil(t, post.User)
			assert.Equal(t, post.UserID, post.User.ID)
		}
		assert.Equal(t, "John Doe", posts[0].User.Name)
	})

	t.Run("WithHasOne", func(t *testing.T) {
		mapper := crud.HasOneMapper[RelUser, RelProfile, int]{
			Fetcher: func(ctx context.Context, userIDs []int) ([]RelProfile, error) {
				return profileRepo.List(ctx, crud.WithIn[RelProfile]("user_id", crud.IntsToAnys(userIDs)...))
			},
			GetPK:      func(u *RelUser) int { return u.ID },
			GetFK:      func(p *RelProfile) int { return p.UserID },
			SetRelated: func(u *RelUser, p *RelProfile) { u.Profile = p },
		}

		users, err := userRepo.List(context.Background(), crud.With[RelUser](mapper))
		require.NoError(t, err)
		require.Len(t, users, 2)

		john := users[0]
		jane := users[1]

		assert.NotNil(t, john.Profile)
		assert.Equal(t, "Johns Bio", john.Profile.Bio)
		assert.Nil(t, jane.Profile)
	})
}
