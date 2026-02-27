package verificationrepo

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	attendanceRecordsTable = "attendance_records"
	qrTokensTable         = "qr_tokens"
)

var (
	ErrTokenInvalid    = errors.New("token invalid or expired")
	ErrAlreadySignedIn = errors.New("already signed in to this class")
)

type VerificationRepository struct {
	db *sqlx.DB
}

func NewVerificationRepository(db *sqlx.DB) *VerificationRepository {
	return &VerificationRepository{db: db}
}

func (r *VerificationRepository) HasSignedIn(ctx context.Context, userID, classID string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx,
		`SELECT 1 FROM `+attendanceRecordsTable+`
		WHERE user_id = @p1 AND class_id = CONVERT(uniqueidentifier, @p2)`,
		userID, classID,
	).Scan(&exists)
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
		`INSERT INTO `+attendanceRecordsTable+` (record_id, user_id, class_id, signed_in_at, method)
		VALUES (NEWID(), @p1, CONVERT(uniqueidentifier, @p2), SYSUTCDATETIME(), @p3)`,
		userID, classID, method,
	)
	if err != nil {
		return fmt.Errorf("failed to insert attendance record: %w", err)
	}
	return nil
}

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

func (r *VerificationRepository) ConsumeToken(ctx context.Context, token string) (classID string, err error) {
	err = r.db.QueryRowContext(
		ctx,
		`UPDATE `+qrTokensTable+` SET used_at = SYSUTCDATETIME()
		OUTPUT LOWER(CONVERT(VARCHAR(36), INSERTED.class_id))
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
