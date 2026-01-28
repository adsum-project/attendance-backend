package authrepo

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const sessionTable = "sessions"

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

type Session struct {
	ID        string
	UserID    string
	Claims    map[string]interface{}
	ExpiresAt time.Time
}

type SessionRepository struct {
	db         *sqlx.DB
	sessionTTL time.Duration
}

// Expected schema:
// CREATE TABLE auth_sessions (
//
//	id TEXT PRIMARY KEY,
//	user_id TEXT NOT NULL,
//	claims_json TEXT NOT NULL,
//	expires_at TIMESTAMP NOT NULL
//
// );
func NewSessionRepository(db *sqlx.DB, sessionTTL time.Duration) *SessionRepository {
	return &SessionRepository{
		db:         db,
		sessionTTL: sessionTTL,
	}
}

func (r *SessionRepository) CreateSession(ctx context.Context, userID string, claims map[string]interface{}) (string, error) {
	if r == nil || r.db == nil {
		return "", fmt.Errorf("db is required")
	}
	if userID == "" {
		return "", fmt.Errorf("userID is required")
	}

	token, err := generateSessionToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to serialize claims: %w", err)
	}

	expiresAt := time.Now().Add(r.sessionTTL)
	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO "+sessionTable+" (id, user_id, claims_json, expires_at) VALUES (@p1, @p2, @p3, @p4)",
		token,
		userID,
		string(claimsJSON),
		expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return token, nil
}

func (r *SessionRepository) GetSession(ctx context.Context, token string) (*Session, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("db is required")
	}
	if token == "" {
		return nil, ErrSessionNotFound
	}

	var userID string
	var claimsJSON string
	var expiresAt time.Time
	err := r.db.QueryRowContext(
		ctx,
		"SELECT user_id, claims_json, expires_at FROM "+sessionTable+" WHERE id = @p1",
		token,
	).Scan(&userID, &claimsJSON, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	if time.Now().After(expiresAt) {
		_ = r.DeleteSession(ctx, token)
		return nil, ErrSessionExpired
	}

	var claims map[string]interface{}
	if err := json.Unmarshal([]byte(claimsJSON), &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &Session{
		ID:        token,
		UserID:    userID,
		Claims:    claims,
		ExpiresAt: expiresAt,
	}, nil
}

func (r *SessionRepository) DeleteSession(ctx context.Context, token string) error {
	if r == nil || r.db == nil || token == "" {
		return nil
	}
	_, err := r.db.ExecContext(ctx, "DELETE FROM "+sessionTable+" WHERE id = @p1", token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func generateSessionToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}
