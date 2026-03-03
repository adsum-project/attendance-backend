package verificationrepo

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	verificationmodels "github.com/adsum-project/attendance-backend/internal/models/verification"
	"github.com/adsum-project/attendance-backend/pkg/utils/pagination"
	"github.com/adsum-project/attendance-backend/pkg/utils/query"
	"github.com/jmoiron/sqlx"
)

const (
	attendanceRecordsTable = "attendance_records"
	qrTokensTable         = "qr_tokens"
	statusPresent         = "present"
	statusAbsent          = "absent"
	statusExcused         = "excused"
	methodNone            = "none"
)

var (
	ErrTokenInvalid    = errors.New("token invalid or expired")
	ErrAlreadySignedIn = errors.New("already signed in to this class")
)

type VerificationRepository struct {
	db *sqlx.DB
}

// NewVerificationRepository creates the verification repo (QR tokens, attendance records).
func NewVerificationRepository(db *sqlx.DB) *VerificationRepository {
	return &VerificationRepository{db: db}
}

func (r *VerificationRepository) HasSignedIn(ctx context.Context, userID, classID string) (bool, error) {
	return r.HasSignedInForOccurrence(ctx, userID, classID, "")
}

// HasSignedInForOccurrence returns true if the user has signed in for the class on the given date.
func (r *VerificationRepository) HasSignedInForOccurrence(ctx context.Context, userID, classID, occurrenceDate string) (bool, error) {
	q := `SELECT 1 FROM ` + attendanceRecordsTable + `
		WHERE user_id = @p1 AND ` + query.GuidWhere("class_id", "@p2") + ` AND status = '` + statusPresent + `'`
	args := []any{userID, classID}
	if occurrenceDate != "" {
		q += ` AND CAST(signed_in_at AS DATE) = CAST(@p3 AS DATE)`
		args = append(args, occurrenceDate)
	}
	var exists int
	err := r.db.QueryRowContext(ctx, q, args...).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check attendance: %w", err)
	}
	return true, nil
}

func (r *VerificationRepository) InsertRecord(ctx context.Context, userID, classID, method string) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO `+attendanceRecordsTable+` (record_id, user_id, class_id, signed_in_at, method, status)
		VALUES (NEWID(), @p1, CONVERT(uniqueidentifier, @p2), SYSUTCDATETIME(), @p3, '`+statusPresent+`')`,
		userID, classID, method,
	)
	if err != nil {
		return fmt.Errorf("failed to insert attendance record: %w", err)
	}
	return nil
}

// CreateSignInToken stores a short-lived token for QR sign-in and returns it.
func (r *VerificationRepository) CreateSignInToken(ctx context.Context, classID string, expiresAt time.Time) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(b)
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO `+qrTokensTable+` (token, class_id, expires_at)
		VALUES (@p1, CONVERT(uniqueidentifier, @p2), @p3)`,
		token, classID, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create sign-in token: %w", err)
	}
	return token, nil
}

func (r *VerificationRepository) IsTokenValid(ctx context.Context, token string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(
		ctx,
		`SELECT 1 FROM `+qrTokensTable+`
		WHERE token = @p1 AND used_at IS NULL AND expires_at > SYSUTCDATETIME()`,
		token,
	).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check token: %w", err)
	}
	return true, nil
}

// ConsumeToken marks the token as used and returns the classID. Returns ErrTokenInvalid if expired or already used.
func (r *VerificationRepository) ConsumeToken(ctx context.Context, token string) (classID string, err error) {
	err = r.db.QueryRowContext(
		ctx,
		`UPDATE `+qrTokensTable+` SET used_at = SYSUTCDATETIME()
		OUTPUT `+query.Guid("INSERTED.class_id")+`
		WHERE token = @p1 AND used_at IS NULL AND expires_at > SYSUTCDATETIME()`,
		token,
	).Scan(&classID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrTokenInvalid
		}
		return "", fmt.Errorf("failed to consume token: %w", err)
	}
	return classID, nil
}

// DeleteExpiredQRTokens removes tokens that have expired.
func (r *VerificationRepository) DeleteExpiredQRTokens(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM `+qrTokensTable+` WHERE expires_at < SYSUTCDATETIME()`)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired QR tokens: %w", err)
	}
	return res.RowsAffected()
}

func (r *VerificationRepository) GetOwnRecords(ctx context.Context, userID string, page, perPage int) ([]verificationmodels.AttendanceRecord, error) {
	page, perPage = pagination.Normalize(page, perPage)
	baseQuery := `SELECT ` + query.Guid("ar.record_id") + ` as record_id,
		'' as user_id,
		` + query.Guid("ar.class_id") + ` as class_id,
		cl.class_name, m.module_code, m.module_name,
		` + query.Room("cl.room") + ` as room,
		` + query.DateTimeISO("ar.signed_in_at") + ` as signed_in_at,
		ISNULL(ar.method, '` + methodNone + `') as method,
		ISNULL(ar.status, 'present') as status
		FROM ` + attendanceRecordsTable + ` ar
		INNER JOIN classes cl ON ar.class_id = cl.class_id
		INNER JOIN modules m ON cl.module_id = m.module_id
		WHERE ar.user_id = @p1
		ORDER BY ar.signed_in_at DESC
		OFFSET @p2 ROWS FETCH NEXT @p3 ROWS ONLY`
	args := []any{userID, (page - 1) * perPage, perPage}
	var rows []verificationmodels.AttendanceRecord
	err := r.db.SelectContext(ctx, &rows, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance records: %w", err)
	}
	return rows, nil
}

func (r *VerificationRepository) GetOwnRecordsCount(ctx context.Context, userID string) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM `+attendanceRecordsTable+` WHERE user_id = @p1`,
		userID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to count attendance records: %w", err)
	}
	return total, nil
}

// GetRecordsByClass returns attendance records for the given class, joined with class and module info.
func (r *VerificationRepository) GetRecordsByClass(ctx context.Context, classID string, page, perPage int) ([]verificationmodels.AttendanceRecord, error) {
	page, perPage = pagination.Normalize(page, perPage)
	baseQuery := `SELECT ` + query.Guid("ar.record_id") + ` as record_id,
		` + query.Guid("ar.user_id") + ` as user_id,
		` + query.Guid("ar.class_id") + ` as class_id,
		cl.class_name, m.module_code, m.module_name,
		` + query.Room("cl.room") + ` as room,
		` + query.DateTimeISO("ar.signed_in_at") + ` as signed_in_at,
		ISNULL(ar.method, '` + methodNone + `') as method,
		ISNULL(ar.status, 'present') as status
		FROM ` + attendanceRecordsTable + ` ar
		INNER JOIN classes cl ON ar.class_id = cl.class_id
		INNER JOIN modules m ON cl.module_id = m.module_id
		WHERE ` + query.GuidWhere("ar.class_id", "@p1") + `
		ORDER BY ar.signed_in_at DESC
		OFFSET @p2 ROWS FETCH NEXT @p3 ROWS ONLY`
	args := []any{classID, (page - 1) * perPage, perPage}
	var rows []verificationmodels.AttendanceRecord
	err := r.db.SelectContext(ctx, &rows, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance records by class: %w", err)
	}
	return rows, nil
}

func (r *VerificationRepository) GetRecordsByClassCount(ctx context.Context, classID string) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM `+attendanceRecordsTable+` WHERE `+query.GuidWhere("class_id", "@p1"),
		classID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to count attendance records by class: %w", err)
	}
	return total, nil
}

// HasRecordForClassOccurrence returns true if the user has any attendance record for the given class on the occurrence date.
func (r *VerificationRepository) HasRecordForClassOccurrence(ctx context.Context, userID, classID, occurrenceDate string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx,
		`SELECT 1 FROM `+attendanceRecordsTable+`
		WHERE user_id = @p1 AND `+query.GuidWhere("class_id", "@p2")+`
		AND CAST(signed_in_at AS DATE) = CAST(@p3 AS DATE)`,
		userID, classID, occurrenceDate,
	).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check record: %w", err)
	}
	return true, nil
}

// InsertAbsentRecord inserts an attendance record with status=absent for the given user, class, and occurrence end time.
func (r *VerificationRepository) InsertAbsentRecord(ctx context.Context, userID, classID, occurrenceEndDatetime string) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO `+attendanceRecordsTable+` (record_id, user_id, class_id, signed_in_at, method, status)
		VALUES (NEWID(), @p1, CONVERT(uniqueidentifier, @p2), CAST(@p3 AS DATETIME2), '`+methodNone+`', '`+statusAbsent+`')`,
		userID, classID, occurrenceEndDatetime,
	)
	if err != nil {
		return fmt.Errorf("failed to insert absent record: %w", err)
	}
	return nil
}

// UpdateRecordStatus updates the status of an attendance record. Only allowed for absent/excused.
func (r *VerificationRepository) UpdateRecordStatus(ctx context.Context, recordID, status string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE `+attendanceRecordsTable+` SET status = @p1
		WHERE `+query.GuidWhere("record_id", "@p2")+` AND status IN ('absent', 'excused')`,
		status, recordID,
	)
	if err != nil {
		return fmt.Errorf("failed to update record status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
