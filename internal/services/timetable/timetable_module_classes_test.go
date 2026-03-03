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

func TestGetClasses(t *testing.T) {
	t.Run("returns not found when module missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM modules WHERE module_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		result, err := svc.GetClasses(context.Background(), "missing", 1, 10)

		require.Error(t, err)
		assert.Nil(t, result)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})
}

func TestGetClass(t *testing.T) {
	t.Run("returns not found when module missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM modules WHERE module_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		class, err := svc.GetClass(context.Background(), "missing", "cl1")

		require.Error(t, err)
		assert.Nil(t, class)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})

	t.Run("returns class when found", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM modules WHERE module_id = ").
			WithArgs("m1").
			WillReturnRows(sqlmock.NewRows([]string{"module_id", "module_code", "module_name", "owner_id", "start_date", "end_date", "created_at", "updated_at"}).
				AddRow("m1", "ABC123", "Math", "owner", "2024-01-01", "2024-12-31", "2024-01-01", "2024-01-01"))
		mock.ExpectQuery("SELECT .+ FROM classes").
			WithArgs("m1", "cl1").
			WillReturnRows(sqlmock.NewRows([]string{"class_id", "module_id", "class_name", "room", "day_of_week", "starts_at", "ends_at", "recurrence", "created_at", "updated_at"}).
				AddRow("cl1", "m1", "Lecture", "ROOM1", 1, "09:00:00", "10:00:00", "weekly", "2024-01-01", "2024-01-01"))

		got, err := svc.GetClass(context.Background(), "m1", "cl1")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "cl1", got.ClassID)
		assert.Equal(t, "Lecture", got.ClassName)
		assert.Equal(t, "ROOM1", got.Room)
	})
}

func TestCreateClass(t *testing.T) {
	t.Run("returns validation error for invalid params", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM modules WHERE module_id = ").
			WithArgs("m1").
			WillReturnRows(sqlmock.NewRows([]string{"module_id", "module_code", "module_name", "owner_id", "start_date", "end_date", "created_at", "updated_at"}).
				AddRow("m1", "ABC123", "Math", "owner", "2024-01-01", "2024-12-31", "2024-01-01", "2024-01-01"))

		_, err := svc.CreateClass(testutil.CtxWithOwner("owner"), "m1", "", "ROOM1", 1, "09:00", "10:00", "weekly")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "className")
	})

	t.Run("returns not found when module missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM modules WHERE module_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		_, err := svc.CreateClass(testutil.CtxWithOwner("owner"), "missing", "Lecture", "ROOM1", 1, "09:00", "10:00", "weekly")

		require.Error(t, err)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})
}

func TestDeleteClass(t *testing.T) {
	t.Run("returns not found when module missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM modules WHERE module_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		err := svc.DeleteClass(testutil.CtxWithOwner("owner"), "missing", "cl1")

		require.Error(t, err)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})
}
