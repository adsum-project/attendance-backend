package testutil

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// CtxWithUser returns a context with the given userID for authorization tests.
func CtxWithUser(userID string) context.Context {
	return context.WithValue(context.Background(), "userID", userID)
}

// CtxWithOwner returns a context with the given ownerID and admin role for owner/admin tests.
func CtxWithOwner(ownerID string) context.Context {
	ctx := context.WithValue(context.Background(), "userID", ownerID)
	return context.WithValue(ctx, "claims", map[string]any{"roles": []string{"admin"}})
}

// NewMockedDB creates a *sqlx.DB backed by sqlmock for testing.
// Returns the db, mock for setting expectations, and a cleanup function.
func NewMockedDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	cleanup := func() { db.Close() }
	return db, mock, cleanup
}

// NewMockedSQLxDB creates a *sqlx.DB backed by sqlmock for testing.
func NewMockedSQLxDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, cleanup := NewMockedDB(t)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock, cleanup
}
