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

func TestGetModuleCourses(t *testing.T) {
	t.Run("returns not found when module missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM modules WHERE module_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		result, err := svc.GetModuleCourses(context.Background(), "missing", 1, 10)

		require.Error(t, err)
		assert.Nil(t, result)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})
}

func TestGetCourseModules(t *testing.T) {
	t.Run("returns not found when course missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM courses WHERE course_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		result, err := svc.GetCourseModules(testutil.CtxWithUser("u1"), "missing", 1, 10)

		require.Error(t, err)
		assert.Nil(t, result)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})
}

func TestAssignModuleToCourse(t *testing.T) {
	t.Run("returns not found when course missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM courses WHERE course_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		err := svc.AssignModuleToCourse(testutil.CtxWithOwner("owner"), "missing", "m1", 1)

		require.Error(t, err)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})

	t.Run("returns forbidden when not owner or admin", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM courses WHERE course_id = ").
			WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"course_id", "course_code", "course_name", "campus", "owner_id", "created_at", "updated_at"}).
				AddRow("c1", "1234", "Math", "Main", "other-owner", "2024-01-01", "2024-01-01"))

		err := svc.AssignModuleToCourse(testutil.CtxWithUser("random-user"), "c1", "m1", 1)

		require.Error(t, err)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusForbidden, httpErr.StatusCode())
	})
}

func TestUnassignModuleFromCourse(t *testing.T) {
	t.Run("returns not found when course missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM courses WHERE course_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		err := svc.UnassignModuleFromCourse(testutil.CtxWithOwner("owner"), "missing", "m1")

		require.Error(t, err)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})
}
