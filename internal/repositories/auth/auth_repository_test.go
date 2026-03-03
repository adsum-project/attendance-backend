package authrepo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockedRepo(t *testing.T) (*SessionRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewSessionRepository(sqlxDB, 24*time.Hour)
	cleanup := func() { db.Close() }
	return repo, mock, cleanup
}

func TestCreateSession(t *testing.T) {
	t.Run("returns error when userID empty", func(t *testing.T) {
		repo, _, cleanup := newMockedRepo(t)
		defer cleanup()

		_, err := repo.CreateSession(context.Background(), "", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "userID")
	})

	t.Run("creates session and returns token", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		mock.ExpectExec("INSERT INTO sessions").
			WithArgs(sqlmock.AnyArg(), "user-1", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		token, err := repo.CreateSession(context.Background(), "user-1", map[string]interface{}{"roles": []string{"user"}})
		require.NoError(t, err)
		require.NotEmpty(t, token)
		assert.Len(t, token, 64)
	})
}

func TestGetSession(t *testing.T) {
	t.Run("returns ErrSessionNotFound when token empty", func(t *testing.T) {
		repo, _, cleanup := newMockedRepo(t)
		defer cleanup()

		session, err := repo.GetSession(context.Background(), "")
		require.Error(t, err)
		assert.Nil(t, session)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("returns ErrSessionNotFound when session missing", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		mock.ExpectQuery("SELECT .+ FROM sessions").
			WithArgs("bad-token").
			WillReturnError(sql.ErrNoRows)

		session, err := repo.GetSession(context.Background(), "bad-token")
		require.Error(t, err)
		assert.Nil(t, session)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})
}

func TestDeleteSession(t *testing.T) {
	t.Run("succeeds when token empty", func(t *testing.T) {
		repo, _, cleanup := newMockedRepo(t)
		defer cleanup()

		err := repo.DeleteSession(context.Background(), "")
		require.NoError(t, err)
	})

	t.Run("deletes session when found", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		mock.ExpectExec("DELETE FROM sessions").
			WithArgs("token-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteSession(context.Background(), "token-1")
		require.NoError(t, err)
	})
}
