package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	authrepo "github.com/adsum-project/attendance-backend/internal/repositories/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testutil "github.com/adsum-project/attendance-backend/pkg/utils/testing"
)

func newMockedAuthService(t *testing.T) (*AuthService, sqlmock.Sqlmock, func()) {
	t.Helper()
	sqlxDB, mock, cleanup := testutil.NewMockedSQLxDB(t)
	sessionRepo := authrepo.NewSessionRepository(sqlxDB, 24*time.Hour)
	svc := NewAuthServiceForTesting(sessionRepo)
	return svc, mock, cleanup
}

func TestCreateSession(t *testing.T) {
	t.Run("returns error when session repo returns error", func(t *testing.T) {
		svc := NewAuthServiceForTesting(nil)
		token, err := svc.CreateSession(context.Background(), "user-1", nil)
		require.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, errSessionRepoMissing)
	})

	t.Run("creates session and returns token", func(t *testing.T) {
		svc, mock, cleanup := newMockedAuthService(t)
		defer cleanup()

		mock.ExpectExec("INSERT INTO sessions").
			WithArgs(sqlmock.AnyArg(), "user-1", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		token, err := svc.CreateSession(context.Background(), "user-1", map[string]interface{}{"roles": []string{"admin"}})
		require.NoError(t, err)
		require.NotEmpty(t, token)
		assert.Len(t, token, 64)
	})
}

func TestGetSession(t *testing.T) {
	t.Run("returns error when session repo is nil", func(t *testing.T) {
		svc := NewAuthServiceForTesting(nil)
		session, err := svc.GetSession(context.Background(), "token")
		require.Error(t, err)
		assert.Nil(t, session)
		assert.ErrorIs(t, err, errSessionRepoMissing)
	})

	t.Run("returns ErrSessionNotFound when session missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedAuthService(t)
		defer cleanup()

		mock.ExpectQuery("SELECT .+ FROM sessions").
			WithArgs("bad-token").
			WillReturnError(sql.ErrNoRows)

		session, err := svc.GetSession(context.Background(), "bad-token")
		require.Error(t, err)
		assert.Nil(t, session)
		assert.ErrorIs(t, err, authrepo.ErrSessionNotFound)
	})
}

func TestDeleteSession(t *testing.T) {
	t.Run("returns error when session repo is nil", func(t *testing.T) {
		svc := NewAuthServiceForTesting(nil)
		err := svc.DeleteSession(context.Background(), "token")
		require.Error(t, err)
		assert.ErrorIs(t, err, errSessionRepoMissing)
	})

	t.Run("deletes session successfully", func(t *testing.T) {
		svc, mock, cleanup := newMockedAuthService(t)
		defer cleanup()

		mock.ExpectExec("DELETE FROM sessions").
			WithArgs("token-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := svc.DeleteSession(context.Background(), "token-1")
		require.NoError(t, err)
	})
}

func TestGetUserIDFromClaims(t *testing.T) {
	t.Run("returns empty when claims nil", func(t *testing.T) {
		svc := NewAuthServiceForTesting(nil)
		assert.Empty(t, svc.GetUserIDFromClaims(nil))
	})

	t.Run("returns oid when present", func(t *testing.T) {
		svc := NewAuthServiceForTesting(nil)
		claims := map[string]interface{}{"oid": "entra-user-123", "name": "Test"}
		assert.Equal(t, "entra-user-123", svc.GetUserIDFromClaims(claims))
	})

	t.Run("returns empty when oid missing", func(t *testing.T) {
		svc := NewAuthServiceForTesting(nil)
		claims := map[string]interface{}{"name": "Test"}
		assert.Empty(t, svc.GetUserIDFromClaims(claims))
	})

	t.Run("returns empty when oid is empty string", func(t *testing.T) {
		svc := NewAuthServiceForTesting(nil)
		claims := map[string]interface{}{"oid": ""}
		assert.Empty(t, svc.GetUserIDFromClaims(claims))
	})
}
