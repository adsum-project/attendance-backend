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

func TestGetCourseStudents(t *testing.T) {
	t.Run("returns not found when course missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM courses WHERE course_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		result, err := svc.GetCourseStudents(context.Background(), "missing", 1, 10)

		require.Error(t, err)
		assert.Nil(t, result)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})

	t.Run("returns empty list when no students enrolled", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM courses WHERE course_id = ").
			WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"course_id", "course_code", "course_name", "campus", "owner_id", "created_at", "updated_at"}).
				AddRow("c1", "1234", "Math", "Main", "owner", "2024-01-01", "2024-01-01"))
		mock.ExpectQuery("SELECT .+ FROM course_students").
			WithArgs("c1", 0, 10).
			WillReturnRows(sqlmock.NewRows([]string{"course_id", "user_id", "year_of_study"}))
		mock.ExpectQuery("SELECT COUNT").
			WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		result, err := svc.GetCourseStudents(context.Background(), "c1", 1, 10)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Empty(t, result.Data)
	})
}

func TestUnassignStudentFromCourse(t *testing.T) {
	t.Run("returns not found when course missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM courses WHERE course_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		err := svc.UnassignStudentFromCourse(testutil.CtxWithOwner("owner"), "missing", "u1")

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

		err := svc.UnassignStudentFromCourse(testutil.CtxWithUser("random-user"), "c1", "u1")

		require.Error(t, err)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusForbidden, httpErr.StatusCode())
	})
}
