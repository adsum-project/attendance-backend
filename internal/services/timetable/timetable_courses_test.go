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

	timetablerepo "github.com/adsum-project/attendance-backend/internal/repositories/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/errs"
	testutil "github.com/adsum-project/attendance-backend/pkg/utils/testing"
)

func newMockedService(t *testing.T) (*TimetableService, sqlmock.Sqlmock, func()) {
	t.Helper()
	sqlxDB, mock, cleanup := testutil.NewMockedSQLxDB(t)
	repo := timetablerepo.NewTimetableRepository(sqlxDB)
	svc, err := NewTimetableService(repo, nil)
	require.NoError(t, err)
	return svc, mock, cleanup
}

func TestNewTimetableService(t *testing.T) {
	t.Run("returns error when repo is nil", func(t *testing.T) {
		svc, err := NewTimetableService(nil, nil)
		require.Error(t, err)
		assert.Nil(t, svc)
		assert.Contains(t, err.Error(), "required")
	})
}

func TestCreateCourse(t *testing.T) {
	t.Run("returns validation error for invalid course code", func(t *testing.T) {
		svc, _, cleanup := newMockedService(t)
		defer cleanup()

		_, err := svc.CreateCourse(testutil.CtxWithUser("u1"), "ABC", "Test Course", "Main")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "courseCode")
	})

	t.Run("returns conflict when course code exists", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT 1 FROM courses WHERE course_code = ").
			WithArgs("1234").
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		_, err := svc.CreateCourse(testutil.CtxWithUser("u1"), "1234", "Valid Course", "Main")

		require.Error(t, err)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusConflict, httpErr.StatusCode())
	})

	t.Run("creates course and returns id when valid", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT 1 FROM courses WHERE course_code = ").
			WithArgs("1234").
			WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery("INSERT INTO courses").
			WithArgs("1234", "Valid Course", "Main", "u1").
			WillReturnRows(sqlmock.NewRows([]string{"course_id"}).AddRow("course-id-1"))

		id, err := svc.CreateCourse(testutil.CtxWithUser("u1"), "1234", "Valid Course", "Main")

		require.NoError(t, err)
		assert.Equal(t, "course-id-1", id)
	})
}

func TestGetCourse(t *testing.T) {
	t.Run("returns not found when course missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM courses WHERE course_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		course, err := svc.GetCourse(context.Background(), "missing")

		require.Error(t, err)
		assert.Nil(t, course)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})

	t.Run("returns course when found", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM courses WHERE course_id = ").
			WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"course_id", "course_code", "course_name", "campus", "owner_id", "created_at", "updated_at"}).
				AddRow("c1", "1234", "Math", "Main", "owner-1", "2024-01-01", "2024-01-01"))

		got, err := svc.GetCourse(context.Background(), "c1")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "c1", got.CourseID)
		assert.Equal(t, "1234", got.CourseCode)
		assert.Equal(t, "Math", got.CourseName)
		assert.Equal(t, "Main", got.Campus)
	})
}

func TestGetOwnCourses(t *testing.T) {
	t.Run("returns nil when no user in context", func(t *testing.T) {
		svc, _, cleanup := newMockedService(t)
		defer cleanup()

		courses, err := svc.GetOwnCourses(context.Background())

		require.NoError(t, err)
		assert.Nil(t, courses)
	})

	t.Run("returns courses for user in context", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM course_students cs").
			WithArgs("user-1").
			WillReturnRows(sqlmock.NewRows([]string{"course_id", "course_code", "course_name", "campus", "owner_id", "created_at", "updated_at"}).
				AddRow("c1", "1234", "Math", "Main", "owner-1", "2024-01-01", "2024-01-01"))

		got, err := svc.GetOwnCourses(testutil.CtxWithUser("user-1"))

		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "c1", got[0].CourseID)
		assert.Equal(t, "1234", got[0].CourseCode)
	})
}

func TestUpdateCourse(t *testing.T) {
	t.Run("returns not found when course missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM courses WHERE course_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)
		name := "New Name"

		err := svc.UpdateCourse(testutil.CtxWithOwner("owner"), "missing", nil, &name, nil)

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
		name := "New Name"

		err := svc.UpdateCourse(testutil.CtxWithUser("random-user"), "c1", nil, &name, nil)

		require.Error(t, err)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusForbidden, httpErr.StatusCode())
	})
}

func TestDeleteCourse(t *testing.T) {
	t.Run("returns not found when course missing", func(t *testing.T) {
		svc, mock, cleanup := newMockedService(t)
		defer cleanup()
		mock.ExpectQuery("SELECT .+ FROM courses WHERE course_id = ").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		err := svc.DeleteCourse(testutil.CtxWithOwner("owner"), "missing")

		require.Error(t, err)
		var httpErr errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode())
	})
}
