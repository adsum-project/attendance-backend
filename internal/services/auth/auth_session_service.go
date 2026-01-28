package auth

import (
	"context"

	authrepo "github.com/adsum-project/attendance-backend/internal/repo/auth"
)

type Session = authrepo.Session

var (
	ErrSessionNotFound = authrepo.ErrSessionNotFound
	ErrSessionExpired  = authrepo.ErrSessionExpired
)

func (a *Auth) CreateSession(ctx context.Context, userID string, claims map[string]interface{}) (string, error) {
	if a.sessionRepo == nil {
		return "", errSessionRepoMissing
	}
	return a.sessionRepo.CreateSession(ctx, userID, claims)
}

func (a *Auth) GetSession(ctx context.Context, token string) (*Session, error) {
	if a.sessionRepo == nil {
		return nil, errSessionRepoMissing
	}
	return a.sessionRepo.GetSession(ctx, token)
}

func (a *Auth) DeleteSession(ctx context.Context, token string) error {
	if a.sessionRepo == nil {
		return errSessionRepoMissing
	}
	return a.sessionRepo.DeleteSession(ctx, token)
}
