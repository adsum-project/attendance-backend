package verification

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	verificationrepo "github.com/adsum-project/attendance-backend/internal/repositories/verification"
	timetablerepo "github.com/adsum-project/attendance-backend/internal/repositories/timetable"
	"github.com/adsum-project/attendance-backend/internal/services/timetable"
	testutil "github.com/adsum-project/attendance-backend/pkg/utils/testing"
)

func newMockedVerificationService(t *testing.T) (*VerificationService, sqlmock.Sqlmock, func()) {
	t.Helper()
	os.Setenv("EMBEDDINGS_API_URL", "http://localhost:9999")

	sqlxDB, mock, cleanup := testutil.NewMockedSQLxDB(t)
	verificationRepo := verificationrepo.NewVerificationRepository(sqlxDB)
	timetableRepo := timetablerepo.NewTimetableRepository(sqlxDB)
	timetableSvc, err := timetable.NewTimetableService(timetableRepo, nil)
	require.NoError(t, err)

	svc, err := NewVerificationService(verificationRepo, timetableSvc, nil)
	require.NoError(t, err)
	return svc, mock, cleanup
}

func TestNewVerificationService(t *testing.T) {
	t.Run("returns error when EMBEDDINGS_API_URL is empty", func(t *testing.T) {
		os.Unsetenv("EMBEDDINGS_API_URL")
		defer os.Setenv("EMBEDDINGS_API_URL", "http://test")

		sqlxDB, _, cleanup := testutil.NewMockedSQLxDB(t)
		defer cleanup()
		verificationRepo := verificationrepo.NewVerificationRepository(sqlxDB)
		timetableRepo := timetablerepo.NewTimetableRepository(sqlxDB)
		timetableSvc, err := timetable.NewTimetableService(timetableRepo, nil)
		require.NoError(t, err)

		svc, err := NewVerificationService(verificationRepo, timetableSvc, nil)
		require.Error(t, err)
		assert.Nil(t, svc)
		assert.Contains(t, err.Error(), "EMBEDDINGS_API_URL")
	})

	t.Run("returns error when timetable service is nil", func(t *testing.T) {
		os.Setenv("EMBEDDINGS_API_URL", "http://test")
		defer os.Unsetenv("EMBEDDINGS_API_URL")

		sqlxDB, _, cleanup := testutil.NewMockedSQLxDB(t)
		defer cleanup()
		verificationRepo := verificationrepo.NewVerificationRepository(sqlxDB)

		svc, err := NewVerificationService(verificationRepo, nil, nil)
		require.Error(t, err)
		assert.Nil(t, svc)
		assert.Contains(t, err.Error(), "timetable")
	})
}

func TestIsQRTokenValid(t *testing.T) {
	t.Run("returns false when token invalid", func(t *testing.T) {
		svc, mock, cleanup := newMockedVerificationService(t)
		defer cleanup()

		mock.ExpectQuery("SELECT 1 FROM qr_tokens").
			WithArgs("bad-token").
			WillReturnError(sql.ErrNoRows)

		valid, err := svc.IsQRTokenValid(context.Background(), "bad-token")
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("returns true when token valid", func(t *testing.T) {
		svc, mock, cleanup := newMockedVerificationService(t)
		defer cleanup()

		mock.ExpectQuery("SELECT 1 FROM qr_tokens").
			WithArgs("valid-token").
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		valid, err := svc.IsQRTokenValid(context.Background(), "valid-token")
		require.NoError(t, err)
		assert.True(t, valid)
	})
}

func TestUpdateRecordStatus(t *testing.T) {
	t.Run("returns error for invalid status", func(t *testing.T) {
		svc, _, cleanup := newMockedVerificationService(t)
		defer cleanup()

		err := svc.UpdateRecordStatus(context.Background(), "r1", "invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "absent")
	})
}
