package timetable

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adsum-project/attendance-backend/pkg/utils/errs"
	testutil "github.com/adsum-project/attendance-backend/pkg/utils/testing"
)

func TestCreateModule(t *testing.T) {
	t.Run("returns validation error for invalid module code", func(t *testing.T) {
		svc, _, cleanup := newMockedService(t)
		defer cleanup()

		_, err := svc.CreateModule(testutil.CtxWithUser("u1"), "AB", "Valid Name", "2024-01-01", "2024-12-31")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "moduleCode")
	})

	t.Run("returns conflict when module code exists", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT 1 FROM modules WHERE module_code = ").
			WithArgs("ABC123").
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		_, err := svc.CreateModule(testutil.CtxWithUser("u1"), "ABC123", "Valid Name", "2024-01-01", "2024-12-31")

		require.Error(t, err)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusConflict, httpErr.StatusCode())
	})

	t.Run("creates module and returns id when valid", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT 1 FROM modules WHERE module_code = ").
			WithArgs("ABC123").
			WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery("INSERT INTO modules").
			WithArgs("ABC123", "Valid Name", "u1", "2024-01-01", "2024-12-31").
			WillReturnRows(sqlmock.NewRows([]string{"module_id"}).AddRow("module-id-1"))

		id, err := svc.CreateModule(testutil.CtxWithUser("u1"), "ABC123", "Valid Name", "2024-01-01", "2024-12-31")

		require.NoError(t, err)
		assert.Equal(t, "module-id-1", id)
	})
}

func TestGetModule(t *testing.T) {
	t.Run("returns not found when module missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM modules WHERE module_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		module, err := svc.GetModule(context.Background(), "missing")

		require.Error(t, err)
		assert.Nil(t, module)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})

	t.Run("returns module when found", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM modules WHERE module_id = ").
			WithArgs("m1").
			WillReturnRows(sqlmock.NewRows([]string{"module_id", "module_code", "module_name", "owner_id", "start_date", "end_date", "created_at", "updated_at"}).
				AddRow("m1", "ABC123", "Math", "owner-1", "2024-01-01", "2024-12-31", "2024-01-01", "2024-01-01"))

		got, err := svc.GetModule(context.Background(), "m1")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "m1", got.ModuleID)
		assert.Equal(t, "ABC123", got.ModuleCode)
	})
}

func TestDeleteModule(t *testing.T) {
	t.Run("returns not found when module missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM modules WHERE module_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		err := svc.DeleteModule(testutil.CtxWithOwner("owner"), "missing")

		require.Error(t, err)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})
}
